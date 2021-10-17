
```bash
operator-sdk init --domain sunnyh.easy --repo github.com/sunnyh1220/keight-dev/operator-sdk-sample --owner sunnyh

operator-sdk create api --group app --version v1alpha1 --kind AppService --resource --controller


make manifests

make install
# make uninstall

make run

kubectl apply -f config/samples/app_v1alpha1_appservice.yaml
kubectl delete -f config/samples/app_v1alpha1_appservice.yaml

make docker-build IMG="operator-sdk-sample:v0.0.1"
make docker-push IMG="operator-sdk-sample:v0.0.1"
make deploy IMG="operator-sdk-sample:v0.0.1"
```