package selector

import (
	"alice_gorpc/plugin/consul"
	balancer2 "alice_gorpc/selector/balancer"
	"fmt"
)

type Selector struct {
	balancerName string
}

var selectorMap = make(map[string]*Selector)

var DefaultSelector = &Selector{}

func GetSelector(name string) *Selector {
	if selector, ok := selectorMap[name]; ok {
		return selector
	}
	return DefaultSelector
}

func (s *Selector) Select(serviceName string) (string, error) {
	nodes, err := consul.DefaultConsul.Resolve(serviceName)
	if err != nil {
		return "", err
	}
	balancer := balancer2.GetBalancer(s.balancerName)
	node := balancer.Balance(serviceName, nodes)
	if node == nil {
		return "", fmt.Errorf("no services find in %s", serviceName)
	}
	return node.Value, nil
}
