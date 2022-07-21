package replicas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	// internal packages
	"github.com/go-redis/redis/v8"
	e "github.com/taylorsmcclure/kube-server/internal/errors"
	"github.com/taylorsmcclure/kube-server/internal/logger"
	"github.com/taylorsmcclure/kube-server/internal/responses"

	// Kubernetes packages
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// Handles the /v1/replicas endpoint
func V1Replicas(w http.ResponseWriter, r *http.Request, kClient kubernetes.Interface, rClient *redis.Client) {
	// Check if namespace and deployment are in the request
	reqURI := strings.Split(r.URL.Path, "/")
	if len(reqURI) < 5 {
		responses.ReturnJsonResponse(w, 400, &e.GenericError{Code: 400, Message: "No namespace and/or deployment provided"})
		return
	}
	// TODO: we should have better validation regex
	namespace := reqURI[3]
	deployment := reqURI[4]

	// Support both GET and POST requests on the replicas endpoint
	switch r.Method {
	// Handle the GET request
	case http.MethodGet, http.MethodHead:
		resp, err := getReplicas(kClient, rClient, namespace, deployment)
		if err != nil {
			// Handle k8s API specific errors and send to the client
			if statusError, isStatus := err.(*errors.StatusError); isStatus {
				responses.ReturnJsonResponse(w, int(statusError.ErrStatus.Code), &e.GenericError{Code: int(statusError.ErrStatus.Code), Message: fmt.Sprint(err)})
			} else {
				responses.ReturnJsonResponse(w, 500, &e.GenericError{Code: 500, Message: "Internal server error"})
			}
			return
		}
		responses.ReturnJsonResponse(w, 200, resp)
	// Handle the POST request
	case http.MethodPost:
		// Get replica_size from the data in the POST
		var req setReplicasRequest
		// Don't allow any other json fields in payload except for what's in setReplicasRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		err := decoder.Decode(&req)
		if err != nil {
			responses.ReturnJsonResponse(w, 400, &e.GenericError{Code: 400, Message: "Bad request"})
			return
		}
		// Set the replicas
		resp, err := setReplicas(kClient, rClient, namespace, deployment, req.ReplicaSize)
		if err != nil {
			if statusError, isStatus := err.(*errors.StatusError); isStatus {
				responses.ReturnJsonResponse(w, int(statusError.ErrStatus.Code), &e.GenericError{Code: int(statusError.ErrStatus.Code), Message: fmt.Sprint(err)})
			} else {
				responses.ReturnJsonResponse(w, 500, &e.GenericError{Code: 500, Message: "Internal server error"})
			}
			return
		}
		responses.ReturnJsonResponse(w, 200, resp)
	default:
		responses.ReturnJsonResponse(w, 405, e.GenericError{Code: 405, Message: "method not allowed"})
	}
}

// Parse incoming payload from client
type setReplicasRequest struct {
	ReplicaSize int32 `json:"replica_size"`
}

// Response to client when GET request is made
type getReplicasResponse struct {
	Namespace       string `json:"namespace"`
	Deployment      string `json:"deployment_name"`
	CurrentReplicas int32  `json:"current_replicas"`
	DesiredReplicas int32  `json:"desired_replicas"`
	Drift           bool   `json:"state_drift"`
	Code            int    `json:"http_status_code"`
}

// Response to client when they make a POST request
type setReplicasResponse struct {
	Namespace         string `json:"namespace"`
	Deployment        string `json:"deployment_name"`
	CurrentReplicas   int32  `json:"current_replicas"`
	DesiredReplicas   int32  `json:"desired_replicas"`
	RequestedReplicas int32  `json:"requested_replicas"`
	Drift             bool   `json:"state_drift"`
	Code              int    `json:"http_status_code"`
}

// Value of the replicas key in Redis
type redisValue struct {
	DesiredReplicas int32 `json:"desired_replicas"`
	CurrentReplicas int32 `json:"current_replicas"`
	Drift           bool  `json:"state_drift"`
}

// Helper to form the key for state in Redis
func genRedisKey(namespace, deployment string) string {
	return fmt.Sprintf("%s-%s", namespace, deployment)
}

// Gets replicas of a deployment and checks its state in Redis
func getReplicas(kClient kubernetes.Interface, rClient *redis.Client, namespace string, deployment string) (*getReplicasResponse, error) {
	defer e.NonFatal()

	// Get the deployment and replicas
	deployResp, err := kClient.AppsV1().Deployments(namespace).Get(context.TODO(), deployment, metav1.GetOptions{})
	// Catch k8s API specific errors
	if err != nil {
		if statusError, isStatus := err.(*errors.StatusError); isStatus {
			return nil, statusError
		} else {
			return nil, err
		}
	}

	// Get the redis key to set the state
	redisKey := genRedisKey(namespace, deployment)
	redisGetValue, keyExists, err := getState(rClient, redisKey)
	if err != nil {
		logger.Log.Errorf("error getting state for key %s from Redis: %s", redisKey, err)
		return nil, err
	}

	var redisSetValue *redisValue
	// Logic handling if this is the first time we've seen this deployment
	if keyExists {
		if redisGetValue.DesiredReplicas != *deployResp.Spec.Replicas {
			logger.Log.Debugf("difference detected for deployment %s, k8s_replicas:%d, redis_replicas:%d", deployment, *deployResp.Spec.Replicas, redisGetValue.DesiredReplicas)
			redisSetValue = &redisValue{DesiredReplicas: redisGetValue.DesiredReplicas, CurrentReplicas: *deployResp.Spec.Replicas, Drift: true}
		} else {
			// No need to set the key again if there is no drift, just return the current values
			logger.Log.Debugf("desired replicas for %s match, returning k8s + redis data and not setting anything in Redis", redisKey)
			resp := &getReplicasResponse{Code: 200, Namespace: namespace, Deployment: deployment, CurrentReplicas: *deployResp.Spec.Replicas, DesiredReplicas: redisGetValue.DesiredReplicas, Drift: redisGetValue.Drift}
			return resp, nil
		}
	} else {
		redisSetValue = &redisValue{DesiredReplicas: redisGetValue.DesiredReplicas, CurrentReplicas: *deployResp.Spec.Replicas, Drift: redisGetValue.Drift}
	}

	// Sends the update values to the Redis function
	_, err = setState(rClient, redisKey, redisSetValue, *deployResp.Spec.Replicas)
	if err != nil {
		logger.Log.Errorf("error setting state for key %s in Redis: %s", redisKey, err)
		return nil, err
	}

	resp := &getReplicasResponse{Code: 200, Namespace: namespace, Deployment: deployment, CurrentReplicas: *deployResp.Spec.Replicas,
		DesiredReplicas: redisSetValue.DesiredReplicas, Drift: redisSetValue.Drift}

	return resp, nil
}

// Sets the replicas of a deployment and stores its state in Redis
func setReplicas(kClient kubernetes.Interface, rClient *redis.Client, namespace string, deployment string, replicas int32) (*setReplicasResponse, error) {
	defer e.NonFatal()

	// Get the deployment and replicas for the current state
	deployResp, err := kClient.AppsV1().Deployments(namespace).Get(context.TODO(), deployment, metav1.GetOptions{})
	// Catch k8s API specific errors
	if err != nil {
		if statusError, isStatus := err.(*errors.StatusError); isStatus {
			return nil, statusError
		} else {
			return nil, err
		}
	}

	// Calls the k8s API and uses a PATCH to update the replicas of the deployment
	patchReplicas := []byte(fmt.Sprintf(`{"spec":{"replicas": %d}}`, replicas))
	_, err = kClient.AppsV1().Deployments(namespace).Patch(context.TODO(), deployment, types.MergePatchType, patchReplicas, metav1.PatchOptions{})
	// Catch k8s API specific errors
	if err != nil {
		if statusError, isStatus := err.(*errors.StatusError); isStatus {
			return nil, statusError
		} else {
			return nil, err
		}
	}

	// Gets the deployment key from Redis
	redisKey := genRedisKey(namespace, deployment)
	redisGetValue, _, err := getState(rClient, redisKey)
	if err != nil {
		logger.Log.Errorf("error getting state for key %s from Redis: %s", redisKey, err)
		return nil, err
	}

	redisSetValue := &redisValue{DesiredReplicas: replicas, CurrentReplicas: replicas, Drift: false}

	// Sets the redis key with updated values
	_, err = setState(rClient, redisKey, redisSetValue, replicas)
	if err != nil {
		logger.Log.Errorf("error setting state for key %s in Redis: %s", redisKey, err)
		return nil, err
	}

	resp := &setReplicasResponse{Code: 200, Namespace: namespace, Deployment: deployment, DesiredReplicas: redisGetValue.DesiredReplicas,
		RequestedReplicas: replicas, CurrentReplicas: *deployResp.Spec.Replicas, Drift: redisSetValue.Drift}

	return resp, nil
}

// Gets State via the value of the Redis key
func getState(rClient *redis.Client, redisKey string) (*redisValue, bool, error) {
	var redisGetValues *redisValue

	// Context for Redis connections
	rCtx := context.Background()

	// Check if the key exists, if not return back with false
	keyExists := rClient.Exists(rCtx, redisKey).Val()
	if keyExists == 0 {
		logger.Log.Debugf("key %s does not exist in Redis, returning false", redisKey)
		return &redisValue{0, 0, true}, false, nil
	}

	// Gets the existing key in Redis
	rGet, err := rClient.Get(rCtx, redisKey).Result()
	if err != nil {
		logger.Log.Errorf("error getting key %s from Redis: %s", redisKey, err)
		return nil, false, err
	}

	logger.Log.Debugf("key %s found in Redis", redisKey)

	// Unmarshal the value from Redis into the redisValue struct
	err = json.Unmarshal([]byte(rGet), &redisGetValues)
	if err != nil {
		logger.Log.Errorf("error unmarshalling key %s from Redis: %s", redisKey, err)
		return nil, false, err
	}

	return redisGetValues, true, nil
}

// Sets the state in Redis and passes in a redisValue for reference
func setState(rClient *redis.Client, redisKey string, redisNewValue *redisValue, replicas int32) (*redisValue, error) {
	redisSetValues := &redisValue{DesiredReplicas: redisNewValue.DesiredReplicas,
		CurrentReplicas: replicas, Drift: redisNewValue.Drift}

	// Context for Redis connections
	rCtx := context.Background()

	logger.Log.Debugf("setting key %s in Redis", redisKey)
	// Marshals the values to JSON
	redisJson, err := json.Marshal(redisSetValues)
	if err != nil {
		logger.Log.Errorf("error marshalling key %s to Redis: %s", redisKey, err)
		return nil, err
	}

	// Sets the new values in Redis
	_, err = rClient.Set(rCtx, redisKey, redisJson, 0).Result()
	if err != nil {
		logger.Log.Errorf("error setting key %s in Redis: %s", redisKey, err)
		return nil, err
	}

	return redisSetValues, nil
}
