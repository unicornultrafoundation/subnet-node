package fromctx

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/troian/pubsub"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	subnetclientset "github.com/unicornultrafoundation/subnet-node/pkg/k8s/client/clientset/versioned"
)

type Key string

const (
	CtxKeyKubeConfig         = Key("kubeconfig")
	CtxKeyKubeRESTClient     = Key("kube-restclient")
	CtxKeyKubeClientSet      = Key("kube-clientset")
	CtxKeySubnetClientSet    = Key("subnet-clientset")
	CtxKeyPubSub             = Key("pubsub")
	CtxKeyLifecycle          = Key("lifecycle")
	CtxKeyErrGroup           = Key("errgroup")
	CtxKeyLogc               = Key("logc")
	CtxKeyStartupCh          = Key("startup-ch")
	CtxKeyInventoryUnderTest = Key("inventory-under-test")
	CtxKeyPersistentConfig   = Key("persistent-config")
	CtxKeyCertIssuer         = Key("cert-issuer")
)

var (
	ErrNotFound         = errors.New("fromctx: not found")
	ErrValueInvalidType = errors.New("fromctx: invalid type")
)

type options struct {
	logName Key
}

type LogcOption func(*options) error

func CmdSetContextValue(cmd *cobra.Command, key, val interface{}) {
	cmd.SetContext(context.WithValue(cmd.Context(), key, val))
}

// WithLogc add logger object to the context
// key defaults to the "log"
// use WithLogName("<custom name>") to set custom key
func WithLogc(ctx context.Context, lg *logrus.Logger, opts ...LogcOption) context.Context {
	opt, _ := applyOptions(opts...)

	ctx = context.WithValue(ctx, opt.logName, lg)

	return ctx
}

func LogcFromCtx(ctx context.Context, opts ...LogcOption) *logrus.Logger {
	opt, _ := applyOptions(opts...)

	var logger *logrus.Logger
	if lg, valid := ctx.Value(opt.logName).(*logrus.Logger); valid {
		logger = lg
	}

	return logger
}

func LogrFromCtx(ctx context.Context) logr.Logger {
	lg, _ := logr.FromContext(ctx)
	return lg
}

func StartupChFromCtx(ctx context.Context) (chan<- struct{}, error) {
	val := ctx.Value(CtxKeyStartupCh)
	if val == nil {
		return nil, fmt.Errorf("%w: context does not have startup channel set", ErrNotFound)
	}

	res, valid := val.(chan<- struct{})
	if !valid {
		return nil, ErrValueInvalidType
	}

	return res, nil
}

func MustStartupChFromCtx(ctx context.Context) chan<- struct{} {
	val, err := StartupChFromCtx(ctx)
	if err != nil {
		panic(err.Error())
	}

	return val
}

func KubeRESTClientFromCtx(ctx context.Context) (*rest.RESTClient, error) {
	val := ctx.Value(CtxKeyKubeRESTClient)
	if val == nil {
		return nil, fmt.Errorf("%w: context does not have kube rest client set", ErrNotFound)
	}

	res, valid := val.(*rest.RESTClient)
	if !valid {
		return nil, ErrValueInvalidType
	}

	return res, nil
}

func MustKubeRESTClientFromCtx(ctx context.Context) *rest.RESTClient {
	val, err := KubeRESTClientFromCtx(ctx)
	if err != nil {
		panic(err.Error())
	}

	return val
}

func KubeConfigFromCtx(ctx context.Context) (*rest.Config, error) {
	val := ctx.Value(CtxKeyKubeConfig)
	if val == nil {
		return nil, fmt.Errorf("%w: context does not have kubeconfig set", ErrNotFound)
	}

	res, valid := val.(*rest.Config)
	if !valid {
		return nil, ErrValueInvalidType
	}

	return res, nil
}

func MustKubeConfigFromCtx(ctx context.Context) *rest.Config {
	val, err := KubeConfigFromCtx(ctx)
	if err != nil {
		panic(err.Error())
	}

	return val
}

func KubeClientFromCtx(ctx context.Context) (kubernetes.Interface, error) {
	val := ctx.Value(CtxKeyKubeClientSet)
	if val == nil {
		return nil, fmt.Errorf("%w: context does not have kube client set", ErrNotFound)
	}

	res, valid := val.(kubernetes.Interface)
	if !valid {
		return nil, ErrValueInvalidType
	}

	return res, nil
}

func MustKubeClientFromCtx(ctx context.Context) kubernetes.Interface {
	val, err := KubeClientFromCtx(ctx)
	if err != nil {
		panic(err.Error())
	}

	return val
}

func SubnetClientFromCtx(ctx context.Context) (subnetclientset.Interface, error) {
	val := ctx.Value(CtxKeySubnetClientSet)
	if val == nil {
		return nil, fmt.Errorf("%w: context does not have subnet client set", ErrNotFound)
	}

	res, valid := val.(subnetclientset.Interface)
	if !valid {
		return nil, ErrValueInvalidType
	}

	return res, nil
}

func MustSubnetClientFromCtx(ctx context.Context) subnetclientset.Interface {
	val, err := SubnetClientFromCtx(ctx)
	if err != nil {
		panic(err.Error())
	}

	return val
}

func ErrGroupFromCtx(ctx context.Context) (*errgroup.Group, error) {
	val := ctx.Value(CtxKeyErrGroup)
	if val == nil {
		return nil, fmt.Errorf("%w: context does not have errgroup set", ErrNotFound)
	}

	res, valid := val.(*errgroup.Group)
	if !valid {
		return nil, ErrValueInvalidType
	}

	return res, nil
}

func MustErrGroupFromCtx(ctx context.Context) *errgroup.Group {
	val, err := ErrGroupFromCtx(ctx)
	if err != nil {
		panic(err.Error())
	}

	return val
}

func PubSubFromCtx(ctx context.Context) (pubsub.PubSub, error) {
	val := ctx.Value(CtxKeyPubSub)
	if val == nil {
		return nil, fmt.Errorf("%w: context does not have pubsub set", ErrNotFound)
	}

	res, valid := val.(pubsub.PubSub)
	if !valid {
		return nil, ErrValueInvalidType
	}

	return res, nil
}

func MustPubSubFromCtx(ctx context.Context) pubsub.PubSub {
	val, err := PubSubFromCtx(ctx)
	if err != nil {
		panic(err.Error())
	}

	return val
}

func IsInventoryUnderTestFromCtx(ctx context.Context) bool {
	val := ctx.Value(CtxKeyInventoryUnderTest)
	if val == nil {
		return false
	}

	if v, valid := val.(bool); valid {
		return v
	}

	return false
}

func applyOptions(opts ...LogcOption) (options, error) {
	obj := &options{}
	for _, opt := range opts {
		if err := opt(obj); err != nil {
			return options{}, err
		}
	}

	if obj.logName == "" {
		obj.logName = CtxKeyLogc
	}

	return *obj, nil
}

func ApplyToContext(ctx context.Context, config map[interface{}]interface{}) context.Context {
	for k, v := range config {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx
}
