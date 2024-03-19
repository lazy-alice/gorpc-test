package consul

import (
	"alice_gorpc/plugin"
	"fmt"
	"github.com/hashicorp/consul/api"
	"net/http"
	"strconv"
)

type Consul struct {
	opts        *plugin.Options
	client      *api.Client
	config      *api.Config
	queryOption *api.QueryOptions
	writeOption *api.WriteOptions
}

var DefaultConsul = Consul{
	opts: &plugin.Options{},
}

func init() {
	plugin.Register("consul", DefaultConsul)
}

func (c *Consul) Init(opts ...plugin.Option) error {
	for _, o := range opts {
		o(c.opts)
	}

	if len(c.opts.Services) == 0 || c.opts.SvrAddr == "" || c.opts.SelectorSvrAddr == "" {
		return fmt.Errorf("consul init error, len(services) : %d, svrAddr : %s, selectorSvrAddr : %s",
			len(c.opts.Services), c.opts.SvrAddr, c.opts.SelectorSvrAddr)
	}

	if err := c.InitConfig(); err != nil {
		return err
	}

	for _, serviceName := range c.opts.Services {
		//生成对应的检查对象
		check := &api.AgentServiceCheck{
			TCP:                            c.opts.SvrAddr,
			Timeout:                        "5s",
			Interval:                       "5s",
			DeregisterCriticalServiceAfter: "10s",
		}

		//生成注册对象
		mp := make(map[string]string)
		mp["weight"] = ""
		registration := new(api.AgentServiceRegistration)
		registration.Name = serviceName
		registration.ID = "1"
		registration.Port = 49999
		registration.Tags = []string{"test"}
		registration.Address = "127.0.0.1"
		registration.Check = check
		registration.Meta = mp

		err := c.client.Agent().ServiceRegister(registration)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Consul) InitConfig() error {
	config := api.DefaultConfig()
	c.config = config

	config.HttpClient = http.DefaultClient
	config.Address = c.opts.SelectorSvrAddr
	config.Scheme = "http"

	client, err := api.NewClient(config)
	if err != nil {
		return err
	}
	c.client = client

	return nil
}

func (c *Consul) Resolve(serviceName string) ([]*Node, error) {
	var nodes []*Node
	catalogService, _, err := c.client.Catalog().Service(serviceName, "", c.queryOption)
	if err != nil {
		return nil, err
	}
	if len(catalogService) > 0 {
		for _, s := range catalogService {
			address := fmt.Sprintf("%s:%d", s.Address, s.ServicePort)
			weight, err := strconv.Atoi(s.NodeMeta["weight"])
			if err != nil {
				weight = 0
			}
			nodes = append(nodes, &Node{
				Key:    serviceName,
				Value:  address,
				Weight: weight,
			})
		}
	}
	return nodes, nil
}

func Init(consulSvrAddr string, opts ...plugin.Option) error {
	for _, o := range opts {
		o(DefaultConsul.opts)
	}

	DefaultConsul.opts.SelectorSvrAddr = consulSvrAddr
	err := DefaultConsul.InitConfig()
	return err
}
