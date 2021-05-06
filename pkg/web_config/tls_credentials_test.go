package web_config_test

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/web_config"
	v1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestMountVolumes(t *testing.T) {
	ts := []struct {
		keySecret       v1.SecretKeySelector
		cert            monitoringv1.SecretOrConfigMap
		clientCA        monitoringv1.SecretOrConfigMap
		expectedVolumes []v1.Volume
		expectedMounts  []v1.VolumeMount
	}{
		{
			clientCA: monitoringv1.SecretOrConfigMap{
				Secret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "some-secret",
					},
					Key: "tls.client_ca",
				},
			},
			cert: monitoringv1.SecretOrConfigMap{
				Secret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "some-secret",
					},
					Key: "tls.crt",
				},
			},
			keySecret: v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "some-secret",
				},
				Key: "tls.key",
			},
			expectedVolumes: []v1.Volume{
				{
					Name: "web-config-tls-secret-key-some-secret",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
				{
					Name: "web-config-tls-secret-cert-some-secret",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
				{
					Name: "web-config-tls-secret-client-ca-some-secret",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "some-secret",
						},
					},
				},
			},
			expectedMounts: []v1.VolumeMount{
				{
					Name:             "web-config-tls-secret-key-some-secret",
					ReadOnly:         true,
					MountPath:        "/tls-assets-prefix/secret__some-secret_tls.key",
					SubPath:          "tls.key",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "web-config-tls-secret-cert-some-secret",
					ReadOnly:         true,
					MountPath:        "/tls-assets-prefix/secret__some-secret_tls.crt",
					SubPath:          "tls.crt",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				{
					Name:             "web-config-tls-secret-client-ca-some-secret",
					ReadOnly:         true,
					MountPath:        "/tls-assets-prefix/secret__some-secret_tls.client_ca",
					SubPath:          "tls.client_ca",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			},
		},
	}

	for _, tt := range ts {
		tlsAssets := web_config.NewTLSCredentials("/tls-assets-prefix", tt.keySecret, tt.cert, tt.clientCA)
		volumes, mounts := tlsAssets.Mount()

		if !reflect.DeepEqual(volumes, tt.expectedVolumes) {
			t.Errorf("invalid volumes,\ngot  %v,\nwant %v", volumes, tt.expectedVolumes)
		}

		if !reflect.DeepEqual(mounts, tt.expectedMounts) {
			t.Errorf("invalid mounts,\ngot  %v,\nwant %v", mounts, tt.expectedMounts)
		}
	}
}

func TestMount(t *testing.T) {
	ts := []struct {
		keySecret        v1.SecretKeySelector
		cert             monitoringv1.SecretOrConfigMap
		clientCA         monitoringv1.SecretOrConfigMap
		expectedKeyPath  string
		expectedCertPath string
		expectedCAPath   string
	}{
		{
			keySecret: v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "some-secret",
				},
				Key: "tls.key",
			},
			cert: monitoringv1.SecretOrConfigMap{
				Secret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "some-secret",
					},
					Key: "tls.crt",
				},
			},
			clientCA: monitoringv1.SecretOrConfigMap{
				Secret: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "some-secret",
					},
					Key: "tls.client_ca",
				},
			},
			expectedKeyPath:  "/tls-assets-prefix/secret__some-secret_tls.key",
			expectedCertPath: "/tls-assets-prefix/secret__some-secret_tls.crt",
			expectedCAPath:   "/tls-assets-prefix/secret__some-secret_tls.client_ca",
		},
	}

	for _, tt := range ts {
		tlsAssets := web_config.NewTLSCredentials("/tls-assets-prefix", tt.keySecret, tt.cert, tt.clientCA)

		keyPath := tlsAssets.GetKeyMountPath()
		if keyPath != tt.expectedKeyPath {
			t.Errorf("invalid keySecret path, got %s, want %s", keyPath, tt.expectedKeyPath)
		}

		certPath := tlsAssets.GetCertMountPath()
		if certPath != tt.expectedCertPath {
			t.Errorf("invalid cert path, got %s, want %s", certPath, tt.expectedCertPath)
		}

		caPath := tlsAssets.GetCAMountPath()
		if caPath != tt.expectedCAPath {
			t.Errorf("invalid cert  path, got %s, want %s", caPath, tt.expectedCAPath)
		}
	}
}
