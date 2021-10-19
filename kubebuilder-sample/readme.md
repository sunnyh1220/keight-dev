

## kubebuilder sample step by step

kubebuilder install
```bash
# go1.15 使用 kubebuilder3.0.0
$ curl -L -o kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.0.0/kubebuilder_linux_amd64
$ chmod +x kubebuilder && mv kubebuilder /usr/local/bin/

$ kubebuilder version
Version: main.version{KubeBuilderVersion:"3.0.0", KubernetesVendor:"1.19.2", GitCommit:"533874b302e9bf94cd7105831f8a543458752973", BuildDate:"2021-04-28T16:23:59Z", GoOs:"linux", GoArch:"amd64"}

```

Create a Project
```bash
mkdir kubebuilder-sample
cd kubebuilder-sample
kubebuilder init --domain sunnyh.easy --owner sunnyh --repo github.com/sunnyh1220/keight-dev/kubebuilder-sample

```

Create an API
```bash
kubebuilder create api --group batch --version v1 --kind CronJob

```

Designing an APi
```go
// api/v1/cronjob_types.go

```

Implementing a controller
```go 
// controllers/cronjob_controller.go
```

Run
```bash
# install crd
make install

# run controller
maek run

# apply cronjob sample
kubectl apply -f config/samples/batch_v1_cronjob.yaml
```