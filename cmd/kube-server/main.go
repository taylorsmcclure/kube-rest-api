package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	// Logging package
	log "github.com/sirupsen/logrus"

	// internal packages
	"github.com/taylorsmcclure/kube-server/internal/logger"
	k8sredis "github.com/taylorsmcclure/kube-server/internal/redis"

	// internal http handlers
	"github.com/taylorsmcclure/kube-server/internal/deployments"
	"github.com/taylorsmcclure/kube-server/internal/healthcheck"
	"github.com/taylorsmcclure/kube-server/internal/replicas"

	// k8s client packages
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// Gorilla Mux for routing
	"github.com/gorilla/mux"
)

// Version of the application; this is overwritten by the build process with -ldflags
const (
	Version = "development"
)

func main() {
	logger := logger.Setup(false)

	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Could not find user home directory : %s", err)
	}

	// Command line arguments
	var port, kubeconfig, rAddr, ca, cert, key, rClientCert, rCACert, rClientKey string
	var local, verbose, version bool
	flag.StringVar(&port, "port", "8080", "server port")
	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir, ".kube", "config"), "path to the kubeconfig file")
	flag.StringVar(&rAddr, "raddr", "localhost:6379", "Address of the Redis server, like: localhost:6379")
	flag.StringVar(&ca, "ca", "", "path to ca cert for the server")
	flag.StringVar(&cert, "cert", "", "path to cert for the server")
	flag.StringVar(&key, "key", "", "path to key for the server")
	flag.StringVar(&rCACert, "rca", "", "path to ca cert for Redis")
	flag.StringVar(&rClientCert, "rcert", "", "path to cert for Redis")
	flag.StringVar(&rClientKey, "rkey", "", "path to key for Redis")
	flag.BoolVar(&version, "version", false, "prints out the version of the application")
	flag.BoolVar(&local, "local", false, "use kubeconfig on local machine instead of cluster ServiceAccount")
	flag.BoolVar(&verbose, "verbose", false, "Enables verbose output")

	flag.Parse()

	// Set verbose logging if needed
	if verbose {
		logger.SetLevel(log.DebugLevel)
		logger.Debug("Logging verbosely")
	}

	if version {
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	// Generate the Kubernetes client set to access the cluster
	kClient, err := clusterLogin(local, kubeconfig)
	if err != nil {
		logger.Fatalf("Error creating kubernetes client: %s", err)
	}

	// Create a CA cert pool for Redis and server mTLS
	caCertPool := x509.NewCertPool()

	// Load the server CA and append to the CA pool
	redisCACert, err := os.ReadFile(rCACert)
	if err != nil {
		logger.Fatal(err)
	}
	caCertPool.AppendCertsFromPEM(redisCACert)

	// Load Redis client cert and key
	redisClientCert, err := tls.LoadX509KeyPair(rClientCert, rClientKey)
	if err != nil {
		logger.Fatal(err)
	}

	// Create the TLS Config for mTLS connection to Redis
	redisTLSConfig := &tls.Config{
		Certificates:       []tls.Certificate{redisClientCert},
		InsecureSkipVerify: true,
	}

	// Create the Redis client
	rClient, err := k8sredis.NewClient(rAddr, redisTLSConfig)
	if err != nil {
		logger.Fatalf("Error creating redis client: %s", err)
	}

	// Gorilla Mux was chosen for the router over the built-in due to it handling parameters in the URI better
	// We are passing in the kubernetes clientSet and redis client to the handlers where appropriate
	r := mux.NewRouter()
	r.HandleFunc("/v1/deployments", func(w http.ResponseWriter, r *http.Request) {
		deployments.V1Deployments(w, r, kClient)
	})
	r.HandleFunc("/v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		healthcheck.V1HealthCheck(w, r, kClient, Version)
	})
	r.HandleFunc("/v1/replicas/{namespace}/{deployment}", func(w http.ResponseWriter, r *http.Request) {
		replicas.V1Replicas(w, r, kClient, rClient)
	})
	// Catches replicas requests with incomplete paths
	r.HandleFunc("/v1/replicas/{namespace}", func(w http.ResponseWriter, r *http.Request) {
		replicas.V1Replicas(w, r, kClient, rClient)
	})
	r.HandleFunc("/v1/replicas", func(w http.ResponseWriter, r *http.Request) {
		replicas.V1Replicas(w, r, kClient, rClient)
	})

	// TODO: we should implement a logging middleware which gorilla mux supports natively
	// We should also track API requests with UUIDs and forward them to clients for easier troubleshooting
	// r.Use(loggingMiddleware)

	// Create the mTLS server
	// Load the server CA and append to the CA pool
	caCert, err := os.ReadFile(ca)
	if err != nil {
		logger.Fatal(err)
	}
	caCertPool.AppendCertsFromPEM(caCert)

	// Create the TLS Config with the CA pool and enable Client certificate validation
	serverTLSConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:         ":" + port,
		TLSConfig:    serverTLSConfig,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start the TLS server with Gorilla Mux as the router
	logger.Infof("Application version is: %s", Version)
	logger.Infof("Starting server on localhost:%s", port)
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		logger.Fatalf("Error: %v", err)
	}

}

// Logs into the kubernetes cluster and get the clientset
func clusterLogin(local bool, kubeconfig string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if local {
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return &kubernetes.Clientset{}, err
		}
	} else {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return &kubernetes.Clientset{}, err
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &kubernetes.Clientset{}, err
	}

	return clientset, nil
}
