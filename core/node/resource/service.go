package resource

import "github.com/cskr/pubsub"

type Service struct {
	PubSub *pubsub.PubSub `optional:"true"`
}
