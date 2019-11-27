package config

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/prometheus/common/log"
	yaml "gopkg.in/yaml.v2"
)

// Config is the Go representation of the yaml config file.
type Config struct {
	Databases map[string]DatabaseConfig `yaml:"databases"`
}

// SafeConfig wraps Config for concurrency-safe operations.
type SafeConfig struct {
	sync.RWMutex
	C *Config
}

// Credentials is the Go representation of the credentials section in the yaml
// config file.
type DatabaseConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"pass"`
}

func (sc *SafeConfig) ReloadConfig(configFile string) error {
	var c = &Config{}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Errorf("Error reading config file: %s", err)
		return err
	}

	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		log.Errorf("Error parsing config file: %s", err)
		return err
	}

	sc.Lock()
	sc.C = c
	sc.Unlock()

	log.Infoln("Loaded config file")
	return nil
}

// CredentialsForTarget returns the Credentials for a given target, or the
// default. It is concurrency-safe.
func (sc *SafeConfig) DatabaseConfigForTarget(target string) (DatabaseConfig, error) {
	sc.Lock()
	defer sc.Unlock()
	if databaseConfig, ok := sc.C.Databases[target]; ok {
		return DatabaseConfig{
			User:     databaseConfig.User,
			Password: databaseConfig.Password,
		}, nil
	}
	if databaseConfig, ok := sc.C.Databases["default"]; ok {
		return DatabaseConfig{
			User:     databaseConfig.User,
			Password: databaseConfig.Password,
		}, nil
	}
	return DatabaseConfig{}, fmt.Errorf("no credentials found for target %s", target)
}
