package web_config

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"path"
)

var (
	volumePrefix = "web-config-tls-"
)

// TLSCredentials are the credentials used for web TLS.
type TLSCredentials struct {
	// mountPath is the directory where TLS credentials are intended to be mounted
	mountPath string

	// keySecret is the kubernetes secret containing the TLS key
	keySecret v1.SecretKeySelector
	// cert is the kubernetes secret or configmap containing the TLS certificate
	cert monitoringv1.SecretOrConfigMap
	// clientCA is the kubernetes secret or configmap containing the client CA certificate
	clientCA monitoringv1.SecretOrConfigMap
}

// NewTLSCredentials creates new TLSCredentials from secrets of configmaps.
func NewTLSCredentials(
	mountPath string,
	key v1.SecretKeySelector,
	cert monitoringv1.SecretOrConfigMap,
	clientCA monitoringv1.SecretOrConfigMap,
) *TLSCredentials {
	return &TLSCredentials{
		mountPath: mountPath,
		keySecret: key,
		cert:      cert,
		clientCA:  clientCA,
	}
}

// GetKeyMountPath is the mount path of the TLS key inside a prometheus container.
func (a *TLSCredentials) GetKeyMountPath() string {
	secret := monitoringv1.SecretOrConfigMap{Secret: &a.keySecret}
	return a.assetPath(secret)
}

// GetCertMountPath is the mount path of the TLS certificate inside a prometheus container,
func (a *TLSCredentials) GetCertMountPath() string {
	if a.cert.ConfigMap != nil || a.cert.Secret != nil {
		return a.assetPath(a.cert)
	}

	return ""
}

// GetCAMountPath is the mount path of the client CA certificate inside a prometheus container.
func (a *TLSCredentials) GetCAMountPath() string {
	if a.clientCA.ConfigMap != nil || a.clientCA.Secret != nil {
		return a.assetPath(a.clientCA)
	}

	return ""
}

func (a *TLSCredentials) assetPath(sel monitoringv1.SecretOrConfigMap) string {
	return path.Join(a.mountPath, assets.TLSAssetKeyFromSelector("", sel).String())
}

// Mount creates volumes and volume mounts referencing the TLS credentials.
func (a *TLSCredentials) Mount() ([]v1.Volume, []v1.VolumeMount) {
	var volumes []v1.Volume
	var mounts []v1.VolumeMount

	prefix := volumePrefix + "secret-key-"
	volumes, mounts = a.mountSecret(volumes, mounts, &a.keySecret, prefix, a.GetKeyMountPath())

	if a.cert.Secret != nil {
		prefix := volumePrefix + "secret-cert-"
		volumes, mounts = a.mountSecret(volumes, mounts, a.cert.Secret, prefix, a.GetCertMountPath())
	} else if a.cert.ConfigMap != nil {
		prefix := volumePrefix + "configmap-cert-"
		volumes, mounts = a.mountConfigMap(volumes, mounts, a.cert.ConfigMap, prefix, a.GetCertMountPath())
	}

	if a.clientCA.Secret != nil {
		prefix := volumePrefix + "secret-client-ca-"
		volumes, mounts = a.mountSecret(volumes, mounts, a.clientCA.Secret, prefix, a.GetCAMountPath())
	} else if a.clientCA.ConfigMap != nil {
		prefix := volumePrefix + "configmap-client-ca-"
		volumes, mounts = a.mountConfigMap(volumes, mounts, a.clientCA.ConfigMap, prefix, a.GetCAMountPath())
	}

	return volumes, mounts
}

func (a *TLSCredentials) mountSecret(
	volumes []v1.Volume,
	mounts []v1.VolumeMount,
	secret *corev1.SecretKeySelector,
	volumePrefix string,
	mountPath string,
) ([]v1.Volume, []v1.VolumeMount) {
	volumeName := volumePrefix + secret.Name
	volumes = append(volumes, v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secret.Name,
			},
		},
	})

	mounts = append(mounts, v1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: mountPath,
		SubPath:   secret.Key,
	})

	return volumes, mounts
}

func (a *TLSCredentials) mountConfigMap(
	volumes []v1.Volume,
	mounts []v1.VolumeMount,
	configMap *corev1.ConfigMapKeySelector,
	volumePrefix string,
	mountPath string,
) ([]v1.Volume, []v1.VolumeMount) {
	volumeName := volumePrefix + configMap.Name
	volumes = append(volumes, v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMap.Name,
				},
			},
		},
	})

	mounts = append(mounts, v1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: mountPath,
		SubPath:   configMap.Key,
	})

	return volumes, mounts
}
