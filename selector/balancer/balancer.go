package balancer

import (
	"alice_gorpc/plugin/consul"
)

type Balancer interface {
	Balance(string, []*consul.Node) *consul.Node
}

var balancerMap = make(map[string]Balancer, 0)

var DefaultRandomBalancer = newRandomBalancer()

func GetBalancer(name string) Balancer {
	if balancer, ok := balancerMap[name]; ok {
		return balancer
	}
	return DefaultRandomBalancer
}
