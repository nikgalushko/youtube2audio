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

func New(consulAddr string) Config {
	return Config{}
}

func (c Config) Reload() (Config, error) {
	cfg := Config{}
	consulCfg := consulapi.DefaultConfig()
	consulCfg.Address = "85.90.244.67:8500"
	consulClient, err := consulapi.NewClient(consulCfg)
	if err != nil {
		log.Printf("[FATAL] starting error %s\n", err.Error())
		return cfg, err
	}

	qo := consulapi.QueryOptions{
		AllowStale: true,
		WaitIndex:  c.lastIndex,
	}

	kvPairs, _, err := consulClient.KV().List("test", &qo)
	if err != nil {
		log.Printf("[FATAL] consul query error %s\n", err.Error())
		return cfg, err
	}

	for _, item := range kvPairs {
		item.Key = strings.Replace(item.Key, "test", "", 1)
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
