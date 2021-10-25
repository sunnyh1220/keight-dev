

## etcd创建
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

### etcd备份
```bash
kubebuilder create api --group etcd --version v1alpha1 --kind EtcdBackup
```

```bash
docker build --target backup -t hisunyh/etcd-operator-backup:v0.0.1 -f Dockerfile .
docker push hisunyh/etcd-operator-backup:v0.0.1

kubectl create secret generic minio-access-secret --from-literal=MINIO_ACCESS_KEY=Q3AM3UQ867SPQQA43P2F --from-literal=MINIO_SECRET_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG

kubectl apply -f config/samples/etcd_v1alpha1_etcdbackup.yaml
```