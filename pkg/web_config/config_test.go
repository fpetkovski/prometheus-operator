package web_config_test

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/web_config"
	v1 "k8s.io/api/core/v1"
	"testing"
)

var falseVal = false

func TestGenerateWebConfig(t *testing.T) {
	tc := []struct {
		name         string
		webTlsConfig *monitoringv1.WebTLSConfig
		expected     string
	}{
		{
			name: "minimal TLS config with certificate from secret",
			webTlsConfig: &monitoringv1.WebTLSConfig{
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.crt",
					},
				},
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "tls.key",
				},
			},
			expected:
			`tls_server_config:
  cert_file: /web_certs_path_prefix/secret__test-secret_tls.crt
  key_file: /web_certs_path_prefix/secret__test-secret_tls.key
`,
		},
		{
			name: "minimal TLS config with certificate from configmap",
			webTlsConfig: &monitoringv1.WebTLSConfig{
				Cert: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-configmap",
						},
						Key: "tls.crt",
					},
				},
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "tls.key",
				},
			},
			expected:
			`tls_server_config:
  cert_file: /web_certs_path_prefix/configmap__test-configmap_tls.crt
  key_file: /web_certs_path_prefix/secret__test-secret_tls.key
`,
		},
		{
			name: "minimal TLS config with client CA from configmap",
			webTlsConfig: &monitoringv1.WebTLSConfig{
				Cert: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-configmap",
						},
						Key: "tls.crt",
					},
				},
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "tls.key",
				},
				ClientCA: monitoringv1.SecretOrConfigMap{
					ConfigMap: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-configmap",
						},
						Key: "tls.client_ca",
					},
				},
			},
			expected:
			`tls_server_config:
  cert_file: /web_certs_path_prefix/configmap__test-configmap_tls.crt
  key_file: /web_certs_path_prefix/secret__test-secret_tls.key
  client_ca_file: /web_certs_path_prefix/configmap__test-configmap_tls.client_ca
`,
		},
		{
			name: "TLS config with all parameters from secrets",
			webTlsConfig: &monitoringv1.WebTLSConfig{
				ClientCA: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.ca",
					},
				},
				Cert: monitoringv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "tls.crt",
					},
				},
				KeySecret: v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "tls.keySecret",
				},
				ClientAuthType:           "RequireAnyClientCert",
				MinVersion:               "TLS11",
				MaxVersion:               "TLS13",
				CipherSuites:             []string{"cipher-1", "cipher-2"},
				PreferServerCipherSuites: &falseVal,
				CurvePreferences:         []string{"curve-1", "curve-2"},
			},
			expected:
`tls_server_config:
  cert_file: /web_certs_path_prefix/secret__test-secret_tls.crt
  key_file: /web_certs_path_prefix/secret__test-secret_tls.keySecret
  client_auth_type: RequireAnyClientCert
  client_ca_file: /web_certs_path_prefix/secret__test-secret_tls.ca
  min_version: TLS11
  max_version: TLS13
  cipher_suites:
  - cipher-1
  - cipher-2
  prefer_server_cipher_suites: false
  curve_preferences:
  - curve-1
  - curve-2
`,
		},
	}

	for _, tt := range tc {
		config := web_config.NewConfig("test-secret")
		actual, err := config.GenerateConfigFileContents("/web_certs_path_prefix", tt.webTlsConfig)
		if err != nil {
			t.Fatal(err)
		}

		if tt.expected != string(actual) {
			t.Fatalf("%s failed.\n\nGot %s\nwant %s\n", tt.name, actual, tt.expected)
		}
	}
}
