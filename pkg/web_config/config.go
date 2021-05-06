package web_config

import (
	"fmt"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

var (
	volumeName = "web-config"
)

// Config is the web configuration for a prometheus instance.
type Config struct {
	webConfigSecretName string
}

// NewConfig creates a new Config.
func NewConfig(secretName string) *Config {
	return &Config{
		webConfigSecretName: secretName,
	}
}

// GenerateConfigFileContents generates a new web config a returns its contents.
// The format of the config file is available in the official prometheus documentation:
// https://prometheus.io/docs/prometheus/latest/configuration/https/#https-and-authentication
func (c Config) GenerateConfigFileContents(assetsPathPrefix string, tls *monitoringv1.WebTLSConfig) ([]byte, error) {
	if tls == nil {
		return yaml.Marshal(yaml.MapSlice{})
	}

	assets := NewTLSCredentials(assetsPathPrefix, tls.KeySecret, tls.Cert, tls.ClientCA)

	tlsServerConfig := yaml.MapSlice{}
	if certPath := assets.GetCertMountPath(); certPath != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{Key: "cert_file", Value: certPath})
	}

	if keyPath := assets.GetKeyMountPath(); keyPath != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{Key: "key_file", Value: keyPath})
	}

	if tls.ClientAuthType != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "client_auth_type",
			Value: tls.ClientAuthType,
		})
	}

	if caPath := assets.GetCAMountPath(); caPath != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{Key: "client_ca_file", Value: caPath})
	}

	if tls.MinVersion != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "min_version",
			Value: tls.MinVersion,
		})
	}

	if tls.MaxVersion != "" {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "max_version",
			Value: tls.MaxVersion,
		})
	}

	if len(tls.CipherSuites) != 0 {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "cipher_suites",
			Value: tls.CipherSuites,
		})
	}

	if tls.PreferServerCipherSuites != nil {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "prefer_server_cipher_suites",
			Value: tls.PreferServerCipherSuites,
		})
	}

	if len(tls.CurvePreferences) != 0 {
		tlsServerConfig = append(tlsServerConfig, yaml.MapItem{
			Key:   "curve_preferences",
			Value: tls.CurvePreferences,
		})
	}

	cfg := yaml.MapSlice{
		{
			Key:   "tls_server_config",
			Value: tlsServerConfig,
		},
	}

	return yaml.Marshal(cfg)
}

// Mount create a volumes and a volume mount referencing with the configuration file.
// In addition, Mount returns a web.config.file command line option referencing
// the file in the volume mount.
func (c Config) Mount(destinationPath string) (string, v1.Volume, v1.VolumeMount) {
	arg := c.makeArg(destinationPath)
	volume := c.makeVolume()
	mount := c.makeVolumeMount(destinationPath)

	return arg, volume, mount
}

func (c Config) makeArg(filePath string) string {
	return fmt.Sprintf("--web.config.file=%s", filePath)
}

func (c Config) makeVolume() v1.Volume {
	return v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: c.webConfigSecretName,
			},
		},
	}
}

func (c Config) makeVolumeMount(filePath string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      volumeName,
		SubPath:   "web-config.yaml",
		ReadOnly:  true,
		MountPath: filePath,
	}
}
