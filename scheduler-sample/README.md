## Scheduler Plugins

> reference:
>
> https://github.com/kubernetes-sigs/scheduler-plugins
>
> [K8s - Create a kube-scheduler plugin](https://medium.com/@juliorenner123/k8s-creating-a-kube-scheduler-plugin-8a826c486a1)



```bash
# build
docker build -t hisunyh/scheduler-sample:v0.0.1 -f Dockerfile .
docker push hisunyh/scheduler-sample:v0.0.1

# deploy
kubectl apply -f scheduler-sample-deploy.yaml

# test
kubectl apply -f test-deployment.yaml
```

