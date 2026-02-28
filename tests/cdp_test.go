
package tests

import (
	"encoding/json"
	"testing"

	"github.com/kexas-project/kexas/internal/cdp"
)

func TestResponseError_Error(t *testing.T) {
	var err *cdp.ResponseError = &cdp.ResponseError{
		Code:    -32601,
		Message: "Method not found",
	}

	var expected string = "cdp error -32601: Method not found"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestRequest_MarshalJSON(t *testing.T) {
	var req *cdp.Request = &cdp.Request{
		ID:     123,
		Method: "Page.navigate",
		Params: map[string]interface{}{
			"url": "https://example.com",
		},
	}

	var data []byte
	var err error
	data, err = json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var unmarshaled cdp.Request
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if unmarshaled.ID != 123 {
		t.Errorf("expected ID 123, got %d", unmarshaled.ID)
	}
	if unmarshaled.Method != "Page.navigate" {
		t.Errorf("expected method 'Page.navigate', got '%s'", unmarshaled.Method)
	}
	if unmarshaled.Params["url"] != "https://example.com" {
		t.Error("expected url parameter")
	}
}

func TestResponse_UnmarshalJSON_Success(t *testing.T) {
	var data []byte = []byte(`{"id":1,"result":{"frameId":"ABC123"}}`)

	var resp cdp.Response
	var err error = json.Unmarshal(data, &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != 1 {
		t.Errorf("expected ID 1, got %d", resp.ID)
	}
	if resp.Result["frameId"] != "ABC123" {
		t.Error("expected frameId in result")
	}
	if resp.Error != nil {
		t.Error("expected no error in response")
	}
}

func TestResponse_UnmarshalJSON_Error(t *testing.T) {
	var data []byte = []byte(`{"id":2,"error":{"code":-32601,"message":"Method not found"}}`)

	var resp cdp.Response
	var err error = json.Unmarshal(data, &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != 2 {
		t.Errorf("expected ID 2, got %d", resp.ID)
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", resp.Error.Code)
	}
	if resp.Error.Message != "Method not found" {
		t.Errorf("expected error message 'Method not found', got '%s'", resp.Error.Message)
	}
}

func TestEvent_UnmarshalJSON(t *testing.T) {
	var data []byte = []byte(`{"method":"Page.loadEventFired","params":{"timestamp":123456.789}}`)

	var event cdp.Event
	var err error = json.Unmarshal(data, &event)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	if event.Method != "Page.loadEventFired" {
		t.Errorf("expected method 'Page.loadEventFired', got '%s'", event.Method)
	}
	if event.Params["timestamp"] != 123456.789 {
		t.Error("expected timestamp in params")
	}
}
