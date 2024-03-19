package balancer

import (
	"alice_gorpc/plugin/consul"
	"math/rand"
	"time"
)

type randomBalancer struct{}

func newRandomBalancer() *randomBalancer {
	return &randomBalancer{}
}

func (r *randomBalancer) Balance(serviceName string, nodes []*consul.Node) *consul.Node {
	if len(nodes) == 0 {
		return nil
	}
	rand.Seed(time.Now().Unix())
	num := rand.Intn(len(nodes))
	return nodes[num]
}
