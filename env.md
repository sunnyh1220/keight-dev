
## Env
go version 1.15.x

### kubectl
```bash
# curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
curl -LO https://dl.k8s.io/release/v1.19.15/bin/linux/amd64/kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

### minikube
```bash
# install
curl -Lo minikube https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/releases/v1.19.0/minikube-linux-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/

# uninstall
rm /usr/local/bin/minikube
rm -rf ~/.minikube

# Start a cluster 
minikube start --driver=docker --kubernetes-version=v1.19.15 --image-mirror-country cn

# Stop your local cluster
minikube stop

# Delete your local cluster
minikube delete
```

### operator-sdk
reference: https://sdk.operatorframework.io/docs/installation/
```bash
# install
export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
export OS=$(uname | awk '{print tolower($0)}')

export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.7.2
curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}

chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk


# remove
sudo rm /usr/local/bin/operator-sdk
```

### kubebuilder
```bash
# go1.15 使用 kubebuilder3.0.0
$ curl -L -o kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.0.0/kubebuilder_linux_amd64
$ chmod +x kubebuilder && mv kubebuilder /usr/local/bin/

$ kubebuilder version
Version: main.version{KubeBuilderVersion:"3.0.0", KubernetesVendor:"1.19.2", GitCommit:"533874b302e9bf94cd7105831f8a543458752973", BuildDate:"2021-04-28T16:23:59Z", GoOs:"linux", GoArch:"amd64"}

```