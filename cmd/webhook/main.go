// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.com/intel/network-resources-injector/pkg/webhook"
)

func main() {
	/* load configuration */
	port := flag.Int("port", 8443, "The port on which to serve.")
	address := flag.String("bind-address", "0.0.0.0", "The IP address on which to listen for the --port port.")
	cert := flag.String("tls-cert-file", "cert.pem", "File containing the default x509 Certificate for HTTPS.")
	key := flag.String("tls-private-key-file", "key.pem", "File containing the default x509 private key matching --tls-cert-file.")
	resourceNameKeys := flag.String("network-resource-name-keys", "k8s.v1.cni.cncf.io/resourceName", "comma separated resource name keys --network-resource-name-keys.")
	flag.Parse()

	if *port < 1024 || *port > 65535 {
		glog.Fatalf("invalid port number. Choose between 1024 and 65535")
	}

	if *address == "" || *cert == "" || *key == "" || *resourceNameKeys == "" {
		glog.Fatalf("input argument(s) not defined correctly")
	}

	glog.Infof("starting mutating admission controller for network resources injection")

	keyPair, err := webhook.NewTlsKeypairReloader(*cert, *key)
	if err != nil {
		glog.Fatalf("error load certificate: %s", err.Error())
	}

	/* init API client */
	webhook.SetupInClusterClient()

	err = webhook.SetResourceNameKeys(*resourceNameKeys)
	if err != nil {
		glog.Fatalf("error in setting resource name keys: %s", err.Error())
	}

	go func() {
		/* register handlers */
		var httpServer *http.Server

		http.HandleFunc("/mutate", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mutate" {
				http.NotFound(w, r)
				return
			}
			if r.Method != http.MethodPost {
				http.Error(w, "Invalid HTTP verb requested", 405)
				return
			}
			webhook.MutateHandler(w,r)
		})

		/* start serving */
		httpServer = &http.Server{
			Addr:              fmt.Sprintf("%s:%d", *address, *port),
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      10 * time.Second,
			MaxHeaderBytes:    1 << 20,
			ReadHeaderTimeout: 1 * time.Second,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384},
				PreferServerCipherSuites: true,
				InsecureSkipVerify: false,
				CipherSuites: []uint16 {
					// tls 1.2
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					// tls 1.3 configuration not supported
				},
				GetCertificate: keyPair.GetCertificateFunc(),
			},
		}

		err := httpServer.ListenAndServeTLS("", "")
		if err != nil {
			glog.Fatalf("error starting web server: %v", err)
		}
	}()

	/* watch the cert file and restart http sever if the file updated. */
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Fatalf("error starting fsnotify watcher: %v", err)
	}
	defer watcher.Close()

	certUpdated := false
	keyUpdated := false

	for {
		watcher.Add(*cert)
		watcher.Add(*key)

		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}
			glog.Infof("watcher event: %v", event)
			mask := fsnotify.Create | fsnotify.Rename | fsnotify.Remove |
				fsnotify.Write | fsnotify.Chmod
			if (event.Op & mask) != 0 {
				glog.Infof("modified file: %v", event.Name)
				if event.Name == *cert {
					certUpdated = true
				}
				if event.Name == *key {
					keyUpdated = true
				}
				if keyUpdated && certUpdated {
					if err := keyPair.Reload(); err != nil {
						glog.Fatalf("Failed to reload certificate: %v", err)
					}
					certUpdated = false
					keyUpdated = false
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			glog.Infof("watcher error: %v", err)
		}
	}
}
