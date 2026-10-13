package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestOKEnvelope(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	OK(c, map[string]string{"id": "1"}, "done")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !body.Success {
		t.Error("success must be true")
	}
	if body.Message != "done" {
		t.Errorf("message = %q, want done", body.Message)
	}
}

func TestErrorEnvelope(t *testing.T) {
	for _, tc := range []struct {
		fn       func(*gin.Context, string, string)
		expected int
		code     string
	}{
		{BadRequest, http.StatusBadRequest, "BAD"},
		{Unauthorized, http.StatusUnauthorized, "UNAUTH"},
		{Forbidden, http.StatusForbidden, "FORBIDDEN"},
		{NotFound, http.StatusNotFound, "NOT_FOUND"},
		{Conflict, http.StatusConflict, "CONFLICT"},
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		tc.fn(c, tc.code, "msg")

		if w.Code != tc.expected {
			t.Errorf("%s status = %d, want %d", tc.code, w.Code, tc.expected)
		}

		var body APIErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if body.Success {
			t.Error("success must be false on error")
		}
		if body.Error.Code != tc.code {
			t.Errorf("code = %q, want %q", body.Error.Code, tc.code)
		}
	}
}

func TestInternalServerError_DoesNotLeak(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Passing a raw error string as a message must not include err.Error().
	InternalServerError(c, "DB_ERROR", "An internal error occurred. Please try again later.")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	if got := w.Body.String(); !contains(got, "An internal error occurred") {
		t.Errorf("expected generic message, got %s", got)
	}
}

func TestEmptyCodeGetsDefault(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	BadRequest(c, "", "msg")

	var body APIErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Code != "BAD_REQUEST" {
		t.Errorf("empty code should default to BAD_REQUEST, got %q", body.Error.Code)
	}
}

func TestOKDefaultsMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	OK(c, nil, "")

	var body APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Message != "Success" {
		t.Errorf("empty message should default to Success, got %q", body.Message)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
