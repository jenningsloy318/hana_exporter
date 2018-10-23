package config

import (
	"fmt"
	"io/ioutil"
	"sync"

	"log"
	yaml "gopkg.in/yaml.v2"
)

// Config is the Go representation of the yaml config file.
type Config struct {
	Credentials  map[string]Credentials `yaml:"Credentials"`
	JaegerConfig JaegerConfig           `yaml:"Jaeger"`
}

// SafeConfig wraps Config for concurrency-safe operations.
type SafeConfig struct {
	sync.RWMutex
	C *Config
}

// Credentials is the Go representation of the credentials section in the yaml
// config file.
type Credentials struct {
	User     string `yaml:"user"`
	Password string `yaml:"pass"`
}

type JaegerConfig struct {
	AgentEndpointURI     string `yaml:"agentEndpointURI"`
	CollectorEndpointURI string `yaml:"collectorEndpointURI"`
}

func (sc *SafeConfig) ReloadConfig(configFile string) error {
	var c = &Config{}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {

		log.Printf("Error reading config file: %s", err)
		return err
	}

	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		log.Printf("Error parsing config file: %s", err)
		return err
	}

	sc.Lock()
	sc.C = c
	sc.Unlock()

	log. Println("Loaded config file")
	return nil
}

// CredentialsForTarget returns the Credentials for a given target, or the
// default. It is concurrency-safe.
func (sc *SafeConfig) CredentialsForTarget(target string) (Credentials, error) {
	sc.Lock()
	defer sc.Unlock()
	if credentials, ok := sc.C.Credentials[target]; ok {
		return Credentials{
			User:     credentials.User,
			Password: credentials.Password,
		}, nil
	}
	if credentials, ok := sc.C.Credentials["default"]; ok {
		return Credentials{
			User:     credentials.User,
			Password: credentials.Password,
		}, nil
	}
	return Credentials{}, fmt.Errorf("no credentials found for target %s", target)
}

func (sc *SafeConfig) ParseJaegerConfig() (JaegerConfig, error) {
	sc.Lock()
	defer sc.Unlock()
	jaegerConfig := sc.C.JaegerConfig
	return JaegerConfig{
		AgentEndpointURI:     jaegerConfig.AgentEndpointURI,
		CollectorEndpointURI: jaegerConfig.CollectorEndpointURI,
	}, nil

}
