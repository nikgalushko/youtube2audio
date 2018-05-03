package config

import (
	"log"
	"path"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
)

type Config struct {
	lastIndex  uint64
	Converters Nodes
	Apis       Nodes
}

type ConfigReader struct {
	prefix    string
	lastIndex uint64
	client    *consulapi.Client
}

func NewConfigReader(addr, prefix string) (ConfigReader, error) {
	c := ConfigReader{prefix: prefix, lastIndex: 0}
	consulCfg := consulapi.DefaultConfig()
	consulCfg.Address = addr
	consulClient, err := consulapi.NewClient(consulCfg)
	if err != nil {
		log.Printf("[FATAL] starting error %s\n", err.Error())
		return c, err
	}

	c.client = consulClient
	return c, nil
}

func (cr *ConfigReader) Read() (Config, error) {
	cfg := Config{}

	qo := consulapi.QueryOptions{
		AllowStale: true,
		WaitIndex:  cr.lastIndex,
	}

	kvPairs, qm, err := cr.client.KV().List(cr.prefix, &qo)
	if err != nil {
		log.Printf("[FATAL] consul query error %s\n", err.Error())
		return cfg, err
	}

	cr.lastIndex = qm.LastIndex

	for _, item := range kvPairs {
		item.Key = strings.Replace(item.Key, cr.prefix, "", 1)
		dir, file := path.Split(item.Key)
		if file == "" {
			continue
		}

		node := Node{Adress: string(item.Value), Name: file}
		switch path.Base(dir) {
		case converter:
			cfg.Converters.Add(node)
		case api:
			cfg.Apis.Add(node)
		}
	}
	return cfg, err
}
