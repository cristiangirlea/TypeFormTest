package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"typeform-test/models"
	"typeform-test/store"
)

func TestHealthHandler(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s, nil)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.Health)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"status":"ok"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestCreateForm(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s, nil)

	payload := `{"title": "My New Form"}`
	req, _ := http.NewRequest("POST", "/forms", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	http.HandlerFunc(h.CreateForm).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["id"] == "" {
		t.Error("expected non-empty id")
	}
}

func TestAddQuestion(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /forms/{id}/questions", h.AddQuestion)

	f := &models.Form{ID: "test-id", Title: "Test Form"}
	s.CreateForm(t.Context(), f)

	payload := `{"text": "What is your name?", "type": "text"}`
	req, _ := http.NewRequest("POST", "/forms/test-id/questions", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var q models.Question
	json.Unmarshal(rr.Body.Bytes(), &q)
	if q.Text != "What is your name?" {
		t.Errorf("expected question text 'What is your name?', got '%s'", q.Text)
	}
}

func TestSaveForm(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /forms/{id}/save", h.SaveForm)

	f := &models.Form{
		ID:    "test-id",
		Title: "Test Form",
		Questions: []models.Question{
			{ID: "q1", Text: "Q1", Type: "text"},
		},
	}
	s.CreateForm(t.Context(), f)

	req, _ := http.NewRequest("POST", "/forms/test-id/save", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var savedForm models.Form
	json.Unmarshal(rr.Body.Bytes(), &savedForm)
	if savedForm.ShareSlug == "" {
		t.Error("expected non-empty share slug")
	}
}

func TestSaveFormNoQuestions(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /forms/{id}/save", h.SaveForm)

	f := &models.Form{ID: "empty-id", Title: "Empty Form", Questions: []models.Question{}}
	s.CreateForm(t.Context(), f)

	req, _ := http.NewRequest("POST", "/forms/empty-id/save", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestListForms(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /forms", h.ListForms)

	s.CreateForm(t.Context(), &models.Form{ID: "f1", Title: "Form 1"})
	s.CreateForm(t.Context(), &models.Form{ID: "f2", Title: "Form 2"})

	req, _ := http.NewRequest("GET", "/forms", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var forms []models.Form
	json.Unmarshal(rr.Body.Bytes(), &forms)
	if len(forms) != 2 {
		t.Errorf("expected 2 forms, got %d", len(forms))
	}
}

func TestGetFormBySlug(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /form/{slug}", h.GetFormBySlug)

	s.CreateForm(t.Context(), &models.Form{ID: "f1", Title: "Form 1", ShareSlug: "myslug"})

	req, _ := http.NewRequest("GET", "/form/myslug", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var form models.Form
	json.Unmarshal(rr.Body.Bytes(), &form)
	if form.ID != "f1" {
		t.Errorf("expected form ID f1, got %s", form.ID)
	}
}

func TestSubmitResponse(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /form/{slug}/responses", h.SubmitResponse)

	s.CreateForm(t.Context(), &models.Form{ID: "f1", ShareSlug: "myslug"})

	payload := `{"answers": [{"questionID": "q1", "value": "John"}]}`
	req, _ := http.NewRequest("POST", "/form/myslug/responses", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}
