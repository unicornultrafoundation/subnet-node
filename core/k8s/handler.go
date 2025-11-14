package k8s

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/builder"
	etypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/expiry"
	apclient "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/provider/client"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/util"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/apitypes"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	kubeclienterrors "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/errors"
	cip "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/ip"
	cfromctx "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/fromctx"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *service) RequestDeployment(ctx context.Context, deploymentID dtypes.DeploymentID, sdlManifest sdl.SDL) error {
	// Check if the deployment already has a active lease
	status, err := s.expiryService.CheckDeploymentExpiry(ctx, fmt.Sprintf("%d", deploymentID.DSeq))
	if err != nil {
		return err
	}
	if status.Status != etypes.DeploymentExpiryStatusActive {
		message := fmt.Sprintf("Deployment does not have a active lease with ID %d. Lease status: %s", deploymentID.DSeq, status.Status)
		if status.Status == etypes.DeploymentExpiryStatusExpired {
			message = fmt.Sprintf("Lease %d is expired. You need to add more funds to this lease to keep it active or it will be deleted in %d seconds", deploymentID.DSeq, status.TimeLeft)
		}
		return errors.New(message)
	}
	return s.manifestService.Submit(ctx, deploymentID, sdlManifest)
}

func (s *service) GetAllLeaseStatus(ctx context.Context) ([]apitypes.DeploymentStatus, error) {
	ctx = context.WithValue(ctx, builder.SettingsKey, s.config.ClusterSettings[builder.SettingsKey])

	deploymentStatuses := make([]apitypes.DeploymentStatus, 0)

	for _, manager := range s.managers {
		leaseID := manager.deployment.LeaseID()
		lease, err := s.client.LeaseStatus(ctx, leaseID)
		if err != nil {
			return nil, err
		}
		manifestGroups := make([]apitypes.ManifestGroup, 0)

		for group, services := range lease {
			manifestGroups = append(manifestGroups, apitypes.ManifestGroup{
				Group:    group,
				Services: services,
			})
		}

		expiryStatus, err := s.expiryService.CheckDeploymentExpiry(ctx, fmt.Sprintf("%d", leaseID.DSeq))
		if err != nil {
			return nil, fmt.Errorf("failed to check deployment expiry: %w", err)
		}

		deploymentStatuses = append(deploymentStatuses, apitypes.DeploymentStatus{
			LeaseID:        leaseID,
			ManifestGroups: manifestGroups,
			Status:         expiryStatus.Status,
			TimeLeft:       expiryStatus.TimeLeft,
		})
	}
	return deploymentStatuses, nil
}

func (s *service) GetLeaseStatus(ctx context.Context, leaseID mtypes.LeaseID) (apclient.LeaseStatus, error) {
	ctx = context.WithValue(ctx, builder.SettingsKey, s.config.ClusterSettings[builder.SettingsKey])

	result := apclient.LeaseStatus{}

	found, manifestGroup, err := s.client.GetManifestGroup(ctx, leaseID)
	if err != nil {
		return result, err
	}

	if !found {
		return result, nil
	}

	var ipLeaseStatus []cip.LeaseIPStatus

	if clIP := cfromctx.ClientIPFromContext(ctx); clIP != nil {
		hasLeasedIPs := false

	ipManifestGroupSearchLoop:
		for _, service := range manifestGroup.Services {
			for _, expose := range service.Expose {
				if len(expose.IP) != 0 {
					hasLeasedIPs = true
					break ipManifestGroupSearchLoop
				}
			}
		}

		if hasLeasedIPs {
			s.log.WithField("lease-id", leaseID).Debug("querying for IP address status")
			ipLeaseStatus, err = clIP.GetIPAddressStatus(ctx, leaseID.OrderID())
			if err != nil {
				return result, err
			}
			result.IPs = make(map[string][]apclient.LeasedIPStatus)

			for _, ipLease := range ipLeaseStatus {
				entries := result.IPs[ipLease.ServiceName]
				if entries == nil {
					entries = make([]apclient.LeasedIPStatus, 0)
				}

				entries = append(entries, apclient.LeasedIPStatus{
					Port:         ipLease.Port,
					ExternalPort: ipLease.ExternalPort,
					Protocol:     ipLease.Protocol,
					IP:           ipLease.IP,
				})

				result.IPs[ipLease.ServiceName] = entries
			}
		}
	}

	hasForwardedPorts := false
portManifestGroupSearchLoop:
	for _, service := range manifestGroup.Services {
		for _, expose := range service.Expose {
			if expose.Global {
				hasForwardedPorts = true
				break portManifestGroupSearchLoop
			}
		}
	}
	if hasForwardedPorts {
		result.ForwardedPorts, err = s.client.ForwardedPortStatus(ctx, leaseID)
		if err != nil {
			return result, err
		}
	}

	result.Services, err = s.client.LeaseStatus(ctx, leaseID)
	if err != nil {
		if errors.Is(err, kubeclienterrors.ErrNoDeploymentForLease) {
			return result, err
		}
		if errors.Is(err, kubeclienterrors.ErrLeaseNotFound) {
			return result, err
		}
		if kubeErrors.IsNotFound(err) {
			return result, err
		}
		return result, err
	}

	// Add hostname verification tokens and status to the response
	// Query ProviderHost CRDs to get verification tokens and check DNS status
	ac, err := fromctx.SubnetClientFromCtx(ctx)
	if err == nil {
		labelSelector := &strings.Builder{}
		kubeSelectorForLease(labelSelector, leaseID)
		phList, err := ac.SubnetV1().ProviderHosts("lease").List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		})
		if err == nil && len(phList.Items) > 0 {
			hostnameVerification := make(map[string]apclient.HostnameVerificationInfo) // hostname -> verification info

			for _, ph := range phList.Items {
				hostname := ph.Spec.Hostname
				var token string
				if ph.Annotations != nil {
					if t, ok := ph.Annotations[builder.SubnetNodeVerificationToken]; ok {
						token = t
					}
				}

				skipDNSVerification := ph.Annotations != nil && ph.Annotations[builder.SubnetNodeSkipDNSVerification] == "true"
				if skipDNSVerification {
					hostnameVerification[hostname] = apclient.HostnameVerificationInfo{
						Token:    token,
						Verified: true,
						Message:  "verification skipped (provider-managed hostname)",
					}
					continue
				}

				// Read verification status from ProviderHost annotations (set by hostname operator)
				// This avoids per-request DNS lookups
				var verified bool
				var message string
				if ph.Annotations != nil {
					if status, ok := ph.Annotations[builder.SubnetNodeDNSVerificationStatus]; ok {
						verified = status == "verified"
						if msg, ok := ph.Annotations[builder.SubnetNodeDNSVerificationMessage]; ok {
							message = msg
						} else {
							message = "verification status unknown"
						}
					} else {
						// No status stored yet - fallback to DNS lookup (shouldn't happen in normal operation)
						verified, message = checkDNSVerificationStatus(ctx, hostname, token)
					}
				} else {
					// No annotations - fallback to DNS lookup
					verified, message = checkDNSVerificationStatus(ctx, hostname, token)
				}

				hostnameVerification[hostname] = apclient.HostnameVerificationInfo{
					Token:    token,
					Verified: verified,
					Message:  message,
				}
			}

			if len(hostnameVerification) > 0 {
				// Add verification info to response as a new field
				result.HostnameVerification = hostnameVerification
			}
		}
	}

	expiryStatus, err := s.expiryService.CheckDeploymentExpiry(ctx, fmt.Sprintf("%d", leaseID.DSeq))
	if err != nil {
		return result, err
	}
	result.Status = expiryStatus.Status
	result.TimeLeft = expiryStatus.TimeLeft

	return result, nil
}

func (s *service) DeleteDeployment(lid mtypes.LeaseID) error {
	if err := s.bus.Publish(mtypes.EventLeaseClosed{
		ID: lid,
	}); err != nil {
		return fmt.Errorf("send lease closed request failed: %w", err)
	}

	return nil
}

// kubeSelectorForLease builds a Kubernetes label selector string for a lease ID
func kubeSelectorForLease(dst *strings.Builder, lID mtypes.LeaseID) {
	_, _ = fmt.Fprintf(dst, "%s=%s", builder.SubnetNodeLeaseOwnerLabelName, lID.Owner)
	_, _ = fmt.Fprintf(dst, ",%s=%d", builder.SubnetNodeLeaseDSeqLabelName, lID.DSeq)
	_, _ = fmt.Fprintf(dst, ",%s=%d", builder.SubnetNodeLeaseGSeqLabelName, lID.GSeq)
	_, _ = fmt.Fprintf(dst, ",%s=%d", builder.SubnetNodeLeaseOSeqLabelName, lID.OSeq)
}

// checkDNSVerificationStatus checks if the DNS TXT record for the hostname matches the token
func checkDNSVerificationStatus(ctx context.Context, hostname, token string) (bool, string) {
	verified, message := util.VerifyDNSVerification(ctx, hostname, token, 10*time.Second)
	return verified, message
}
