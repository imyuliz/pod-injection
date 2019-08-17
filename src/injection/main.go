package main

import (
	"crypto/tls"
	"flag"
	"injection/webhook"
	"net/http"

	mlog "github.com/maxwell92/log"
)

var (
	glog = mlog.Log
)

// TLSConf tls config
type TLSConf struct {
	KeyFile  string `json:"kertFile,omitempty"`
	CertFile string `json:"certFile,omitempty"`
}

func main() {
	var tlsConf TLSConf
	// get command line parameters
	flag.StringVar(&tlsConf.CertFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&tlsConf.KeyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()
	pair, err := tls.LoadX509KeyPair(tlsConf.CertFile, tlsConf.KeyFile)
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}
	server := http.Server{Addr: ":443", TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}}}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", webhook.ServerPodInjection)
	server.Handler = mux
	glog.Infoln("Server starting...")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		glog.Errorf("ListenAndServeTLS failed: %v", err)
	}
}
