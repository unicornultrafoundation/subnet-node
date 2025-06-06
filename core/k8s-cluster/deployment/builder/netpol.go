package builder

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NetPol represents a Network Policy configuration
type NetPol struct {
	baseBuilder
	ingressRules []networkingv1.NetworkPolicyIngressRule
	egressRules  []networkingv1.NetworkPolicyEgressRule
	policyTypes  []networkingv1.PolicyType
	ports        []int32
}

// NewNetPol creates a new Network Policy configuration
func NewNetPol(b baseBuilder) *NetPol {
	return &NetPol{
		baseBuilder: b,
		policyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
		ports: make([]int32, 0),
	}
}

// Name returns the name of the Network Policy
func (b *NetPol) Name() string {
	return fmt.Sprintf("%s-netpol", b.deployment.Name)
}

// NS returns the namespace of the Network Policy
func (b *NetPol) NS() string {
	return b.deployment.Namespace
}

// AddPort adds a port to the Network Policy
func (b *NetPol) AddPort(port int32) {
	b.ports = append(b.ports, port)
}

// AddIngressRule adds an ingress rule to the Network Policy
func (b *NetPol) AddIngressRule(rule networkingv1.NetworkPolicyIngressRule) {
	b.ingressRules = append(b.ingressRules, rule)
}

// AddEgressRule adds an egress rule to the Network Policy
func (b *NetPol) AddEgressRule(rule networkingv1.NetworkPolicyEgressRule) {
	b.egressRules = append(b.egressRules, rule)
}

// Create creates a new Network Policy
func (b *NetPol) Create() (*networkingv1.NetworkPolicy, error) {
	if !b.settings.NetworkPoliciesEnabled {
		return nil, nil
	}

	netpol := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name(),
			Namespace: b.NS(),
			Labels:    b.labels(),
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: b.labels(),
			},
			PolicyTypes: b.policyTypes,
			Ingress:     b.ingressRules,
			Egress:      b.egressRules,
		},
	}

	return netpol, nil
}

// Update updates an existing Network Policy
func (b *NetPol) Update(obj *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	if !b.settings.NetworkPoliciesEnabled {
		return nil, nil
	}

	obj.Labels = b.labels()
	obj.Spec.PodSelector = metav1.LabelSelector{
		MatchLabels: b.labels(),
	}
	obj.Spec.PolicyTypes = b.policyTypes
	obj.Spec.Ingress = b.ingressRules
	obj.Spec.Egress = b.egressRules

	return obj, nil
}

// labels returns the labels for the Network Policy
func (b *NetPol) labels() map[string]string {
	return map[string]string{
		"app": b.deployment.Name,
	}
}

// Validate validates the Network Policy configuration
func (b *NetPol) Validate() error {
	if !b.settings.NetworkPoliciesEnabled {
		return nil
	}

	if len(b.ingressRules) == 0 && len(b.egressRules) == 0 {
		return fmt.Errorf("network policy must have at least one ingress or egress rule")
	}

	return nil
}

// buildPorts builds the network policy ports
func (b *NetPol) buildPorts() []networkingv1.NetworkPolicyPort {
	ports := make([]networkingv1.NetworkPolicyPort, 0, len(b.ports))
	for _, port := range b.ports {
		ports = append(ports, networkingv1.NetworkPolicyPort{
			Protocol: nil, // Allow both TCP and UDP
			Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: port},
		})
	}
	return ports
}
