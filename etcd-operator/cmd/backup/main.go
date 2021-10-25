package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	etcdv1alpha1 "github.com/sunnyh1220/keight-dev/etcd-operator/api/v1alpha1"
	uploader "github.com/sunnyh1220/keight-dev/etcd-operator/pkg/file"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/snapshot"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

func loggedError(log logr.Logger, err error, message string) error {
	log.Error(err, message)
	return fmt.Errorf("%s: %s", message, err)
}

func main() {
	var (
		backupTempDir          string
		etcdURL                string
		etcdDialTimeoutSeconds int64
		timeoutSeconds         int64
		backupURL              string
	)

	flag.StringVar(&backupTempDir, "backup-tmp-dir", os.TempDir(), "The directory to temporarily place backups before they are uploaded to their destination.")
	flag.StringVar(&etcdURL, "etcd-url", "http://localhost:2379", "URL for etcd.")
	flag.Int64Var(&etcdDialTimeoutSeconds, "etcd-dial-timeout-seconds", 5, "Timeout, in seconds, for dialing the Etcd API.")
	flag.Int64Var(&timeoutSeconds, "timeout-seconds", 60, "Timeout, in seconds, of the whole restore operation.")
	flag.StringVar(&backupURL, "backup-url", "", "URL for the backup storage.")
	flag.Parse()

	zapLogger := zap.NewRaw(zap.UseDevMode(true))
	ctrl.SetLogger(zapr.NewLogger(zapLogger))
	log := ctrl.Log.WithName("backup-agent")

	// 解析备份上传对象存储参数
	storageType, bucketName, objectName, err := uploader.ParseBackupURL(backupURL)
	if err != nil {
		panic(loggedError(log, err, "failed to parse etcd backup url"))
	}

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
	defer ctxCancel()

	log.Info("Connecting to Etcd and getting snapshot")
	// 本地临时保存数据目录
	localPath := filepath.Join(backupTempDir, "snapshot.db")
	etcdClient := snapshot.NewV3(zapLogger.Named("etcd-client"))
	// 保存etcd snapshot到local path
	err = etcdClient.Save(ctx, clientv3.Config{
		Endpoints:   []string{etcdURL},
		DialTimeout: time.Second * time.Duration(etcdDialTimeoutSeconds),
	}, localPath)
	if err != nil {
		panic(loggedError(log, err, "failed to get etcd snapshot"))
	}

	log.Info("Uploading snapshot...")
	switch storageType {
	case string(etcdv1alpha1.EtcdBackupStorageTypeS3): // s3
		// 数据保存到本地成功
		// 上传到S3
		size, err := handleS3(ctx, localPath, bucketName, objectName)
		if err != nil {
			panic(loggedError(log, err, "failed to upload backup etcd"))
		}
		log.WithValues("upload-size", size).Info("Backup completed")
	case string(etcdv1alpha1.EtcdBackupStorageTypeOSS): // oss
	default:
		panic(loggedError(log, fmt.Errorf("storage type error"), fmt.Sprintf("unknown StorageType: %v", storageType)))
	}

}

func handleS3(ctx context.Context, localPath string, bucketName string, objectName string) (int64, error) {
	// 根据传递进来的参数（环境变量）获取s3配置信息
	endpoint := os.Getenv("ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	//	endpoint := "play.min.io"
	//	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	//	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	//	bucketName := "1-sunny-backup"
	//	objectName := "etcd-snapshot.db"
	s3Uploader := uploader.NewS3Uploader(endpoint, accessKeyID, secretAccessKey)
	return s3Uploader.Upload(ctx, localPath, bucketName, objectName)
}
