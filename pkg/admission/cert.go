package admission

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/zoumo/golib/cert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/caicloud/clientset/kubernetes"
	"github.com/caicloud/workload/pkg/type/v1alpha1"
)

const (
	defaultTLSSecretName = "workload-admission-tls"
)

type tlsProvider struct {
	client     kubernetes.Interface
	secretName string
}

func newTLSProvider(client kubernetes.Interface) *tlsProvider {
	tlsSecretName := os.Getenv("TLS_SECRET")
	if tlsSecretName == "" {
		tlsSecretName = defaultTLSSecretName
	}
	return &tlsProvider{
		client:     client,
		secretName: tlsSecretName,
	}
}

// retrieveTLSConfig retrieves the TLS config
func (p *tlsProvider) retrieveTLSConfig() (*tls.Config, []byte, error) {
	keyPEM, certPEM, err := p.ensureKeyAndCert()
	if err != nil {
		return nil, nil, err
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, certPEM, nil
}

// ensureKeyAndCert tries to get key and cert from secret firstly.
// Otherwise, it generates a new self-signed certificate and then stores the self-signed certificate in secret.
// The key and cert in secret for future use.
func (p *tlsProvider) ensureKeyAndCert() (key, cert []byte, err error) {
	// try to get tls cert from secret
	secret, err := p.client.CoreV1().Secrets(metav1.NamespaceSystem).Get(p.secretName, metav1.GetOptions{})
	// successfully get secret
	if err == nil {
		// get key and cert from secret
		klog.Infof("Restore key and cert from secret %v/%v", metav1.NamespaceSystem, p.secretName)
		certPEM := secret.Data[corev1.TLSCertKey]
		keyPEM := secret.Data[corev1.TLSPrivateKeyKey]
		return keyPEM, certPEM, nil
	}
	// occur error
	if !errors.IsNotFound(err) {
		return nil, nil, err
	}

	// generate private key
	keyPEMBlock, cerPEMBlock, err := p.genKeyAndCert()
	if err != nil {
		return nil, nil, err
	}

	if err := p.storeKeyAndCertInSecret(p.secretName, keyPEMBlock.Raw, cerPEMBlock.Raw); err != nil {
		return nil, nil, err
	}

	return keyPEMBlock.Raw, cerPEMBlock.Raw, nil
}

// genKeyAndCert generates private key and certificate
func (p *tlsProvider) genKeyAndCert() (*cert.PEM, *cert.PEM, error) {
	// generate private key
	key, err := cert.NewRSAPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := cert.NewSelfSignedCertificateBytes(
		cert.Options{
			CommonName:   fmt.Sprintf("%v.%v.svc", v1alpha1.WorkloadAdmissionServiceName, metav1.NamespaceSystem),
			Organization: []string{"caicloud.io"},
			DNSNames: []string{
				v1alpha1.WorkloadAdmissionServiceName,
				fmt.Sprintf("%v.%v", v1alpha1.WorkloadAdmissionServiceName, metav1.NamespaceSystem),
				fmt.Sprintf("%v.%v.svc", v1alpha1.WorkloadAdmissionServiceName, metav1.NamespaceSystem),
			},
		},
		key,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate certificate request, %v", err)
	}
	return cert.NewPEMForRSAKey(key), cert.NewPEMForCertificate(certBytes), nil
}

// storeKeyAndCertInSecret stores key and certificate ...
func (p *tlsProvider) storeKeyAndCertInSecret(secretName string, keyBytes, certBytes []byte) error {
	klog.Infof("Store key and cert in secret %v/%v", metav1.NamespaceSystem, secretName)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSPrivateKeyKey: keyBytes,
			corev1.TLSCertKey:       certBytes,
		},
	}

	_, err := p.client.CoreV1().Secrets(metav1.NamespaceSystem).Create(secret)
	return err
}
