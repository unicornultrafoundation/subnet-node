package kube

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentStatus_TTL_TimeLeft(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	deploymentID := "test-ttl"
	creationTime := time.Now().Add(-10 * time.Minute) // created 10 minutes ago
	ttl := int64(30)                                  // 30 minutes

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentID,
			Annotations: map[string]string{
				"ttl": "30",
			},
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
	}
	_, err := fakeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	assert.NoError(t, err)

	// No deployments/services/ingresses needed for TTL test
	status, err := client.getDeploymentStatusInternal(context.Background(), deploymentID)
	assert.NoError(t, err)
	assert.Equal(t, ttl, status.TTL)
	// Should be 20 minutes left (30 - 10), allow 1 min margin for test timing
	assert.GreaterOrEqual(t, status.TimeLeft, int64(19))
	assert.LessOrEqual(t, status.TimeLeft, int64(21))

	// Test expired TTL
	nsExpired := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "expired-ttl",
			Annotations: map[string]string{
				"ttl": "5",
			},
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
		},
	}
	_, err = fakeClient.CoreV1().Namespaces().Create(context.Background(), nsExpired, metav1.CreateOptions{})
	assert.NoError(t, err)
	status, err = client.getDeploymentStatusInternal(context.Background(), "expired-ttl")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), status.TTL)
	assert.Equal(t, int64(0), status.TimeLeft)
}

func TestDeploymentStatus_TTL_TimeLeft_EdgeCases(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	t.Run("No TTL set", func(t *testing.T) {
		id := "no-ttl"
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              id,
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
		}
		_, err := fakeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		assert.NoError(t, err)
		status, err := client.getDeploymentStatusInternal(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), status.TTL)
		assert.Equal(t, int64(-1), status.TimeLeft)
	})

	t.Run("TTL is 0", func(t *testing.T) {
		id := "ttl-zero"
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              id,
				Annotations:       map[string]string{"ttl": "0"},
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
		}
		_, err := fakeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		assert.NoError(t, err)
		status, err := client.getDeploymentStatusInternal(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), status.TTL)
		assert.Equal(t, int64(0), status.TimeLeft)
	})

	t.Run("TTL is 1, just created", func(t *testing.T) {
		id := "ttl-one-fresh"
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              id,
				Annotations:       map[string]string{"ttl": "1"},
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
		}
		_, err := fakeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		assert.NoError(t, err)
		status, err := client.getDeploymentStatusInternal(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), status.TTL)
		assert.Equal(t, int64(1), status.TimeLeft)
	})

	t.Run("TTL is 1, just expired", func(t *testing.T) {
		id := "ttl-one-expired"
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              id,
				Annotations:       map[string]string{"ttl": "1"},
				CreationTimestamp: metav1.Time{Time: time.Now().Add(-2 * time.Minute)},
			},
		}
		_, err := fakeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		assert.NoError(t, err)
		status, err := client.getDeploymentStatusInternal(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), status.TTL)
		assert.Equal(t, int64(0), status.TimeLeft)
	})

	t.Run("TTL is negative", func(t *testing.T) {
		id := "ttl-negative"
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              id,
				Annotations:       map[string]string{"ttl": "-5"},
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
		}
		_, err := fakeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		assert.NoError(t, err)
		status, err := client.getDeploymentStatusInternal(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, int64(-5), status.TTL)
		assert.Equal(t, int64(0), status.TimeLeft)
	})

	t.Run("TTL is large, far in future", func(t *testing.T) {
		id := "ttl-large"
		ttl := int64(10000)
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              id,
				Annotations:       map[string]string{"ttl": "10000"},
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
		}
		_, err := fakeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		assert.NoError(t, err)
		status, err := client.getDeploymentStatusInternal(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, ttl, status.TTL)
		assert.GreaterOrEqual(t, status.TimeLeft, ttl-1)
		assert.LessOrEqual(t, status.TimeLeft, ttl)
	})
}
