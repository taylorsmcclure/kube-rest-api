package deployments

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	// internal packages
	e "github.com/taylorsmcclure/kube-server/internal/errors"
	"github.com/taylorsmcclure/kube-server/internal/logger"
	"github.com/taylorsmcclure/kube-server/internal/responses"

	// k8s api packages
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// JSON response for deployments
type getDeploymentsResponse struct {
	Code        int               `json:"http_response_code"`
	Deployments []DeployNamespace `json:"deployments"`
}

// Struct for a list of deployments and namespaces
type DeployNamespace struct {
	Deployment string `json:"deployment_name"`
	Namespace  string `json:"namespace"`
}

// Error type for when there are no deployments found
type errNoDeployments error

// Lists all deployments on the cluster
func V1Deployments(w http.ResponseWriter, r *http.Request, kClient kubernetes.Interface) {
	// Catch fatal errors that would otherwise cause the server to quit
	defer e.NonFatal()

	// We may want to support more methods in the future, so we'll use a switch statement
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		var namespace string
		// Allow filtering by namespace deployments?namespace=<namespace>
		if len(r.URL.Query()["namespace"]) == 0 {
			namespace = ""
		} else {
			namespace = r.URL.Query()["namespace"][0]
		}
		// Get the deployments
		resp, err := getDeployments(kClient, namespace)
		if err != nil {
			// Handle specific error types
			// TODO: implement a more scalable HTTP error handling system and DRY it up
			switch err.(type) {
			case errNoDeployments:
				responses.ReturnJsonResponse(w, 404, e.GenericError{Code: 404, Message: fmt.Sprint(err)})
				return
			default:
				logger.Log.Error(err)
				responses.ReturnJsonResponse(w, 500, e.GenericError{Code: 500, Message: "Internal server"})
			}
		}
		responses.ReturnJsonResponse(w, 200, resp)
	default:
		responses.ReturnJsonResponse(w, 405, e.GenericError{Code: 405, Message: "method not allowed"})
	}
}

// Gets all deployments on the cluster or filter by namespace
func getDeployments(kClient kubernetes.Interface, namespace string) (*getDeploymentsResponse, error) {
	deployments, err := kClient.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return &getDeploymentsResponse{}, err
	}

	// If there are no deployments, return a specific error
	if len(deployments.Items) == 0 {
		return &getDeploymentsResponse{}, errNoDeployments(errors.New("no deployments found"))
	}

	var availableDeployments []DeployNamespace
	for _, d := range deployments.Items {
		availableDeployments = append(availableDeployments, DeployNamespace{Deployment: d.Name, Namespace: d.Namespace})
	}

	resp := &getDeploymentsResponse{Code: 200, Deployments: availableDeployments}

	return resp, nil
}
