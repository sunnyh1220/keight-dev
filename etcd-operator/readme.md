

```bash
kubebuilder init --domain sunnyh.easy --owner sunnyh --repo github.com/sunnyh1220/keight-dev/etcd-operator

kubebuilder create api --group etcd --version v1alpha1 --kind EtcdCluster
```

```bash
kubectl apply -f config/samples/etcd_v1alpha1_etcdcluster.yaml
```


```bash
etcdctl --endpoints etcdcluster-sample-0.etcdcluster-sample:2379,etcdcluster-sample-1.etcdcluster-sample:2379,etcdcluster-sample-2.etcdcluster-sample:2379 endpoint status --write-out=table
```