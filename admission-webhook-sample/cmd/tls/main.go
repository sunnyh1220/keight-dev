package main

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sunnyh1220/keight-dev/admission-webhook-sample/pkg/file"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"math/big"
	"os"
	"time"
)

func main() {
	var caPEM, serverCertPEM, serverPrivKeyPEM *bytes.Buffer
	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Province:           []string{"Beijing"},
			Locality:           []string{"Beijing"},
			OrganizationalUnit: []string{"easy.sunnyh"},
			Organization:       []string{"easy.sunnyh"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true, // 根证书
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode CA cert
	caPEM = new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	dnsNames := []string{"admission-webhook-sample",
		"admission-webhook-sample.default", "admission-webhook-sample.default.svc"}
	commonName := "admission-webhook-sample.default.svc"

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"easy.sunnyh"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode the server cert and key
	serverCertPEM = new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	err = os.MkdirAll("/etc/webhook/certs/", 0666)
	if err != nil {
		log.Panic(err)
	}
	err = file.WriteFile("/etc/webhook/certs/tls.crt", serverCertPEM)
	if err != nil {
		log.Panic(err)
	}

	err = file.WriteFile("/etc/webhook/certs/tls.key", serverPrivKeyPEM)
	if err != nil {
		log.Panic(err)
	}

	log.Println("webhook server tls generated successfully")

	if err := CreateAdmissionConfig(caPEM); err != nil {
		log.Panic(err)
	}

	log.Println("webhook admission configuration object generated successfully")

}

func initKubeClient() (*kubernetes.Clientset, error) {
	var (
		err    error
		config *rest.Config
	)
	if config, err = rest.InClusterConfig(); err != nil {
		return nil, err
	}

	// 创建 Clientset 对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func CreateAdmissionConfig(caCert *bytes.Buffer) error {
	var (
		webhookNamespace, _ = os.LookupEnv("WEBHOOK_NAMESPACE")
		mutationCfgName, _  = os.LookupEnv("MUTATE_CONFIG")
		validateCfgName, _  = os.LookupEnv("VALIDATE_CONFIG")
		webhookService, _   = os.LookupEnv("WEBHOOK_SERVICE")
		validatePath, _     = os.LookupEnv("VALIDATE_PATH")
		mutationPath, _     = os.LookupEnv("MUTATE_PATH")
	)

	clientset, err := initKubeClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	if validateCfgName != "" {
		validateConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: validateCfgName,
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "easy.sunnyh.admission-webhook-sample",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						CABundle: caCert.Bytes(),
						Service: &admissionregistrationv1.ServiceReference{
							Name:      webhookService,
							Namespace: webhookNamespace,
							Path:      &validatePath,
						},
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{""},
								APIVersions: []string{"v1"},
								Resources:   []string{"pods"},
							},
						},
					},
					FailurePolicy: func() *admissionregistrationv1.FailurePolicyType {
						pt := admissionregistrationv1.Fail
						return &pt
					}(),
					AdmissionReviewVersions: []string{"v1"},
					SideEffects: func() *admissionregistrationv1.SideEffectClass {
						se := admissionregistrationv1.SideEffectClassNone
						return &se
					}(),
				},
			},
		}

		validateAdmissionClient := clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations()
		_, err := validateAdmissionClient.Get(ctx, validateCfgName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				if _, err = validateAdmissionClient.Create(ctx, validateConfig, metav1.CreateOptions{}); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if _, err = validateAdmissionClient.Update(ctx, validateConfig, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}

	}

	if mutationCfgName != "" {
		mutateConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: mutationCfgName,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{{
				Name: "easy.sunnyh.admission-webhook-sample-mutate",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caCert.Bytes(), // CA bundle created earlier
					Service: &admissionregistrationv1.ServiceReference{
						Name:      webhookService,
						Namespace: webhookNamespace,
						Path:      &mutationPath,
					},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{{Operations: []admissionregistrationv1.OperationType{
					admissionregistrationv1.Create},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{"apps", ""},
						APIVersions: []string{"v1"},
						Resources:   []string{"deployments", "services"},
					},
				}},
				FailurePolicy: func() *admissionregistrationv1.FailurePolicyType {
					pt := admissionregistrationv1.Fail
					return &pt
				}(),
				AdmissionReviewVersions: []string{"v1"},
				SideEffects: func() *admissionregistrationv1.SideEffectClass {
					se := admissionregistrationv1.SideEffectClassNone
					return &se
				}(),
			}},
		}

		mutateAdmissionClient := clientset.AdmissionregistrationV1().MutatingWebhookConfigurations()
		_, err := mutateAdmissionClient.Get(ctx, mutationCfgName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				if _, err = mutateAdmissionClient.Create(ctx, mutateConfig, metav1.CreateOptions{}); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if _, err = mutateAdmissionClient.Update(ctx, mutateConfig, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}

	}

	return nil
}
