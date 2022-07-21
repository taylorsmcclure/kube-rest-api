package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var ServerPort string

func main() {

	// Command line arguments
	var port, ca, cert, key string
	flag.StringVar(&port, "port", "8443", "server port")
	flag.StringVar(&ca, "ca", "", "path to ca cert for the server")
	flag.StringVar(&cert, "cert", "", "path to cert for the server")
	flag.StringVar(&key, "key", "", "path to key for the server")

	flag.Parse()

	ServerPort = port

	kubeServerCerts, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	serverCACert, err := ioutil.ReadFile(ca)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool.AppendCertsFromPEM(serverCACert)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{kubeServerCerts},
		RootCAs:            caCertPool,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	testCases := []struct {
		name          string
		description   string
		endpoint      string
		method        string
		expectSuccess bool
		statusCode    int
		payload       postPayload
	}{
		{
			name:          "healthz-get",
			description:   "Testing healthz endpoint",
			expectSuccess: true,
			method:        "GET",
			endpoint:      "/v1/healthz",
			statusCode:    200,
		},
		{
			name:          "healthz-404",
			description:   "Testing malformed healthz endpoint",
			expectSuccess: true,
			method:        "GET",
			endpoint:      "/v1/heaalthz",
			statusCode:    404,
		},
		{
			name:          "deployments-get",
			description:   "Testing GET on deployments endpoint",
			expectSuccess: true,
			method:        "GET",
			endpoint:      "/v1/deployments",
			statusCode:    200,
		},
		{
			name:          "deployments-filter-get",
			description:   "Testing GET on deployments endpoint with a namespace filter",
			expectSuccess: true,
			method:        "GET",
			endpoint:      "/v1/deployments?namespace=busybox-test",
			statusCode:    200,
		},
		{
			name:          "replicas-get",
			description:   "Testing GET on replicas endpoint",
			expectSuccess: true,
			method:        "GET",
			endpoint:      "/v1/replicas/busybox-test/busybox-deployment0",
			statusCode:    200,
		},
		{
			name:          "replicas-post",
			description:   "Testing POST on replicas endpoint",
			expectSuccess: true,
			method:        "POST",
			endpoint:      "/v1/replicas/busybox-test/busybox-deployment0",
			statusCode:    200,
			payload: postPayload{
				ReplicaSize: 1,
			},
		},
	}

	// Loop to run each test case
	for _, test := range testCases {
		code, err := getTest(test.endpoint, client, test.method, &test.payload)
		if err != nil {
			log.Fatal(err)
		}
		if test.expectSuccess && code != test.statusCode {
			log.Fatalf("%s: expected status code %d, got %d", test.name, test.statusCode, code)
		} else {
			fmt.Printf("%s: passed\n", test.name)
		}
	}

}

type postPayload struct {
	ReplicaSize int `json:"replica_size"`
}

func getTest(endpoint string, client *http.Client, method string, payload *postPayload) (int, error) {
	var statusCode int
	url := "https://localhost:" + ServerPort + endpoint

	switch {
	case method == "GET":
		fmt.Printf("GET %s\n", url)
		resp, err := client.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		statusCode = resp.StatusCode
		return statusCode, nil
	case method == "POST":
		fmt.Printf("POST %s\nPAYLOAD: %v", url, payload)

		reqJson, err := json.Marshal(payload)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqJson))
		if err != nil {
			log.Fatal(err)
		}
		statusCode = resp.StatusCode
		return statusCode, nil
	default:
		fmt.Printf("Invalid method %s", method)
		return 0, nil
	}

}
