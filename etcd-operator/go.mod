module github.com/sunnyh1220/keight-dev/etcd-operator

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.2.0
	github.com/minio/minio-go/v7 v7.0.15
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200819165624-17cef6e3e9d5
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.7.2
)
