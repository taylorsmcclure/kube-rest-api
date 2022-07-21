package replicas

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/taylorsmcclure/kube-server/internal/logger"

	"github.com/go-redis/redismock/v8"
)

// I don't like being dependent on the internal package, but
// this causes a nil pointer exception if it isn't initialized
func init() {
	logger.Setup(false)
}

// TODO: I would like to test the HTTP routing for /replicas, but I cannot
// find an easy way to mock the redis client with two different states
// and send it to the V1Replicas function.

// Tests getting a deployment state object from Redis
func TestGetState(t *testing.T) {
	testCases := []struct {
		name             string
		description      string
		expectSuccess    bool
		redisKey         string
		keyExists        bool
		expectedResponse redisValue
	}{
		{
			name:          "replicas-first-get",
			description:   "This is the first time the server has seen the deployment, so it will return 0,0,false",
			redisKey:      "namespace-first-replicas-deployment",
			expectSuccess: true,
			keyExists:     false,
			expectedResponse: redisValue{
				DesiredReplicas: 0,
				CurrentReplicas: 0,
				Drift:           true,
			},
		},

		{
			name:          "replicas-get-no-drift",
			description:   "This key exists in Redis with a value",
			redisKey:      "namespace-replicas-deployment",
			expectSuccess: true,
			keyExists:     true,
			expectedResponse: redisValue{
				DesiredReplicas: 4,
				CurrentReplicas: 4,
				Drift:           false,
			},
		},
		{
			name:          "replicas-get-with-drift",
			description:   "This key exists in Redis with a value and has drift",
			redisKey:      "namespace-drift-replicas-deployment",
			expectSuccess: true,
			keyExists:     true,
			expectedResponse: redisValue{
				DesiredReplicas: 4,
				CurrentReplicas: 2,
				Drift:           true,
			},
		},
		{
			name:          "replicas-get-drift-incorrect",
			description:   "This key exists in Redis with a value and has drift but is not marked as having drift",
			redisKey:      "namespace-drift-replicas-deployment-false",
			expectSuccess: true,
			keyExists:     true,
			expectedResponse: redisValue{
				DesiredReplicas: 4,
				CurrentReplicas: 2,
				Drift:           false,
			},
		},
	}

	// Loop to run each test case
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			// Init mock Redis client
			db, mock := redismock.NewClientMock()

			// Handle the key exists or not and the appropriate Redis response
			if !test.keyExists {
				mock.ExpectExists(test.redisKey).RedisNil()
			} else {
				mock.ExpectExists(test.redisKey).SetVal(1)
			}

			// To match we are marshalling to JSON as well
			rGetJson, err := json.Marshal(test.expectedResponse)
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectGet(test.redisKey).SetVal(string(rGetJson))

			// Run the function with the mock client and stubbed data
			testResp, key, err := getState(db, test.redisKey)
			if err != nil {
				t.Fatal(err)
			}

			// Test the input against the output of the function
			// Also test if the correct values were returned if the key does not exist
			switch {
			case test.expectSuccess && !reflect.DeepEqual(testResp, &test.expectedResponse):
				t.Errorf("Fail: got %v want %v",
					testResp, &test.expectedResponse)
			case test.keyExists && !key:
				t.Errorf("Redis key was present but: got %v want %v", key, test.keyExists)
			default:
				t.Logf("test passed %v", testResp)
			}

		},
		)
	}

}

// Test for setting the state via Redis
func TestSetState(t *testing.T) {
	testCases := []struct {
		name             string
		description      string
		expectSuccess    bool
		redisKey         string
		replicas         int32
		expectedResponse redisValue
	}{
		{
			name:          "replicas-first-set",
			description:   "This is the first time the server will set a key",
			redisKey:      "namespace-first-replicas-deployment",
			expectSuccess: true,
			replicas:      4,
			expectedResponse: redisValue{
				DesiredReplicas: 4,
				CurrentReplicas: 4,
				Drift:           true,
			},
		},

		{
			name:          "replicas-scale-down",
			description:   "This will update a key with a new desired and current replicas value",
			redisKey:      "namespace-replicas-scale-down",
			expectSuccess: true,
			replicas:      2,
			expectedResponse: redisValue{
				DesiredReplicas: 2,
				CurrentReplicas: 2,
				Drift:           false,
			},
		},

		{
			name:          "replicas-scale-up",
			description:   "This will update a key with a new desired and current replicas value",
			redisKey:      "namespace-replicas-scale-up",
			expectSuccess: true,
			replicas:      6,
			expectedResponse: redisValue{
				DesiredReplicas: 6,
				CurrentReplicas: 6,
				Drift:           false,
			},
		},
	}

	// Loop to run each test case
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			// Init mock Redis client
			db, mock := redismock.NewClientMock()

			var rSetValues redisValue
			if !test.expectSuccess {
				rSetValues = redisValue{DesiredReplicas: test.replicas, CurrentReplicas: test.replicas, Drift: false}
			} else {
				rSetValues = test.expectedResponse
			}

			// To match we are marshalling to JSON as well
			rSetJson, err := json.Marshal(rSetValues)
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectSet(test.redisKey, rSetJson, 0).SetVal("")

			// Run the function with the mock client and stubbed data
			testResp, err := setState(db, test.redisKey, &test.expectedResponse, test.replicas)
			if err != nil {
				t.Fatal(err)
			}

			// Test the input against the output of the function
			switch {
			case test.expectSuccess && !reflect.DeepEqual(testResp, &test.expectedResponse):
				t.Errorf("Fail: got %v want %v",
					testResp, &test.expectedResponse)
			case !test.expectSuccess && reflect.DeepEqual(testResp, &test.expectedResponse):
				t.Errorf("Fail: got %v want %v", testResp, test.expectSuccess)
			default:
				t.Logf("test passed %v", testResp)
			}

		},
		)
	}

}
