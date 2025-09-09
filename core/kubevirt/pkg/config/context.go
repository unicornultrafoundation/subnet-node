package config

type (
	_scaledKey struct{}
)

type Options struct {
	Namespace       string
	Threadiness     int
	HTTPListenPort  int
	HTTPSListenPort int

	RancherEmbedded bool
	RancherURL      string
	HCIMode         bool
}

// type Scaled struct {
// 	Ctx         context.Context
// 	VirtFactory *kubevirtv1.Factory
// }
