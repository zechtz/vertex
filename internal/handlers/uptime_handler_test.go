package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/zechtz/vertex/internal/models"
)

// ServiceManagerInterface for testing
type ServiceManagerInterface interface {
	GetServices() []models.Service
	GetServiceByUUID(uuid string) (*models.Service, bool)
}

// MockServiceManager for testing
type MockServiceManager struct{}

func (msm *MockServiceManager) GetServices() []models.Service {
	return []models.Service{
		{
			ID:           "test-service-1",
			Name:         "Test Service 1",
			Port:         8080,
			Status:       "running",
			HealthStatus: "healthy",
		},
		{
			ID:           "test-service-2",
			Name:         "Test Service 2",
			Port:         8081,
			Status:       "stopped",
			HealthStatus: "unknown",
		},
	}
}

func (msm *MockServiceManager) GetServiceByUUID(uuid string) (*models.Service, bool) {
	if uuid == "test-service-1" {
		return &models.Service{
			ID:           "test-service-1",
			Name:         "Test Service 1",
			Port:         8080,
			Status:       "running",
			HealthStatus: "healthy",
		}, true
	}
	return nil, false
}

// MockHandler for testing
type MockHandler struct {
	serviceManager ServiceManagerInterface
}

func (h *MockHandler) getUptimeStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	services := h.serviceManager.GetServices()

	serviceStats := make(map[string]interface{})
	for _, service := range services {
		serviceStats[service.ID] = map[string]interface{}{
			"serviceName":  service.Name,
			"serviceId":    service.ID,
			"port":         service.Port,
			"status":       service.Status,
			"healthStatus": service.HealthStatus,
			"stats": map[string]interface{}{
				"totalRestarts":       0,
				"uptimePercentage24h": 100.0,
				"uptimePercentage7d":  100.0,
				"mtbf":                0,
				"lastDowntime":        nil,
				"totalDowntime24h":    0,
				"totalDowntime7d":     0,
			},
		}
	}

	response := map[string]interface{}{
		"statistics": serviceStats,
		"summary": map[string]interface{}{
			"totalServices":     len(services),
			"runningServices":   countRunningServices(services),
			"unhealthyServices": countUnhealthyServices(services),
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (h *MockHandler) getServiceUptimeStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	serviceID := vars["id"]

	service, exists := h.serviceManager.GetServiceByUUID(serviceID)
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"serviceName":  service.Name,
		"serviceId":    service.ID,
		"port":         service.Port,
		"status":       service.Status,
		"healthStatus": service.HealthStatus,
		"stats": map[string]interface{}{
			"totalRestarts":       0,
			"uptimePercentage24h": 100.0,
			"uptimePercentage7d":  100.0,
			"mtbf":                0,
			"lastDowntime":        nil,
			"totalDowntime24h":    0,
			"totalDowntime7d":     0,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func TestGetUptimeStatisticsHandler(t *testing.T) {
	// Create a mock handler
	mockServiceManager := &MockServiceManager{}
	handler := &MockHandler{
		serviceManager: mockServiceManager,
	}

	// Create request
	req, err := http.NewRequest("GET", "/api/uptime/statistics", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.getUptimeStatisticsHandler(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("could not parse response: %v", err)
	}

	// Verify response structure
	if _, exists := response["statistics"]; !exists {
		t.Error("response missing statistics field")
	}

	if _, exists := response["summary"]; !exists {
		t.Error("response missing summary field")
	}

	// Check summary
	summary := response["summary"].(map[string]interface{})
	if summary["totalServices"] != float64(2) {
		t.Errorf("expected 2 total services, got %v", summary["totalServices"])
	}
}

func TestGetServiceUptimeStatisticsHandler(t *testing.T) {
	// Create a mock handler
	mockServiceManager := &MockServiceManager{}
	handler := &MockHandler{
		serviceManager: mockServiceManager,
	}

	// Create request with service ID
	req, err := http.NewRequest("GET", "/api/uptime/statistics/test-service-1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add URL vars (simulate mux router)
	req = mux.SetURLVars(req, map[string]string{"id": "test-service-1"})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.getServiceUptimeStatisticsHandler(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("could not parse response: %v", err)
	}

	// Verify response structure
	if response["serviceName"] != "Test Service 1" {
		t.Errorf("expected service name 'Test Service 1', got %v", response["serviceName"])
	}

	if response["serviceId"] != "test-service-1" {
		t.Errorf("expected service ID 'test-service-1', got %v", response["serviceId"])
	}
}

func TestGetServiceUptimeStatisticsHandler_NotFound(t *testing.T) {
	// Create a mock handler
	mockServiceManager := &MockServiceManager{}
	handler := &MockHandler{
		serviceManager: mockServiceManager,
	}

	// Create request with non-existent service ID
	req, err := http.NewRequest("GET", "/api/uptime/statistics/non-existent", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add URL vars (simulate mux router)
	req = mux.SetURLVars(req, map[string]string{"id": "non-existent"})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.getServiceUptimeStatisticsHandler(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}
