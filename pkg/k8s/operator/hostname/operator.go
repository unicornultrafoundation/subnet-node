package hostname

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/pager"

	"github.com/sirupsen/logrus"
	manifest "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	ethereumcommon "github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/builder"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/clientcommon"
	kubeclienterrors "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/errors"
	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	chostname "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/hostname"
	clusterutil "github.com/unicornultrafoundation/subnet-node/core/k8s/util"
	crd "github.com/unicornultrafoundation/subnet-node/pkg/k8s/apis/subnet.node/v1"
	subnetclientset "github.com/unicornultrafoundation/subnet-node/pkg/k8s/client/clientset/versioned"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/common"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

var (
	errExpectedResourceNotFound = fmt.Errorf("%w: resource not found", common.ErrObservationStopped)
)

type hostnameOperator struct {
	ctx                context.Context
	hostnames          map[string]managedHostname
	leasesIgnored      common.IgnoreList
	ns                 string
	log                *logrus.Logger
	kc                 kubernetes.Interface
	ac                 subnetclientset.Interface
	cfg                common.OperatorConfig
	server             common.OperatorHTTP
	flagHostnamesData  common.PrepareFlagFn
	flagIgnoreListData common.PrepareFlagFn
	dnsVerifyTimeout   time.Duration
	dnsVerifyInterval  uint32
}

func newHostnameOperator(ctx context.Context, logger *logrus.Logger, ns string, config common.OperatorConfig, ilc common.IgnoreListConfig) (*hostnameOperator, error) {
	kc, err := fromctx.KubeClientFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	ac, err := fromctx.SubnetClientFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	opHTTP, err := common.NewOperatorHTTP()
	if err != nil {
		return nil, err
	}

	op := &hostnameOperator{
		ctx:               ctx,
		hostnames:         make(map[string]managedHostname),
		ns:                ns,
		log:               logger,
		kc:                kc,
		ac:                ac,
		cfg:               config,
		server:            opHTTP,
		leasesIgnored:     common.NewIgnoreList(ilc),
		dnsVerifyTimeout:  10 * time.Second,
		dnsVerifyInterval: 1,
	}

	op.flagIgnoreListData = op.server.AddPreparedEndpoint("/ignore-list", op.prepareIgnoreListData)
	op.flagHostnamesData = op.server.AddPreparedEndpoint("/managed-hostnames", op.prepareHostnamesData)

	return op, nil
}

func (op *hostnameOperator) run() error {
	op.log.Debug("hostname operator start")

	for {
		lastAttempt := time.Now()
		err := op.monitorUntilError()
		if errors.Is(err, context.Canceled) {
			op.log.Debug("hostname operator terminate")
			return err
		}

		op.log.Error("observation stopped", "err", err)

		// don't spin if there is a condition causing fast failure
		elapsed := time.Since(lastAttempt)
		if elapsed < op.cfg.RetryDelay {
			op.log.Info("delaying")
			select {
			case <-op.ctx.Done():
				return op.ctx.Err()
			case <-time.After(op.cfg.RetryDelay):
				// delay complete
			}
		}
	}
}

func (op *hostnameOperator) monitorUntilError() error {
	op.hostnames = make(map[string]managedHostname)
	ctx, cancel := context.WithCancel(op.ctx)
	defer cancel()

	op.log.Info("starting observation")

	connections, err := op.getHostnameDeploymentConnections(ctx)
	if err != nil {
		op.log.Error("unable to get connections", "err", err.Error())
		return err
	}

	for _, conn := range connections {
		leaseID := conn.GetLeaseID()
		hostname := conn.GetHostname()
		entry := managedHostname{
			lastEvent:           nil,
			presentLease:        leaseID,
			presentServiceName:  conn.GetServiceName(),
			presentExternalPort: uint32(conn.GetExternalPort()), // nolint: gosec
		}

		op.hostnames[hostname] = entry
		op.log.Debug("identified existing hostname connection",
			"hostname", hostname,
			"lease", entry.presentLease,
			"service", entry.presentServiceName,
			"port", entry.presentExternalPort)
	}
	op.flagHostnamesData()

	events, err := op.observeHostnameState(ctx)
	if err != nil {
		return err
	}

	pruneTicker := time.NewTicker(op.cfg.PruneInterval)
	defer pruneTicker.Stop()
	prepareTicker := time.NewTicker(op.cfg.WebRefreshInterval)
	defer prepareTicker.Stop()
	dnsVerifyTicker := time.NewTicker(time.Duration(op.dnsVerifyInterval) * time.Minute)
	defer dnsVerifyTicker.Stop()

	var exitError error
loop:
	for {
		select {
		case <-ctx.Done():
			exitError = ctx.Err()
			break loop
		case ev, ok := <-events:
			if !ok {
				exitError = common.ErrObservationStopped
				break loop
			}
			err = op.applyEvent(ctx, ev)
			if err != nil {
				op.log.Error("failed applying event", "err", err)
				exitError = err
				break loop
			}
		case <-pruneTicker.C:
			op.prune()
		case <-prepareTicker.C:
			if err := op.server.PrepareAll(); err != nil {
				op.log.Error("preparing web data failed", "err", err)
			}
		case <-dnsVerifyTicker.C:
			op.verifyAllHostnames(ctx)
		}
	}

	op.log.Debug("hostname operator done")

	return exitError
}

func (op *hostnameOperator) prepareIgnoreListData(pd common.PreparedResult) error {
	op.log.Debug("preparing ignore-list")
	return op.leasesIgnored.Prepare(pd)
}

func (op *hostnameOperator) prepareHostnamesData(pd common.PreparedResult) error {
	op.log.Debug("preparing managed-hostnames")
	buf := &bytes.Buffer{}
	data := make(map[string]interface{})

	for hostname, entry := range op.hostnames {
		preparedEntry := struct {
			LeaseID      mtypes.LeaseID
			Namespace    string
			ExternalPort uint32
			ServiceName  string
			LastUpdate   string
		}{
			LeaseID:      entry.presentLease,
			Namespace:    clusterutil.LeaseIDToNamespace(entry.presentLease),
			ExternalPort: entry.presentExternalPort,
			ServiceName:  entry.presentServiceName,
			LastUpdate:   entry.lastChangeAt.String(),
		}
		data[hostname] = preparedEntry
	}

	enc := json.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return err
	}

	pd.Set(buf.Bytes())
	return nil
}

func (op *hostnameOperator) prune() {
	if op.leasesIgnored.Prune() {
		op.flagIgnoreListData()
	}
}

func errorIsKubernetesResourceNotFound(failure error) bool {
	// check the error, only consider errors that are obviously
	// indicating a missing resource
	// otherwise simple errors like network issues could wind up with all CRDs
	// being ignored

	if kerrors.IsNotFound(failure) {
		return true
	}

	if errors.Is(failure, errExpectedResourceNotFound) {
		return true
	}

	errStr := failure.Error()
	// unless the error indicates a resource was not found, no action
	return strings.Contains(errStr, "not found")
}

func (op *hostnameOperator) recordEventError(ev chostname.ResourceEvent, failure error) {
	// no error, no action
	if failure == nil {
		return
	}

	mark := errorIsKubernetesResourceNotFound(failure)

	if !mark {
		return
	}

	op.log.Info("recording error for", "lease", ev.GetLeaseID().String(), "err", failure)

	op.leasesIgnored.AddError(mtypes.LeaseIDToKey(ev.GetLeaseID()), failure, ev.GetHostname())
	op.flagIgnoreListData()
}

func (op *hostnameOperator) isEventIgnored(ev chostname.ResourceEvent) bool {
	return op.leasesIgnored.IsFlagged(mtypes.LeaseIDToKey(ev.GetLeaseID()))
}

func (op *hostnameOperator) applyEvent(ctx context.Context, ev chostname.ResourceEvent) error {
	op.log.Debug("apply event", "event-type", ev.GetEventType(), "hostname", ev.GetHostname())
	switch ev.GetEventType() {
	case ctypes.ProviderResourceDelete:
		// note that on delete the resource might be gone anyways because the namespace is deleted
		return op.applyDeleteEvent(ctx, ev)
	case ctypes.ProviderResourceAdd, ctypes.ProviderResourceUpdate:
		if op.isEventIgnored(ev) {
			op.log.Info("ignoring event for", "lease", ev.GetLeaseID().String())
			return nil
		}
		err := op.applyAddOrUpdateEvent(ctx, ev)
		op.recordEventError(ev, err)
		return err
	default:
		return fmt.Errorf("%w: unknown event type %v", common.ErrObservationStopped, ev.GetEventType())
	}

}

func (op *hostnameOperator) applyDeleteEvent(ctx context.Context, ev chostname.ResourceEvent) error {
	leaseID := ev.GetLeaseID()
	err := op.removeHostnameFromDeployment(ctx, ev.GetHostname(), leaseID, true)

	if err == nil {
		delete(op.hostnames, ev.GetHostname())
		op.flagHostnamesData()
	}

	return err
}

func buildDirective(ev chostname.ResourceEvent, serviceExpose crd.ManifestServiceExpose) chostname.ConnectToDeploymentDirective {
	// Build the directive based off the event
	directive := chostname.ConnectToDeploymentDirective{
		Hostname:    ev.GetHostname(),
		LeaseID:     ev.GetLeaseID(),
		ServiceName: ev.GetServiceName(),
		ServicePort: int32(ev.GetExternalPort()), // nolint: gosec
	}
	/*
		Populate the configuration options
		selectedExpose.HttpOptions has zero values if this is from an earlier CRD. Just insert
		defaults and move on
	*/
	if serviceExpose.HTTPOptions.MaxBodySize == 0 {
		directive.ReadTimeout = 60000
		directive.SendTimeout = 60000
		directive.NextTimeout = 60000
		directive.MaxBodySize = 1048576
		directive.NextTries = 3
		directive.NextCases = []string{"error", "timeout"}
	} else {
		directive.ReadTimeout = serviceExpose.HTTPOptions.ReadTimeout
		directive.SendTimeout = serviceExpose.HTTPOptions.SendTimeout
		directive.NextTimeout = serviceExpose.HTTPOptions.NextTimeout
		directive.MaxBodySize = serviceExpose.HTTPOptions.MaxBodySize
		directive.NextTries = serviceExpose.HTTPOptions.NextTries
		directive.NextCases = serviceExpose.HTTPOptions.NextCases
	}

	return directive
}

func (op *hostnameOperator) locateServiceFromManifest(ctx context.Context, leaseID mtypes.LeaseID, serviceName string, externalPort uint32) (crd.ManifestServiceExpose, error) {
	// Locate the matching service name & expose directive in the manifest CRD
	found, manifestGroup, err := op.getManifestGroup(ctx, leaseID)
	if err != nil {
		return crd.ManifestServiceExpose{}, err
	}
	if !found {
		/*
			It's possible this code could race to read the CRD, although unlikely. If this fails the operator
			restarts and should work the attempt anyways. If this becomes a pain point then the operator
			can be rewritten to watch for CRD events on the manifest as well, then avoid running this code
			until the manifest exists.
		*/
		return crd.ManifestServiceExpose{}, fmt.Errorf("%w: manifest for %v", errExpectedResourceNotFound, leaseID)
	}

	var selectedService crd.ManifestService
	for _, service := range manifestGroup.Services {
		if service.Name == serviceName {
			selectedService = service
			break
		}
	}

	if selectedService.Count == 0 {
		return crd.ManifestServiceExpose{}, fmt.Errorf("%w: service for %v - %v", errExpectedResourceNotFound, leaseID, serviceName)
	}

	var selectedExpose crd.ManifestServiceExpose
	for _, expose := range selectedService.Expose {
		if !expose.Global {
			continue
		}

		mse := manifest.ServiceExpose{
			Port:         uint32(expose.Port),
			ExternalPort: uint32(expose.ExternalPort),
		}

		if externalPort == uint32(mse.GetExternalPort()) { // nolint: gosec
			selectedExpose = expose
			break
		}
	}

	if selectedExpose.Port == 0 {
		return crd.ManifestServiceExpose{}, fmt.Errorf("%w: service expose for %v - %v - %v", errExpectedResourceNotFound, leaseID, serviceName, externalPort)
	}

	return selectedExpose, nil
}

func (op *hostnameOperator) applyAddOrUpdateEvent(ctx context.Context, ev chostname.ResourceEvent) error {
	selectedExpose, err := op.locateServiceFromManifest(ctx, ev.GetLeaseID(), ev.GetServiceName(), ev.GetExternalPort())
	if err != nil {
		return err
	}

	leaseID := ev.GetLeaseID()

	op.log.Debug("connecting",
		"hostname", ev.GetHostname(),
		"lease", leaseID,
		"service", ev.GetServiceName(),
		"externalPort", ev.GetExternalPort())
	entry, exists := op.hostnames[ev.GetHostname()]

	isSameLease := false
	if exists {
		isSameLease = entry.presentLease.Equals(leaseID)
	} else {
		isSameLease = true
	}

	// Verify DNS before creating/updating Ingress
	// Always verify on both Add and Update events to ensure DNS is still valid
	if err := op.verifyHostnameDNS(ctx, ev.GetHostname()); err != nil {
		op.log.
			WithField("hostname", ev.GetHostname()).
			WithField("lease", leaseID).
			WithField("event-type", ev.GetEventType()).
			WithField("error", err).
			Warn("DNS verification failed, removing ingress if it exists")

		// If Ingress already exists but verification now fails, remove it
		if exists {
			op.log.
				WithField("hostname", ev.GetHostname()).
				WithField("lease", entry.presentLease).
				Info("removing ingress due to failed DNS verification")
			if removeErr := op.removeHostnameFromDeployment(ctx, ev.GetHostname(), entry.presentLease, true); removeErr != nil {
				op.log.
					WithField("hostname", ev.GetHostname()).
					WithField("error", removeErr).
					Error("failed to remove ingress after DNS verification failure")
			} else {
				// Remove from tracking map
				delete(op.hostnames, ev.GetHostname())
				op.flagHostnamesData()
			}
		}

		// Don't create/update Ingress if verification fails
		return fmt.Errorf("DNS verification failed for hostname %q: %w", ev.GetHostname(), err)
	}

	directive := buildDirective(ev, selectedExpose)

	if isSameLease {
		// shouldConnect := false

		if !exists {
			// shouldConnect = true
			op.log.Debug("hostname target is new, applying")
			// Check to see if port or service name is different
		} else if entry.presentExternalPort != ev.GetExternalPort() || entry.presentServiceName != ev.GetServiceName() {
			// shouldConnect = true
			op.log.Debug("hostname target has changed, applying")
		}

		// if shouldConnect {
		op.log.Debug("Updating ingress")
		// Update or create the existing ingress
		err = op.connectHostnameToDeployment(ctx, directive)
		// }
	} else {
		op.log.Debug("Swapping ingress to new deployment")
		//  Delete the ingress in one namespace and recreate it in the correct one
		err = op.removeHostnameFromDeployment(ctx, ev.GetHostname(), entry.presentLease, false)
		if err == nil {
			// Remove the current entry, if the next action succeeds then it gets inserted below
			delete(op.hostnames, ev.GetHostname())
			err = op.connectHostnameToDeployment(ctx, directive)
		}
	}

	if err == nil { // Update stored entry if everything went OK
		entry.presentExternalPort = ev.GetExternalPort()
		entry.presentServiceName = ev.GetServiceName()
		entry.presentLease = leaseID
		entry.lastEvent = ev
		entry.lastChangeAt = time.Now()
		op.hostnames[ev.GetHostname()] = entry
		op.flagHostnamesData()
	}

	return err
}

func (op *hostnameOperator) getHostnameDeploymentConnections(ctx context.Context) ([]chostname.LeaseIDConnection, error) {
	ingressPager := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
		return op.kc.NetworkingV1().Ingresses(metav1.NamespaceAll).List(ctx, opts)
	})

	results := make([]chostname.LeaseIDConnection, 0)
	err := ingressPager.EachListItem(ctx,
		metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=true", builder.SubnetNodeManagedLabelName)},
		func(obj runtime.Object) error {
			ingress := obj.(*netv1.Ingress)
			ingressLeaseID, err := clientcommon.RecoverLeaseIDFromLabels(ingress.Labels)
			if err != nil {
				return err
			}
			if len(ingress.Spec.Rules) != 1 {
				return fmt.Errorf("%w: invalid number of rules %d", kubeclienterrors.ErrInvalidHostnameConnection, len(ingress.Spec.Rules))
			}
			rule := ingress.Spec.Rules[0]

			if len(rule.HTTP.Paths) != 1 {
				return fmt.Errorf("%w: invalid number of paths %d", kubeclienterrors.ErrInvalidHostnameConnection, len(rule.HTTP.Paths))
			}
			rulePath := rule.HTTP.Paths[0]
			results = append(results, leaseIDHostnameConnection{
				leaseID:      ingressLeaseID,
				hostname:     rule.Host,
				externalPort: rulePath.Backend.Service.Port.Number,
				serviceName:  rulePath.Backend.Service.Name,
			})

			return nil
		})

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (op *hostnameOperator) observeHostnameState(ctx context.Context) (<-chan chostname.ResourceEvent, error) {
	phpager := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
		resources, err := op.ac.SubnetV1().ProviderHosts(op.ns).List(ctx, opts)
		return resources, err
	})

	data := make([]crd.ProviderHost, 0, 128)
	err := phpager.EachListItem(ctx, metav1.ListOptions{}, func(obj runtime.Object) error {
		ph := obj.(*crd.ProviderHost)
		data = append(data, *ph)
		return nil
	})

	if err != nil {
		return nil, err
	}

	op.log.Info("starting hostname watch")
	watcher, err := op.ac.SubnetV1().ProviderHosts(op.ns).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	evData := make([]hostnameResourceEvent, len(data))
	for i, v := range data {
		if !ethereumcommon.IsHexAddress(v.Spec.Owner) {
			return nil, fmt.Errorf("invalid owner address")
		}
		ownerAddr := ethereumcommon.HexToAddress(v.Spec.Owner)
		if !ethereumcommon.IsHexAddress(v.Spec.Provider) {
			return nil, fmt.Errorf("invalid provider address")
		}
		providerAddr := ethereumcommon.HexToAddress(v.Spec.Provider)
		ev := hostnameResourceEvent{
			eventType:    ctypes.ProviderResourceAdd,
			hostname:     v.Spec.Hostname,
			oseq:         v.Spec.Oseq,
			gseq:         v.Spec.Gseq,
			dseq:         v.Spec.Dseq,
			owner:        ownerAddr,
			provider:     providerAddr,
			serviceName:  v.Spec.ServiceName,
			externalPort: v.Spec.ExternalPort,
		}
		evData[i] = ev
	}

	data = nil

	output := make(chan chostname.ResourceEvent)

	go func() {
		defer func() {
			close(output)
			watcher.Stop()
		}()

		for _, v := range evData {
			output <- v
		}
		evData = nil // do not hold the reference

		results := watcher.ResultChan()
		for {
			select {
			case result, ok := <-results:
				if !ok { // Channel closed when an error happens
					return
				}
				ph := result.Object.(*crd.ProviderHost)
				if !ethereumcommon.IsHexAddress(ph.Spec.Owner) {
					op.log.Error("invalid owner address in provider host", "addr", ph.Spec.Owner)
					continue // Ignore event
				}
				ownerAddr := ethereumcommon.HexToAddress(ph.Spec.Owner)
				if !ethereumcommon.IsHexAddress(ph.Spec.Provider) {
					op.log.Error("invalid provider address in provider host", "addr", ph.Spec.Provider)
					continue // Ignore event
				}
				providerAddr := ethereumcommon.HexToAddress(ph.Spec.Provider)
				ev := hostnameResourceEvent{
					hostname:     ph.Spec.Hostname,
					dseq:         ph.Spec.Dseq,
					oseq:         ph.Spec.Oseq,
					gseq:         ph.Spec.Gseq,
					owner:        ownerAddr,
					provider:     providerAddr,
					serviceName:  ph.Spec.ServiceName,
					externalPort: ph.Spec.ExternalPort,
				}
				switch result.Type {

				case watch.Added:
					ev.eventType = ctypes.ProviderResourceAdd
				case watch.Modified:
					ev.eventType = ctypes.ProviderResourceUpdate
				case watch.Deleted:
					ev.eventType = ctypes.ProviderResourceDelete

				case watch.Error:
					// Based on examination of the implementation code, this is basically never called anyways
					op.log.Error("watch error", "err", result.Object)

				default:

					continue
				}

				output <- ev
			case <-ctx.Done():
				return
			}
		}
	}()

	return output, nil
}

func (op *hostnameOperator) getManifestGroup(ctx context.Context, lID mtypes.LeaseID) (bool, crd.ManifestGroup, error) {
	leaseNamespace := builder.LidNS(lID)

	obj, err := op.ac.SubnetV1().Manifests(op.ns).Get(ctx, leaseNamespace, metav1.GetOptions{})

	if err != nil {
		if kerrors.IsNotFound(err) {
			op.log.Info("CRD manifest not found", "lease-ns", leaseNamespace)
			return false, crd.ManifestGroup{}, nil
		}

		return false, crd.ManifestGroup{}, err
	}

	return true, obj.Spec.Group, nil
}

func (op *hostnameOperator) removeHostnameFromDeployment(ctx context.Context, hostname string, leaseID mtypes.LeaseID, allowMissing bool) error {
	ns := builder.LidNS(leaseID)
	labelSelector := &strings.Builder{}

	kubeSelectorForLease(labelSelector, leaseID)

	fieldSelector := &strings.Builder{}
	_, _ = fmt.Fprintf(fieldSelector, "metadata.name=%s", hostname)

	// This delete only works if the ingress exists & the labels match the lease ID given
	err := op.kc.NetworkingV1().Ingresses(ns).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		TypeMeta:             metav1.TypeMeta{},
		LabelSelector:        labelSelector.String(),
		FieldSelector:        fieldSelector.String(),
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       nil,
		Limit:                0,
		Continue:             "",
	})

	if err != nil && allowMissing && kerrors.IsNotFound(err) {
		return nil
	}

	return err
}

func (op *hostnameOperator) connectHostnameToDeployment(ctx context.Context, directive chostname.ConnectToDeploymentDirective) error {
	ingressName := directive.Hostname
	ns := builder.LidNS(directive.LeaseID)
	rules := ingressRules(directive.Hostname, directive.ServiceName, directive.ServicePort)

	foundEntry, err := op.kc.NetworkingV1().Ingresses(ns).Get(ctx, ingressName, metav1.GetOptions{})
	// metricsutils.IncCounterVecWithLabelValuesFiltered(kubeCallsCounter, "ingresses-get", err, kubeErrors.IsNotFound)

	labels := make(map[string]string)
	labels[builder.SubnetNodeManagedLabelName] = "true"
	builder.AppendLeaseLabels(directive.LeaseID, labels)

	ingressClassName := subnetNodeIngressClassName
	obj := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingressName,
			Labels:      labels,
			Annotations: kubeNginxIngressAnnotations(directive),
		},
		Spec: netv1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules:            rules,
		},
	}

	switch {
	case err == nil:
		obj.ResourceVersion = foundEntry.ResourceVersion
		_, err = op.kc.NetworkingV1().Ingresses(ns).Update(ctx, obj, metav1.UpdateOptions{})
		// metricsutils.IncCounterVecWithLabelValues(kubeCallsCounter, "networking-ingresses-update", err)
	case kerrors.IsNotFound(err):
		_, err = op.kc.NetworkingV1().Ingresses(ns).Create(ctx, obj, metav1.CreateOptions{})
		// metricsutils.IncCounterVecWithLabelValues(kubeCallsCounter, "networking-ingresses-create", err)
	}

	return err
}

func ingressRules(hostname string, kubeServiceName string, kubeServicePort int32) []netv1.IngressRule {
	// for some reason we need to pass a pointer to this
	pathTypeForAll := netv1.PathTypePrefix
	ruleValue := netv1.HTTPIngressRuleValue{
		Paths: []netv1.HTTPIngressPath{{
			Path:     "/",
			PathType: &pathTypeForAll,
			Backend: netv1.IngressBackend{
				Service: &netv1.IngressServiceBackend{
					Name: kubeServiceName,
					Port: netv1.ServiceBackendPort{
						Number: kubeServicePort,
					},
				},
			},
		}},
	}

	return []netv1.IngressRule{{
		Host:             hostname,
		IngressRuleValue: netv1.IngressRuleValue{HTTP: &ruleValue},
	}}
}

func kubeNginxIngressAnnotations(directive chostname.ConnectToDeploymentDirective) map[string]string {
	// For kubernetes/ingress-nginx
	// https://github.com/kubernetes/ingress-nginx
	const root = "nginx.ingress.kubernetes.io"

	readTimeout := math.Ceil(float64(directive.ReadTimeout) / 1000.0)
	sendTimeout := math.Ceil(float64(directive.SendTimeout) / 1000.0)
	result := map[string]string{
		fmt.Sprintf("%s/proxy-read-timeout", root): fmt.Sprintf("%d", int(readTimeout)),
		fmt.Sprintf("%s/proxy-send-timeout", root): fmt.Sprintf("%d", int(sendTimeout)),

		fmt.Sprintf("%s/proxy-next-upstream-tries", root): strconv.Itoa(int(directive.NextTries)),
		fmt.Sprintf("%s/proxy-body-size", root):           strconv.Itoa(int(directive.MaxBodySize)),
	}

	nextTimeoutKey := fmt.Sprintf("%s/proxy-next-upstream-timeout", root)
	nextTimeout := 0 // default magic value for disable
	if directive.NextTimeout > 0 {
		nextTimeout = int(math.Ceil(float64(directive.NextTimeout) / 1000.0))
	}

	result[nextTimeoutKey] = fmt.Sprintf("%d", nextTimeout)

	strBuilder := strings.Builder{}

	for i, v := range directive.NextCases {
		first := string(v[0])
		isHTTPCode := strings.ContainsAny(first, "12345")

		if isHTTPCode {
			strBuilder.WriteString("http_")
		}
		strBuilder.WriteString(v)

		if i != len(directive.NextCases)-1 {
			// The actual separator is the space character for kubernetes/ingress-nginx
			strBuilder.WriteRune(' ')
		}
	}

	result[fmt.Sprintf("%s/proxy-next-upstream", root)] = strBuilder.String()
	return result
}

func kubeSelectorForLease(dst *strings.Builder, lID mtypes.LeaseID) {
	_, _ = fmt.Fprintf(dst, "%s=%s", builder.SubnetNodeLeaseOwnerLabelName, lID.Owner)
	_, _ = fmt.Fprintf(dst, ",%s=%d", builder.SubnetNodeLeaseDSeqLabelName, lID.DSeq)
	_, _ = fmt.Fprintf(dst, ",%s=%d", builder.SubnetNodeLeaseGSeqLabelName, lID.GSeq)
	_, _ = fmt.Fprintf(dst, ",%s=%d", builder.SubnetNodeLeaseOSeqLabelName, lID.OSeq)
}

// verifyHostnameDNS verifies that the hostname has the correct verification token in its DNS TXT record
func (op *hostnameOperator) verifyHostnameDNS(ctx context.Context, hostname string) error {
	// Get the ProviderHost CRD to retrieve the verification token
	ph, err := op.ac.SubnetV1().ProviderHosts(op.ns).Get(ctx, hostname, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ProviderHost: %w", err)
	}

	// Skip verification for provider-managed hostnames
	if ph.Annotations != nil {
		if skip, ok := ph.Annotations[builder.SubnetNodeSkipDNSVerification]; ok && skip == "true" {
			op.log.
				WithField("hostname", hostname).
				Debug("skipping DNS verification for provider-managed hostname")
			// Update status to indicate verification is skipped
			op.updateVerificationStatus(ctx, hostname, true, "verification skipped (provider-managed hostname)")
			return nil
		}
	}

	// Get token from annotations
	var token string
	if ph.Annotations != nil {
		if t, ok := ph.Annotations[builder.SubnetNodeVerificationToken]; ok {
			token = t
		}
	}

	if token == "" {
		errMsg := fmt.Errorf("no verification token found for hostname %q: token should be generated automatically when hostname is first declared", hostname)
		err := op.updateVerificationStatus(ctx, hostname, false, errMsg.Error())
		if err != nil {
			return fmt.Errorf("failed to update verification status: %w", err)
		}
		return errMsg
	}

	// Use shared DNS verification utility
	verified, message := clusterutil.VerifyDNSVerification(ctx, hostname, token, op.dnsVerifyTimeout)

	if verified {
		op.log.
			WithField("hostname", hostname).
			Debug("DNS verification successful")
		op.updateVerificationStatus(ctx, hostname, true, message)
		return nil
	}

	// Verification failed
	op.updateVerificationStatus(ctx, hostname, false, message)
	return errors.New(message)
}

// updateVerificationStatus updates the ProviderHost CRD with DNS verification status
func (op *hostnameOperator) updateVerificationStatus(ctx context.Context, hostname string, verified bool, message string) error {
	ph, err := op.ac.SubnetV1().ProviderHosts(op.ns).Get(ctx, hostname, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ProviderHost for status update: %w", err)
	}

	if ph.Annotations == nil {
		ph.Annotations = make(map[string]string)
	}

	status := "failed"
	if verified {
		status = "verified"
	}

	ph.Annotations[builder.SubnetNodeDNSVerificationStatus] = status
	ph.Annotations[builder.SubnetNodeDNSVerificationMessage] = message
	ph.Annotations[builder.SubnetNodeDNSVerificationTimestamp] = time.Now().UTC().Format(time.RFC3339)

	_, err = op.ac.SubnetV1().ProviderHosts(op.ns).Update(ctx, ph, metav1.UpdateOptions{})
	if err != nil {
		op.log.
			WithField("hostname", hostname).
			WithField("error", err).
			Warn("failed to update ProviderHost verification status")
	} else {
		op.log.
			WithField("hostname", hostname).
			WithField("verified", verified).
			Debug("updated ProviderHost verification status")
	}

	return err
}

// verifyAllHostnames periodically verifies all active hostnames
// If verification fails, removes the Ingress to maintain consistency with event-driven verification
func (op *hostnameOperator) verifyAllHostnames(ctx context.Context) {
	op.log.Debug("starting periodic DNS verification for all hostnames")
	verifiedCount := 0
	failedCount := 0

	for hostname, entry := range op.hostnames {
		if err := op.verifyHostnameDNS(ctx, hostname); err != nil {
			op.log.
				WithField("hostname", hostname).
				WithField("lease", entry.presentLease).
				WithField("error", err).
				Warn("periodic DNS verification failed, removing ingress")
			failedCount++

			// Remove Ingress if verification fails (consistent with event-driven behavior)
			if removeErr := op.removeHostnameFromDeployment(ctx, hostname, entry.presentLease, true); removeErr != nil {
				op.log.
					WithField("hostname", hostname).
					WithField("error", removeErr).
					Error("failed to remove ingress after periodic DNS verification failure")
			} else {
				// Remove from tracking map
				delete(op.hostnames, hostname)
				op.flagHostnamesData()
			}
		} else {
			verifiedCount++
		}
	}

	op.log.
		WithField("total", len(op.hostnames)).
		WithField("verified", verifiedCount).
		WithField("failed", failedCount).
		Info("periodic DNS verification completed")
}
