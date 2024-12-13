package config

import (
	"embed"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
)

type (
	Config struct {
		CustomerID string           `yaml:"customerID"`
		Git        GitConfig        `yaml:"git" validate:"required"`
		Cluster    ClusterConfig    `yaml:"cluster" validate:"required"`
		Forks      ForksConfig      `yaml:"forks" validate:"required"`
		Cloud      CloudConfig      `yaml:"cloud" validate:"required"`
		Monitoring MonitoringConfig `yaml:"monitoring" validate:"required"`
	}

	GitConfig struct {
		Username        string `yaml:"username"`
		Password        string `yaml:"password"`
		SSHPrivateKey   string `yaml:"sshPrivateKey"`
		UseSSHAgentAuth bool   `yaml:"useSSHAgentAuth"`
	}

	ForksConfig struct {
		KubeaidForkURL       string `yaml:"kubeaid" validate:"required"`
		KubeaidConfigForkURL string `yaml:"kubeaidConfig" validate:"required"`
	}

	ClusterConfig struct {
		ClusterName string `yaml:"name" validate:"required"`

		// NOTE : Currently, only Kubernetes v1.30.0 and v1.31.0 are supported.
		K8sVersion string `yaml:"k8sVersion" validate:"required"`
	}

	CloudConfig struct {
		AWS     *AWSConfig     `yaml:"aws"`
		Azure   *AzureConfig   `yaml:"azure"`
		Hetzner *HetznerConfig `yaml:"hetzner"`
	}

	AWSConfig struct {
		AccessKey    string `yaml:"accessKey" validate:"required"`
		SecretKey    string `yaml:"secretKey" validate:"required"`
		SessionToken string `yaml:"sessionToken"`
		Region       string `yaml:"region" validate:"required"`

		SSHKeyName string `yaml:"sshKeyName" validate:"required"`

		ControlPlaneInstanceType string `yaml:"controlPlaneInstanceType" validate:"required"`
		ControlPlaneAMI          string `yaml:"controlPlaneAMI" validate:"required"`
		ControlPlaneReplicas     int    `yaml:"controlPlaneReplicas" validate:"required"`

		NodeGroups []NodeGroups `yaml:"nodeGroups"`
	}

	NodeGroups struct {
		Name           string            `yaml:"name" validate:"required"`
		Replicas       int               `yaml:"replicas" validate:"required"`
		InstanceType   string            `yaml:"instanceType" validate:"required"`
		SSHKeyName     string            `yaml:"sshKeyName" validate:"required"`
		AMI            AMIConfig         `yaml:"ami" validate:"required"`
		RootVolumeSize int               `yaml:"rootVolumeSize" validate:"required"`
		Labels         map[string]string `yaml:"labels" default:"[]"`
		Taints         []v1.Taint        `yaml:"taints" default:"[]"`
	}

	AMIConfig struct {
		ID string `yaml:"id" validate:"required"`
	}

	AzureConfig struct{}

	HetznerConfig struct{}

	MonitoringConfig struct {
		KubePrometheusVersion string `yaml:"kubePrometheusVersion" validate:"required"`
		GrafanaURL            string `yaml:"grafanaURL"`
		ConnectObmondo        bool   `yaml:"connectObmondo"`
	}
)

//go:embed templates/*
var SampleConfigs embed.FS

func ParseConfigFile(configFile string) *Config {
	configFileContents, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed reading config file : %v", err)
	}
	parsedConfig, err := ParseConfig(string(configFileContents))
	if err != nil {
		log.Fatalf("Failed parsing config file : %v", err)
	}
	return parsedConfig
}

func ParseConfig(configAsString string) (*Config, error) {
	parsedConfig := &Config{}
	if err := yaml.Unmarshal([]byte(configAsString), parsedConfig); err != nil {
		return nil, fmt.Errorf("failed unmarshalling config : %v", err)
	}
	slog.Info("Parsed config")

	// Set defaults.
	if err := defaults.Set(parsedConfig); err != nil {
		log.Fatalf("Failed setting defaults for parsed config : %v", err)
	}

	validateConfig(parsedConfig)

	return parsedConfig, nil
}
