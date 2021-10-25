
查看kube-apiserver是否启用Admission Webhook
如果没有MutatingAdmissionWebhook,ValidatingAdmissionWebhook两个参数，添加后重启apiserver
```bash
# minikube
$ kubectl get pods kube-apiserver-minikube -n kube-system -o yaml | grep enable-admission-plugins
    - --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota

```

检查是否注册admission webhook的api
```bash
$ kubectl api-versions |grep admission
admissionregistration.k8s.io/v1
admissionregistration.k8s.io/v1beta1
```

Admission Review
https://github.com/kubernetes/api/blob/master/admission/v1/types.go#L29


TLS
```bash
# install cfssl
$ wget -q --show-progress --https-only --timestamping \
  https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 \
  https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
$ chmod +x cfssl_linux-amd64 cfssljson_linux-amd64
$ sudo mv cfssl_linux-amd64 /usr/local/bin/cfssl
$ sudo mv cfssljson_linux-amd64 /usr/local/bin/cfssljson
```

创建CA证书机构
```bash
$ mkdir cert && cd cert

$ cat > ca-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "8760h"
    },
    "profiles": {
      "server": {
        "usages": ["signing", "key encipherment", "server auth", "client auth"],
        "expiry": "8760h"
      }
    }
  }
}
EOF

$ cat > ca-csr.json <<EOF
{
    "CN": "kubernetes",
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "names": [
        {
            "C": "CN",
            "L": "BeiJing",
            "ST": "BeiJing",
            "O": "k8s",
            "OU": "System"
        }
    ]
}
EOF
```

生成ca证书和私钥
```bash
cfssl gencert -initca ca-csr.json | cfssljson -bare ca
```

创建server端证书
```bash
$ cat > server-csr.json <<EOF
{
  "CN": "admission",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
        "C": "CN",
        "L": "BeiJing",
        "ST": "BeiJing",
        "O": "k8s",
        "OU": "System"
    }
  ]
}
EOF

# hostname --> 自定义webhook服务在集群中的host
$ cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json \
		-hostname=admission-webhook-sample.default.svc -profile=server server-csr.json | cfssljson -bare server
```

使用生成的 server 证书和私钥创建一个 Secret 对象
```bash
$ kubectl create secret tls admission-webhook-sample-tls \
        --key=server-key.pem \
        --cert=server.pem
```

docker镜像构建
```bash
$ docker build -t hisunyh/admission-webhook-sample:v0.0.1 .
$ docker push hisunyh/admission-webhook-sample:v0.0.1
```

webhook服务部署
```bash
$ kubectl apply -f admission-webhook-sample_deployment.yaml
$ kubectl get all -l app=admission-webhook-sample
```

注册validating webhook
```bash
$ cat > validatingwebhook.yaml <<EOF
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: admission-webhook-sample
webhooks:
- name: easy.sunnyh.admission-webhook-sample
  rules:
  - apiGroups:   [""]
    apiVersions: ["v1"]
    operations:  ["CREATE"]
    resources:   ["pods"]
  clientConfig:
    service:
      namespace: default
      name: admission-webhook-sample
      path: "/validate"
    caBundle: CA_BUNDLE
  admissionReviewVersions: ["v1"]
  sideEffects: None
EOF

```

CA_BUNDLE的值是上面生成 ca.crt 文件内容的 base64 值
```shell
$ cat cert/ca.pem | base64 --wrap=0
```

```bash
kubectl apply -f validatingwebhook.yaml
kubectl get validatingwebhookconfiguration
```

测试
```bash
$ kubectl apply -f test-validate.yaml 
pod/test-pod created
Error from server: error when creating "test-validate.yaml": admission webhook "easy.sunnyh.admission-webhook-sample" denied the request: nginx:latest image comes from an untrusted registry! Only images from [docker.io gcr.io] are allowed.
```

remove
```bash
$ kubectl delete -f validatingwebhook.yaml 
$ kubectl delete -f admission-webhook-sample_deployment.yaml 

```

注册mutating webhook
```bash
$ cat > mutatingwebhook.yaml <<EOF
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: admission-webhook-sample-mutate
webhooks:
- name: easy.sunnyh.admission-webhook-sample-mutate
  clientConfig:
    service:
      namespace: default
      name: admission-webhook-sample
      path: "/mutate"
    caBundle: CA_BUNDLE
  rules:
    - operations: [ "CREATE" ]
      apiGroups: ["apps", ""]
      apiVersions: ["v1"]
      resources: ["deployments","services"]
  admissionReviewVersions: [ "v1" ]
  sideEffects: None
EOF

```


```bash
$ kubectl apply -f mutatingwebhook.yaml
$ kubectl get mutatingwebhookconfiguration
```