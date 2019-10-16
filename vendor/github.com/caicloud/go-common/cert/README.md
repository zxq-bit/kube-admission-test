# cert

## Usage

### GenSelfSignedCertForK8sService

为一个 Kubernetes service 生成自签名证书。

首先根据 service namespace 和 name 生成证书：

```go
serverCert, serverKey, err := GenSelfSignedCertForK8sService(namespace, name)
if err != nil {
	// do something
}
```

`serverCert` 除了用在 web server 的 TLS config 外，还要用到 Kubernetes 的 webhook config 里面，传给 `CABundle` 让 Kubernetes apiserver 使用这个 cert 来认证自签的证书

```go
	// `caBundle` is a PEM encoded CA bundle which will be used to validate the webhook's server certificate.
	// If unspecified, system trust roots on the apiserver are used.
	// +optional
	CABundle []byte `json:"caBundle,omitempty" protobuf:"bytes,2,opt,name=caBundle"`
```

```go
	admrv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Webhooks: []admrv1.Webhook{
			{
				Name:  name + DefaultExternalAdmissionHookNameReign,
				Rules: rules,
				ClientConfig: admrv1.WebhookClientConfig{
					Service: &admrv1.ServiceReference{
						Name:      name,
						Namespace: namespace,
					},
					CABundle: serverCert,
				},
				FailurePolicy: &failurePolicy,
			},
		},
	}
```
