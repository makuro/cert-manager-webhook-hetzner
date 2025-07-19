package main

import (
	"encoding/json"
	"github.com/cert-manager/cert-manager/test/acme"
	"gopkg.in/yaml.v3"
	"math/rand"
	"os"
	"testing"
)

var (
	zone               = os.Getenv("TEST_ZONE_NAME")
	fqdn               string
	configFile         = "testdata/hetzner/config.json"
	secretYamlFilePath = "testdata/hetzner/hetzner-secret.yml"
	secretName         = "hetzner-secret"
	apiKey             = os.Getenv("HCLOUD_DNS_API_TOKEN")
)

type SecretYaml struct {
	ApiVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind,omitempty" json:"kind,omitempty"`
	SecretType string `yaml:"type" json:"type"`
	Metadata   struct {
		Name string `yaml:"name"`
	}
	Data struct {
		ApiKey string `yaml:"api-key"`
	}
}

func TestRunsSuite(t *testing.T) {

	slogger := zapLogger.Sugar()

	secretYaml := SecretYaml{}

	secretYaml.ApiVersion = "v1"
	secretYaml.Kind = "Secret"
	secretYaml.SecretType = "Opaque"
	secretYaml.Metadata.Name = secretName
	secretYaml.Data.ApiKey = apiKey

	secretYamlFile, err := yaml.Marshal(&secretYaml)
	if err != nil {
		slogger.Error(err)
	}
	_ = os.WriteFile(secretYamlFilePath, secretYamlFile, 0644)

	providerConfig := hetznerDNSProviderConfig{
		secretName,
		zone,
		"https://dns.hetzner.com/api/v1",
	}
	file, _ := json.MarshalIndent(providerConfig, "", " ")
	_ = os.WriteFile(configFile, file, 0644)

	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	fqdn = GetRandomString(8) + "." + zone

	fixture := dns.NewFixture(&hetznerDNSProviderSolver{},
		dns.SetResolvedZone(zone),
		dns.SetResolvedFQDN(fqdn),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/hetzner"),
		dns.SetStrict(false),
	)

	fixture.RunConformance(t)

	_ = os.Remove(configFile)
	_ = os.Remove(secretYamlFilePath)

}

func GetRandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
