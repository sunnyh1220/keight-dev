
## Controller Sample Step By Step
> reference： https://github.com/kubernetes/sample-controller
>

### 项目初始化
init project
```bash
mkdir controller-sample && cd controller-sample
go mod init github.com/sunnyh1220/keight-dev/controller-sample
go get k8s.io/apimachinery@v0.19.15
go get k8s.io/client-go@v0.19.15
go get -d k8s.io/code-generator@v0.19.15
```

custom resource define
```bash
mkdir -p config/crd && cd config/crd

tee crd-crontab.yaml <<-'EOF'
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crontabs.stable.sunnyh.easy
spec:
  group: stable.sunnyh.easy
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          description: CronTab is the Schema for the crontabs API
          type: object
          properties:
            spec:
              type: object
              properties:
                cronSpec:
                  type: string
                image:
                  type: string
                replicas:
                  type: integer
  scope: Namespaced
  names:
    kind: CronTab
    plural: crontabs
    singular: crontab
    shortNames:
      - ct
EOF

```

CRD资源类型结构体初始化

```bash
mkdir -p pkg/apis/stable/v1alpha1 && cd pkg/apis/stable/v1alpha1

# doc.go
tee doc.go <<-'EOF'
// +k8s:deepcopy-gen=package
// +groupName=stable.sunnyh.easy

package v1alpha1
EOF

# type.go

# register.go

```

type.go --> 自定义资源类型定义

register.go -->  Resource(),AddToSchema实现


generate自动生成代码
```bash
mkdir hack && cd hack

# tools.go
tee tools.go <<-'EOF'
//go:build tools
// +build tools

// This package imports things required by build scripts, to force `go mod` to see them as dependencies
package tools

import _ "k8s.io/code-generator"
EOF

# update-codegen.sh
# 生成的代码

# verify-codegen.sh
# 校验生成的代码是否是最新

# boilerplate.go.txt
# 生成文件头部模板文件

```

`update-codegen.sh` 根据项目结构修改

```shell
#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

bash "${CODEGEN_PKG}"/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/cnych/controller-demo/pkg/client github.com/cnych/controller-demo/pkg/apis \
  stable:v1alpha1 \
  --output-base "${SCRIPT_ROOT}"/../../../.. \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

# To use your own boilerplate text append:
#   --go-header-file "${SCRIPT_ROOT}"/hack/custom-boilerplate.go.txt
```

`verify-codegen.sh` 不需要修改

```shell
#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

DIFFROOT="${SCRIPT_ROOT}/pkg"
TMP_DIFFROOT="${SCRIPT_ROOT}/_tmp/pkg"
_tmp="${SCRIPT_ROOT}/_tmp"

cleanup() {
  rm -rf "${_tmp}"
}
trap "cleanup" EXIT SIGINT

cleanup

mkdir -p "${TMP_DIFFROOT}"
cp -a "${DIFFROOT}"/* "${TMP_DIFFROOT}"

"${SCRIPT_ROOT}/hack/update-codegen.sh"
echo "diffing ${DIFFROOT} against freshly generated codegen"
ret=0
diff -Naupr "${DIFFROOT}" "${TMP_DIFFROOT}" || ret=$?
cp -a "${TMP_DIFFROOT}"/* "${DIFFROOT}"
if [[ $ret -eq 0 ]]
then
  echo "${DIFFROOT} up to date."
else
  echo "${DIFFROOT} is out of date. Please run hack/update-codegen.sh"
  exit 1
fi
```

code-generator依赖脚本放到vendor
```shell
go mod vendor
```
生成代码
```bash
chmod +x ./hack/update-codegen.sh
./hack/update-codegen.sh
```

```bash
sunnyh@keight-vm:~/go/src/github.com/sunnyh1220/keight-dev/controller-sample$ ./hack/update-codegen.sh
Generating deepcopy funcs
Generating clientset for stable:v1alpha1 at github.com/sunnyh1220/keight-dev/controller-sample/pkg/client/clientset
Generating listers for stable:v1alpha1 at github.com/sunnyh1220/keight-dev/controller-sample/pkg/client/listers
Generating informers for stable:v1alpha1 at github.com/sunnyh1220/keight-dev/controller-sample/pkg/client/informers

```

代码检查
```bash
chmod +x ./hack/verify-codegen.sh
./hack/verify-codegen.sh
```

### 控制器实现
controllers/crontab_controller.go
main.go

编译
```bash
go build -ldflags "-s -w" -v -o crontab-controller .

```

crd install
```bash
kubectl apply -f config/crd/crd-crontab.yaml
```
apply crontab sample
```bash
kubectl apply -f config/samples/stable_v1alpha1_crontab.yaml
```

