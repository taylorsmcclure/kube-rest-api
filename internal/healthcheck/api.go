package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	e "github.com/taylorsmcclure/kube-server/internal/errors"
	"github.com/taylorsmcclure/kube-server/internal/logger"
	"github.com/taylorsmcclure/kube-server/internal/responses"

	"k8s.io/client-go/kubernetes"
)

// Struct for the respons of the endpoint
type getLivezResponse struct {
	Code    int    `json:"http_response_code"`
	Status  string `json:"kubernetes_api_status"`
	Version string `json:"application_version"`
}

type errHealthCheckFailed error

// API endpoint for checking the health of the cluster and application
func V1HealthCheck(w http.ResponseWriter, r *http.Request, kClient kubernetes.Interface, Version string) {
	// Catch fatal errors that would otherwise cause the server to quit
	defer e.NonFatal()

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		resp, err := getLivez(kClient, Version)
		if err != nil {
			// Handle specific error types
			// TODO: implement a more scalable HTTP error handling system and DRY it up
			switch err.(type) {
			case errHealthCheckFailed:
				logger.Log.Error(err)
				responses.ReturnJsonResponse(w, 500, e.GenericError{Code: 500, Message: fmt.Sprint(err)})
				return
			default:
				logger.Log.Error(err)
				responses.ReturnJsonResponse(w, 500, e.GenericError{Code: 500, Message: "Internal server error"})
			}
			return
		}
		responses.ReturnJsonResponse(w, 200, resp)
	default:
		responses.ReturnJsonResponse(w, 405, e.GenericError{Code: 405, Message: "method not allowed"})
	}
}

// Checks the livez endpoint of the cluster
func getLivez(kClient kubernetes.Interface, Version string) (*getLivezResponse, error) {
	defer e.NonFatal()
	// Setting an int to capture the response code
	var statusCode int
	err := kClient.DiscoveryV1().RESTClient().Get().AbsPath("/livez").Do(context.TODO()).StatusCode(&statusCode)
	fmt.Print(err)

	if statusCode != 200 {
		err_message := "kubernetes API /livez check failed, cluster is unhealthy"
		return &getLivezResponse{}, errHealthCheckFailed(errors.New(err_message))
	}

	return &getLivezResponse{Code: 200, Status: "ok", Version: Version}, nil
}
