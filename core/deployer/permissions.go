package k8scluster

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// verifyKubernetesPermissions verifies that the Kubernetes client has the required permissions
func (s *Service) verifyKubernetesPermissions(ctx context.Context) error {
	s.logger.Info("Verifying Kubernetes permissions")

	// Check namespace creation permission
	canCreateNS, err := s.client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:     "create",
				Resource: "namespaces",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		s.logger.Error("Failed to check namespace creation permission", zap.Error(err))
		return fmt.Errorf("failed to check namespace creation permission: %w", err)
	}
	if !canCreateNS.Status.Allowed {
		s.logger.Error("No permission to create namespaces")
		return fmt.Errorf("no permission to create namespaces")
	}

	// Check deployment creation permission
	canCreateDeploy, err := s.client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:     "create",
				Resource: "deployments",
				Group:    "apps",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		s.logger.Error("Failed to check deployment creation permission", zap.Error(err))
		return fmt.Errorf("failed to check deployment creation permission: %w", err)
	}
	if !canCreateDeploy.Status.Allowed {
		s.logger.Error("No permission to create deployments")
		return fmt.Errorf("no permission to create deployments")
	}

	// Check service creation permission
	canCreateSvc, err := s.client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:     "create",
				Resource: "services",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		s.logger.Error("Failed to check service creation permission", zap.Error(err))
		return fmt.Errorf("failed to check service creation permission: %w", err)
	}
	if !canCreateSvc.Status.Allowed {
		s.logger.Error("No permission to create services")
		return fmt.Errorf("no permission to create services")
	}

	s.logger.Info("Kubernetes permissions verified successfully")
	return nil
}
