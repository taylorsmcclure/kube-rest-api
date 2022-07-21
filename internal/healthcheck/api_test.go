package healthcheck

import (
	"testing"

	"github.com/taylorsmcclure/kube-server/internal/logger"
)

// I don't like being dependent on the internal package, but
// this causes a nil pointer exception if it isn't initialized
func init() {
	logger.Setup(false)
}

// Tests the /v1/healthz endpoint
func TestV1HealthCheck(t *testing.T) {
	t.Log("TODO: Implement healthcheck endpoint testing")
}
