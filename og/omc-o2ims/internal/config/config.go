package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

type CRDConfig struct {
	Load bool
	//Update  bool
	//Watcher bool
}

// Config represents the configuration data.
type Config struct {
	ServerPort    int    `yaml:"server_port"`
	DataStoreType string `yaml:"data_store_type"`
	Kubernetes    struct {
		Namespace  string `yaml:"namespace"`
		KubeConfig string `yaml:"kubeconfig"`
	} `yaml:"kubernetes,omitempty"`
	CRD struct {
		Files []string `yaml:"files"`
		Load  bool     `yaml:"load"`
	} `yaml:"crd,omitempty"`
	Logging struct {
		Level    string `yaml:"level"`
		Filename string `yaml:"filename"`
	} `yaml:"logging"`

	DataStore string `yaml:"data_store"`
	//k8s, json, opensearch

	BackendType string `yaml:"backend_type"`
	//omc_rest_v1 , omc_rest_simulator
	Omc struct {
		URL      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"omc,omitempty"`
}

type ConfigInterface interface {
	LoadConfig(configFile string) (*Config, error)
	Validate() error
}

// LoadConfig loads and validates the application configuration.
func LoadConfig(configFile string) (*Config, error) {
	if configFile == "" {
		configFile = findConfigFile()
	}

	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	err = config.Validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

var osStatFunc = os.Stat // Store original function

func resetOsStat() {
	osStatFunc = os.Stat
}

// findConfigFile searches for configuration file in the default directories.
func findConfigFile() string {
	defaultDirs := []string{"/etc/", "/etc/omc-o2ims/", "$HOME/.omc-o2ims/", "/app/", "./"}
	possibleConfigFiles := []string{"config.yaml", "omc-o2ims.yaml", ".omc-o2ims.yaml"}

	for _, dir := range defaultDirs {
		for _, file := range possibleConfigFiles {
			configPath := dir + file

			_, err := osStatFunc(configPath)
			if err == nil {
				return configPath
			}
		}
	}
	return ""
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Perform validation on the configuration
	// Add more validation checks here for other required fields
	// Example:
	if c.ServerPort == 0 {
		return errors.New("server_port is required")
	}

	return nil
}
