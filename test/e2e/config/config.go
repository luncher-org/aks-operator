/*
Copyright © 2022 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/drone/envsubst/v2"
	"sigs.k8s.io/yaml"
)

type E2EConfig struct {
	OperatorChart string `yaml:"operatorChart"`
	CRDChart      string `yaml:"crdChart"`
	ExternalIP    string `yaml:"externalIP"`
	MagicDNS      string `yaml:"magicDNS"`
	BridgeIP      string `yaml:"bridgeIP"`
	ArtifactsDir  string `yaml:"artifactsDir"`

	CertManagerVersion  string `yaml:"certManagerVersion"`
	CertManagerChartURL string `yaml:"certManagerChartURL"`

	RancherVersion  string `yaml:"rancherVersion"`
	RancherChartURL string `yaml:"rancherChartURL"`

	AzureClientID       string `yaml:"azureClientID"`
	AzureClientSecret   string `yaml:"azureClientSecret"`
	AzureSubscriptionID string `yaml:"azureSubscriptionID"`

	AzureResourceGroup string `yaml:"resourceGroupPrefix"`
}

// ReadE2EConfig read config from yaml and substitute variables using envsubst.
// All variables can be overridden by environmental variables.
func ReadE2EConfig(configPath string) (*E2EConfig, error) { //nolint:gocyclo
	config := &E2EConfig{}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if configData == nil {
		return nil, errors.New("config file can't be empty")
	}

	if err := yaml.Unmarshal(configData, config); err != nil {
		return nil, fmt.Errorf("failed to unmarhal config file: %s", err)
	}

	if operatorChart := os.Getenv("OPERATOR_CHART"); operatorChart != "" {
		config.OperatorChart = operatorChart
	}

	if config.OperatorChart == "" {
		return nil, errors.New("no OPERATOR_CHART provided, an operator helm chart is required to run e2e tests")
	}

	if crdChart := os.Getenv("CRD_CHART"); crdChart != "" {
		config.CRDChart = crdChart
	}

	if config.CRDChart == "" {
		return nil, errors.New("no CRD_CHART provided, a crd helm chart is required to run e2e tests")
	}

	if externalIP := os.Getenv("EXTERNAL_IP"); externalIP != "" {
		config.ExternalIP = externalIP
	}

	if config.ExternalIP == "" {
		return nil, errors.New("no EXTERNAL_IP provided, a known (reachable) node external ip it is required to run e2e tests")
	}

	if magicDNS := os.Getenv("MAGIC_DNS"); magicDNS != "" {
		config.MagicDNS = magicDNS
	}

	if bridgeIP := os.Getenv("BRIDGE_IP"); bridgeIP != "" {
		config.BridgeIP = bridgeIP
	}

	if artifactsDir := os.Getenv("ARTIFACTS_DIR"); artifactsDir != "" {
		config.ArtifactsDir = artifactsDir
	}

	if azureClientID := os.Getenv("AZURE_CLIENT_ID"); azureClientID != "" {
		config.AzureClientID = azureClientID
	}

	if azureClientSecret := os.Getenv("AZURE_CLIENT_SECRET"); azureClientSecret != "" {
		config.AzureClientSecret = azureClientSecret
	}

	if azureSubscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID"); azureSubscriptionID != "" {
		config.AzureSubscriptionID = azureSubscriptionID
	}

	if azureResourceGroup := os.Getenv("AZURE_RESOURCE_GROUP"); azureResourceGroup != "" {
		config.AzureResourceGroup = azureResourceGroup
	}

	if certManagerVersion := os.Getenv("CERT_MANAGER_VERSION"); certManagerVersion != "" {
		config.CertManagerVersion = certManagerVersion
	}

	if certManagerURL := os.Getenv("CERT_MANAGER_CHART_URL"); certManagerURL != "" {
		config.CertManagerChartURL = certManagerURL
	}

	if rancherVersion := os.Getenv("RANCHER_VERSION"); rancherVersion != "" {
		config.RancherVersion = rancherVersion
	}

	if rancherURL := os.Getenv("RANCHER_CHART_URL"); rancherURL != "" {
		config.RancherChartURL = rancherURL
	}

	if err := substituteVersions(config); err != nil {
		return nil, err
	}

	return config, validateAzureCredentials(config)
}

func substituteVersions(config *E2EConfig) error {
	certManagerURL, err := envsubst.Eval(config.CertManagerChartURL, func(_ string) string {
		return config.CertManagerVersion
	})
	if err != nil {
		return fmt.Errorf("failed to substitute cert manager chart url: %w", err)
	}
	config.CertManagerChartURL = certManagerURL

	rancherURL, err := envsubst.Eval(config.RancherChartURL, func(_ string) string {
		return config.RancherVersion
	})
	if err != nil {
		return fmt.Errorf("failed to substitute rancher chart url: %w", err)
	}
	config.RancherChartURL = rancherURL

	return nil
}

func validateAzureCredentials(config *E2EConfig) error {
	if config.AzureClientID == "" {
		return errors.New("no AZURE_CLIENT_ID provided, an azure client id is required to run e2e tests")
	}

	if config.AzureClientSecret == "" {
		return errors.New("no AZURE_CLIENT_SECRET provided, an azure client secret is required to run e2e tests")
	}

	if config.AzureSubscriptionID == "" {
		return errors.New("no AZURE_SUBSCRIPTION_ID provided, an azure subscription id is required to run e2e tests")
	}

	return nil
}
