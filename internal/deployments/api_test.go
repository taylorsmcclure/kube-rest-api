package deployments

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/taylorsmcclure/kube-server/internal/logger"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"
)

// I don't like being dependent on the internal package, but
// this causes a nil pointer exception if it isn't initialized
func init() {
	logger.Setup(false)
}

// Tests getDeployments() with multiple cases
func TestGetDeployments(t *testing.T) {

	testCases := []struct {
		name             string
		namespace        string
		deployments      []runtime.Object
		expectSuccess    bool
		expectedResponse getDeploymentsResponse
	}{
		{
			name:      "one_deployment_filter",
			namespace: "test",
			deployments: []runtime.Object{&appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment",
							Namespace: "test",
						},
					},
				},
			},
			},
			expectSuccess: true,
			expectedResponse: getDeploymentsResponse{
				Code: 200,
				Deployments: []DeployNamespace{
					{
						Deployment: "test_deployment",
						Namespace:  "test",
					},
				},
			},
		},
		{
			name:      "no_deployments",
			namespace: "none",
			deployments: []runtime.Object{&appsv1.DeploymentList{
				Items: []appsv1.Deployment{},
			},
			},
			expectSuccess:    false,
			expectedResponse: getDeploymentsResponse{},
		},
		{
			name:      "multi_deployment_filter",
			namespace: "test_multi",
			deployments: []runtime.Object{&appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment0",
							Namespace: "test_multi",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment1",
							Namespace: "test_multi",
						},
					},
				},
			},
			},
			expectSuccess: true,
			expectedResponse: getDeploymentsResponse{
				Code: 200,
				Deployments: []DeployNamespace{
					{
						Deployment: "test_deployment0",
						Namespace:  "test_multi",
					},
					{
						Deployment: "test_deployment1",
						Namespace:  "test_multi",
					},
				},
			},
		},
		{
			name:      "multi_deployment_no_filter",
			namespace: "",
			deployments: []runtime.Object{&appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment0",
							Namespace: "test",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment1",
							Namespace: "test",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_dos_deployment0",
							Namespace: "test_dos",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_tres_deployment0",
							Namespace: "test_tres",
						},
					},
				},
			},
			},
			expectSuccess: true,
			expectedResponse: getDeploymentsResponse{
				Code: 200,
				Deployments: []DeployNamespace{
					{
						Deployment: "test_deployment0",
						Namespace:  "test",
					},
					{
						Deployment: "test_deployment1",
						Namespace:  "test",
					},
					{
						Deployment: "test_dos_deployment0",
						Namespace:  "test_dos",
					},
					{
						Deployment: "test_tres_deployment0",
						Namespace:  "test_tres",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fakeClientset := testclient.NewSimpleClientset(test.deployments...)
			resp, err := getDeployments(
				fakeClientset,
				test.namespace,
			)
			switch {
			case test.expectSuccess && err != nil:
				t.Errorf("expected success, got error: %v", err)
			case !test.expectSuccess && err == nil:
				t.Errorf("expected error, got success")
			case test.expectSuccess && !reflect.DeepEqual(resp, &test.expectedResponse):
				t.Errorf("expected %v, got %v", test.expectedResponse, resp)
			case !test.expectSuccess && !reflect.DeepEqual(resp, &test.expectedResponse):
				t.Errorf("expected %v, got %v", test.expectedResponse, resp)
			default:
				t.Logf("test passed with %v", resp)
			}

		})
	}

}

// Tests the HTTP GET endpoint for /v1/deployments
func TestV1Deployments(t *testing.T) {

	testCases := []struct {
		name             string
		namespace        string
		deployments      []runtime.Object
		expectSuccess    bool
		expectedResponse getDeploymentsResponse
	}{
		{
			name:      "one_deployment_filter",
			namespace: "test",
			deployments: []runtime.Object{&appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment",
							Namespace: "test",
						},
					},
				},
			},
			},
			expectSuccess: true,
			expectedResponse: getDeploymentsResponse{
				Code: 200,
				Deployments: []DeployNamespace{
					{
						Deployment: "test_deployment",
						Namespace:  "test",
					},
				},
			},
		},
		{
			name:      "no_deployments",
			namespace: "none",
			deployments: []runtime.Object{&appsv1.DeploymentList{
				Items: []appsv1.Deployment{},
			},
			},
			expectSuccess:    false,
			expectedResponse: getDeploymentsResponse{},
		},
		{
			name:      "multi_deployment_filter",
			namespace: "test_multi",
			deployments: []runtime.Object{&appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment0",
							Namespace: "test_multi",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment1",
							Namespace: "test_multi",
						},
					},
				},
			},
			},
			expectSuccess: true,
			expectedResponse: getDeploymentsResponse{
				Code: 200,
				Deployments: []DeployNamespace{
					{
						Deployment: "test_deployment0",
						Namespace:  "test_multi",
					},
					{
						Deployment: "test_deployment1",
						Namespace:  "test_multi",
					},
				},
			},
		},
		{
			name:      "multi_deployment_no_filter",
			namespace: "",
			deployments: []runtime.Object{&appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment0",
							Namespace: "test",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_deployment1",
							Namespace: "test",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_dos_deployment0",
							Namespace: "test_dos",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test_tres_deployment0",
							Namespace: "test_tres",
						},
					},
				},
			},
			},
			expectSuccess: true,
			expectedResponse: getDeploymentsResponse{
				Code: 200,
				Deployments: []DeployNamespace{
					{
						Deployment: "test_deployment0",
						Namespace:  "test",
					},
					{
						Deployment: "test_deployment1",
						Namespace:  "test",
					},
					{
						Deployment: "test_dos_deployment0",
						Namespace:  "test_dos",
					},
					{
						Deployment: "test_tres_deployment0",
						Namespace:  "test_tres",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/deployments", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Adds a namespace query to the request
			q := req.URL.Query()
			q.Add("namespace", test.namespace)
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()

			fakeClientset := testclient.NewSimpleClientset(test.deployments...)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				V1Deployments(w, r, fakeClientset)
			})

			handler.ServeHTTP(rr, req)

			switch {
			case test.expectSuccess && rr.Code != http.StatusOK:
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, http.StatusOK)
			default:
				t.Logf("test passed %v", rr.Code)
			}

		},
		)
	}

}
