package sdk

import (
	"context"
	"encoding/json"
	"net/http"
)

// HealthStatus mirrors the response from /healthz endpoint.
type HealthStatus struct {
	Status string `json:"status"` // Status is typically "ok" when the service is healthy
}

// HealthCheck queries the /healthz endpoint to check service health.
//
// This is useful for monitoring and health checks. Returns the health status
// of the catalog service.
//
// Example:
//
//	status, err := client.HealthCheck(ctx)
//	if err != nil {
//		return err
//	}
//	if status.Status == "ok" {
//		fmt.Println("Service is healthy")
//	}
func (c *RawClient) HealthCheck(ctx context.Context, opts ...CallOption) (*HealthStatus, error) {
	callOpts := newCallOptions(opts...)
	resp, err := c.doRaw(ctx, http.MethodGet, "/healthz", nil, callOpts, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var status HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	return &status, nil
}
