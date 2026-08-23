package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hall-petch/internal/core"
)

func TestHandleSyValid(t *testing.T) {
	body := `{"sigma0":50,"ky":0.5,"d":10,"d_unit":"um"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sy", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handleSy(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var out syResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.OK {
		t.Fatal("ok must be true")
	}
	// d = 10 um = 1e-5 m, d^{-1/2} = 316.227..., sigma_y = 50 + 0.5*316.227
	want := 50.0 + 0.5*316.227766
	if out.SigmaY < want-0.01 || out.SigmaY > want+0.01 {
		t.Fatalf("sigma_y = %v, want ~%v", out.SigmaY, want)
	}
	if out.DSqrtInv < 316.0 || out.DSqrtInv > 316.3 {
		t.Fatalf("d_sqrt_inv = %v", out.DSqrtInv)
	}
}

func TestHandleSyBadMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/sy", nil)
	rec := httptest.NewRecorder()
	handleSy(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestHandleSyInvalidDiameter(t *testing.T) {
	body := `{"sigma0":50,"ky":0.5,"d":0,"d_unit":"um"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sy", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handleSy(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	var out errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.OK {
		t.Fatal("ok must be false")
	}
	if out.Error == "" {
		t.Fatal("error message expected")
	}
}

func TestHandleSyBadUnit(t *testing.T) {
	body := `{"sigma0":50,"ky":0.5,"d":10,"d_unit":"furlong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sy", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handleSy(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestServerRouting(t *testing.T) {
	s := NewServer("web", "example")
	if s.Handler() == nil {
		t.Fatal("handler must be non-nil")
	}
	// core sanity: a 4x diameter halves the extra term.
	res1, err := core.Params{Sigma0: 100, Ky: 1.0, D: 1, DUnit: "um"}.Evaluate()
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	res4, err := core.Params{Sigma0: 100, Ky: 1.0, D: 4, DUnit: "um"}.Evaluate()
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if res4.SigmaY-res1.Sigma0 >= res1.SigmaY-res1.Sigma0 {
		t.Fatal("4x diameter must halve the strengthening term")
	}
}
