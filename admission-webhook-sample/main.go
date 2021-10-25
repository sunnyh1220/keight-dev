package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/sunnyh1220/keight-dev/admission-webhook-sample/pkg/server"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	port     int
	certFile string
	keyFile  string
)

func main() {
	flag.IntVar(&port, "port", 443, "Webhook server port.")
	flag.StringVar(&certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	klog.Info(fmt.Sprintf("port=%d, cert-file=%s, key-file=%s", port, certFile, keyFile))

	pair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
		return
	}

	whsvr := &server.WebhookServer{
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
		WhiteListRegistries: strings.Split(os.Getenv("WHITELIST_REGISTRIES"), ","),
	}

	// 定义 http server 和 handler
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", whsvr.Handler)
	mux.HandleFunc("/mutate", whsvr.Handler)
	whsvr.Server.Handler = mux

	// 在一个新的 goroutine 中启动 webhook server
	go func() {
		if err := whsvr.Server.ListenAndServeTLS("", ""); err != nil {
			klog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	klog.Info("Server started")

	// 监听 OS shutdown 信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	klog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	if err := whsvr.Server.Shutdown(context.Background()); err != nil {
		klog.Errorf("HTTP server Shutdown: %v", err)
	}

}
