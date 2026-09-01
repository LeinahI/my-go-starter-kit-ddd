package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultSeederConfigPath = "configs/seeder.yaml"

type SeederConfig struct {
	DemoProducts []SeederDemoProduct `yaml:"demo_products"`
}

type SeederDemoProduct struct {
	Name        string `yaml:"name"`
	Slug        string `yaml:"slug"`
	Description string `yaml:"description"`
	Price       string `yaml:"price"`
	Stock       int    `yaml:"stock"`
}

func LoadSeederConfig(path string) (*SeederConfig, error) {
	if path == "" {
		path = DefaultSeederConfigPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read seeder config %q: %w", path, err)
	}

	var cfg SeederConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse seeder config %q: %w", path, err)
	}
	if len(cfg.DemoProducts) == 0 {
		return nil, fmt.Errorf("seeder config %q: demo_products is required", path)
	}

	for i, p := range cfg.DemoProducts {
		if p.Name == "" {
			return nil, fmt.Errorf("seeder config %q: demo_products[%d].name is required", path, i)
		}
		if p.Slug == "" {
			return nil, fmt.Errorf("seeder config %q: demo_products[%d].slug is required", path, i)
		}
		if p.Price == "" {
			return nil, fmt.Errorf("seeder config %q: demo_products[%d].price is required", path, i)
		}
	}

	return &cfg, nil
}
