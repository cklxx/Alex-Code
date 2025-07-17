package protocol

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCRequest(t *testing.T) {
	req := NewRequest(1, "test_method", map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	})

	// Test serialization
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Test deserialization
	var decoded JSONRPCRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.JSONRPC != JSONRPCVersion {
		t.Errorf("Expected JSONRPC version %s, got %s", JSONRPCVersion, decoded.JSONRPC)
	}

	if decoded.Method != "test_method" {
		t.Errorf("Expected method 'test_method', got %s", decoded.Method)
	}

	if decoded.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", decoded.ID)
	}
}

func TestJSONRPCResponse(t *testing.T) {
	resp := NewResponse(1, map[string]interface{}{
		"result": "success",
	})

	// Test serialization
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Test deserialization
	var decoded JSONRPCResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if decoded.JSONRPC != JSONRPCVersion {
		t.Errorf("Expected JSONRPC version %s, got %s", JSONRPCVersion, decoded.JSONRPC)
	}

	if decoded.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", decoded.ID)
	}

	if decoded.Error != nil {
		t.Errorf("Expected no error, got %v", decoded.Error)
	}
}

func TestJSONRPCErrorResponse(t *testing.T) {
	resp := NewErrorResponse(1, InvalidParams, "Invalid parameters", nil)

	// Test serialization
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	// Test deserialization
	var decoded JSONRPCResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if decoded.JSONRPC != JSONRPCVersion {
		t.Errorf("Expected JSONRPC version %s, got %s", JSONRPCVersion, decoded.JSONRPC)
	}

	if decoded.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", decoded.ID)
	}

	if decoded.Error == nil {
		t.Fatal("Expected error, got nil")
	}

	if decoded.Error.Code != InvalidParams {
		t.Errorf("Expected error code %d, got %d", InvalidParams, decoded.Error.Code)
	}

	if decoded.Error.Message != "Invalid parameters" {
		t.Errorf("Expected error message 'Invalid parameters', got %s", decoded.Error.Message)
	}
}

func TestJSONRPCNotification(t *testing.T) {
	notif := NewNotification("test_notification", map[string]interface{}{
		"event": "test_event",
	})

	// Test serialization
	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	// Test deserialization
	var decoded JSONRPCNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal notification: %v", err)
	}

	if decoded.JSONRPC != JSONRPCVersion {
		t.Errorf("Expected JSONRPC version %s, got %s", JSONRPCVersion, decoded.JSONRPC)
	}

	if decoded.Method != "test_notification" {
		t.Errorf("Expected method 'test_notification', got %s", decoded.Method)
	}
}

func TestMessageTypeDetection(t *testing.T) {
	// Test request detection
	reqData := []byte(`{"jsonrpc": "2.0", "id": 1, "method": "test"}`)
	if !IsRequest(reqData) {
		t.Error("Expected IsRequest to return true for request")
	}
	if IsResponse(reqData) {
		t.Error("Expected IsResponse to return false for request")
	}
	if IsNotification(reqData) {
		t.Error("Expected IsNotification to return false for request")
	}

	// Test response detection
	respData := []byte(`{"jsonrpc": "2.0", "id": 1, "result": "success"}`)
	if IsRequest(respData) {
		t.Error("Expected IsRequest to return false for response")
	}
	if !IsResponse(respData) {
		t.Error("Expected IsResponse to return true for response")
	}
	if IsNotification(respData) {
		t.Error("Expected IsNotification to return false for response")
	}

	// Test notification detection
	notifData := []byte(`{"jsonrpc": "2.0", "method": "test_notification"}`)
	if IsRequest(notifData) {
		t.Error("Expected IsRequest to return false for notification")
	}
	if IsResponse(notifData) {
		t.Error("Expected IsResponse to return false for notification")
	}
	if !IsNotification(notifData) {
		t.Error("Expected IsNotification to return true for notification")
	}
}

func TestJSONRPCError(t *testing.T) {
	err := &JSONRPCError{
		Code:    InvalidRequest,
		Message: "Invalid request",
		Data:    "Additional error data",
	}

	expected := "JSON-RPC Error -32600: Invalid request"
	if err.Error() != expected {
		t.Errorf("Expected error string %s, got %s", expected, err.Error())
	}
}