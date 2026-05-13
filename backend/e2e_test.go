package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"typeform-test/handlers"
	"typeform-test/models"
	"typeform-test/store"
)

func setupE2ETestServer() *httptest.Server {
	s := store.NewMemoryStore()
	h := handlers.NewHandler(s, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /forms", h.CreateForm)
	mux.HandleFunc("POST /forms/{id}/questions", h.AddQuestion)
	mux.HandleFunc("POST /forms/{id}/save", h.SaveForm)
	mux.HandleFunc("GET /forms", h.ListForms)
	mux.HandleFunc("GET /forms/{id}", h.GetForm)
	mux.HandleFunc("GET /form/{slug}", h.GetFormBySlug)
	mux.HandleFunc("POST /form/{slug}/responses", h.SubmitResponse)

	return httptest.NewServer(mux)
}

func TestE2EFormLifecycle(t *testing.T) {
	server := setupE2ETestServer()
	defer server.Close()

	client := server.Client()

	t.Run("health check", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("health request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode health response: %v", err)
		}

		if body["status"] != "ok" {
			t.Fatalf("expected status ok, got %q", body["status"])
		}
	})

	var createdForm models.Form

	t.Run("create form", func(t *testing.T) {
		payload := []byte(`{"title":"Customer Feedback"}`)

		resp, err := client.Post(
			server.URL+"/forms",
			"application/json",
			bytes.NewReader(payload),
		)
		if err != nil {
			t.Fatalf("create form request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&createdForm); err != nil {
			t.Fatalf("failed to decode create form response: %v", err)
		}

		if createdForm.ID == "" {
			t.Fatal("expected created form ID to be set")
		}

		if createdForm.Title != "Customer Feedback" {
			t.Fatalf("expected title %q, got %q", "Customer Feedback", createdForm.Title)
		}

		if len(createdForm.Questions) != 0 {
			t.Fatalf("expected no questions on newly created form, got %d", len(createdForm.Questions))
		}
	})

	t.Run("saving form without questions fails", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodPost,
			server.URL+"/forms/"+createdForm.ID+"/save",
			nil,
		)
		if err != nil {
			t.Fatalf("failed to build save request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("save empty form request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	var addedQuestion models.Question

	t.Run("add question", func(t *testing.T) {
		payload := []byte(`{"text":"What is your name?","type":"text"}`)

		resp, err := client.Post(
			server.URL+"/forms/"+createdForm.ID+"/questions",
			"application/json",
			bytes.NewReader(payload),
		)
		if err != nil {
			t.Fatalf("add question request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&addedQuestion); err != nil {
			t.Fatalf("failed to decode add question response: %v", err)
		}

		if addedQuestion.ID == "" {
			t.Fatal("expected question ID to be set")
		}

		if addedQuestion.Text != "What is your name?" {
			t.Fatalf("expected question text %q, got %q", "What is your name?", addedQuestion.Text)
		}

		if addedQuestion.Type != models.QuestionTypeText {
			t.Fatalf("expected question type %q, got %q", models.QuestionTypeText, addedQuestion.Type)
		}
	})

	var savedForm models.Form

	t.Run("save form generates share slug", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodPost,
			server.URL+"/forms/"+createdForm.ID+"/save",
			nil,
		)
		if err != nil {
			t.Fatalf("failed to build save request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("save form request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&savedForm); err != nil {
			t.Fatalf("failed to decode save form response: %v", err)
		}

		if savedForm.ID != createdForm.ID {
			t.Fatalf("expected saved form ID %q, got %q", createdForm.ID, savedForm.ID)
		}

		if savedForm.ShareSlug == "" {
			t.Fatal("expected share slug to be generated")
		}

		if len(savedForm.Questions) != 1 {
			t.Fatalf("expected 1 question, got %d", len(savedForm.Questions))
		}
	})

	t.Run("get form by ID", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/forms/" + createdForm.ID)
		if err != nil {
			t.Fatalf("get form request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var form models.Form
		if err := json.NewDecoder(resp.Body).Decode(&form); err != nil {
			t.Fatalf("failed to decode get form response: %v", err)
		}

		if form.ID != createdForm.ID {
			t.Fatalf("expected form ID %q, got %q", createdForm.ID, form.ID)
		}

		if form.ShareSlug != savedForm.ShareSlug {
			t.Fatalf("expected share slug %q, got %q", savedForm.ShareSlug, form.ShareSlug)
		}
	})

	t.Run("list forms", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/forms")
		if err != nil {
			t.Fatalf("list forms request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var forms []models.Form
		if err := json.NewDecoder(resp.Body).Decode(&forms); err != nil {
			t.Fatalf("failed to decode list forms response: %v", err)
		}

		if len(forms) != 1 {
			t.Fatalf("expected 1 form, got %d", len(forms))
		}

		if forms[0].ID != createdForm.ID {
			t.Fatalf("expected form ID %q, got %q", createdForm.ID, forms[0].ID)
		}
	})

	t.Run("get form by share slug", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/form/" + savedForm.ShareSlug)
		if err != nil {
			t.Fatalf("get form by slug request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var form models.Form
		if err := json.NewDecoder(resp.Body).Decode(&form); err != nil {
			t.Fatalf("failed to decode get form by slug response: %v", err)
		}

		if form.ID != createdForm.ID {
			t.Fatalf("expected form ID %q, got %q", createdForm.ID, form.ID)
		}
	})

	t.Run("submit response", func(t *testing.T) {
		payload := []byte(`{
			"answers": [
				{
					"questionID": "` + addedQuestion.ID + `",
					"value": "John Doe"
				}
			]
		}`)

		resp, err := client.Post(
			server.URL+"/form/"+savedForm.ShareSlug+"/responses",
			"application/json",
			bytes.NewReader(payload),
		)
		if err != nil {
			t.Fatalf("submit response request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
		}

		var submittedResponse models.Response
		if err := json.NewDecoder(resp.Body).Decode(&submittedResponse); err != nil {
			t.Fatalf("failed to decode submit response body: %v", err)
		}

		if submittedResponse.ID == "" {
			t.Fatal("expected response ID to be set")
		}

		if submittedResponse.FormID != createdForm.ID {
			t.Fatalf("expected response form ID %q, got %q", createdForm.ID, submittedResponse.FormID)
		}

		if len(submittedResponse.Answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(submittedResponse.Answers))
		}

		answer := submittedResponse.Answers[0]

		if answer.QuestionID != addedQuestion.ID {
			t.Fatalf("expected question ID %q, got %q", addedQuestion.ID, answer.QuestionID)
		}

		if answer.Value != "John Doe" {
			t.Fatalf("expected answer value %q, got %q", "John Doe", answer.Value)
		}
	})
}

func TestE2ENotFoundScenarios(t *testing.T) {
	server := setupE2ETestServer()
	defer server.Close()

	client := server.Client()

	tests := []struct {
		name   string
		method string
		path   string
		body   []byte
	}{
		{
			name:   "get missing form by id",
			method: http.MethodGet,
			path:   "/forms/missing-id",
		},
		{
			name:   "add question to missing form",
			method: http.MethodPost,
			path:   "/forms/missing-id/questions",
			body:   []byte(`{"text":"Missing?","type":"text"}`),
		},
		{
			name:   "save missing form",
			method: http.MethodPost,
			path:   "/forms/missing-id/save",
		},
		{
			name:   "get missing form by slug",
			method: http.MethodGet,
			path:   "/form/missing-slug",
		},
		{
			name:   "submit response to missing form slug",
			method: http.MethodPost,
			path:   "/form/missing-slug/responses",
			body:   []byte(`{"answers":[{"questionID":"q1","value":"answer"}]}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Reader
			if tt.body != nil {
				body = bytes.NewReader(tt.body)
			} else {
				body = bytes.NewReader(nil)
			}

			req, err := http.NewRequest(tt.method, server.URL+tt.path, body)
			if err != nil {
				t.Fatalf("failed to build request: %v", err)
			}

			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
			}
		})
	}
}

func TestE2EBadJSONScenarios(t *testing.T) {
	server := setupE2ETestServer()
	defer server.Close()

	client := server.Client()

	t.Run("create form with invalid JSON", func(t *testing.T) {
		resp, err := client.Post(
			server.URL+"/forms",
			"application/json",
			bytes.NewReader([]byte(`{"title":`)),
		)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("add question with invalid JSON", func(t *testing.T) {
		createResp, err := client.Post(
			server.URL+"/forms",
			"application/json",
			bytes.NewReader([]byte(`{"title":"Invalid JSON Test"}`)),
		)
		if err != nil {
			t.Fatalf("create form request failed: %v", err)
		}
		defer createResp.Body.Close()

		var form models.Form
		if err := json.NewDecoder(createResp.Body).Decode(&form); err != nil {
			t.Fatalf("failed to decode create form response: %v", err)
		}

		resp, err := client.Post(
			server.URL+"/forms/"+form.ID+"/questions",
			"application/json",
			bytes.NewReader([]byte(`{"text":`)),
		)
		if err != nil {
			t.Fatalf("add question request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("submit response with invalid JSON", func(t *testing.T) {
		createResp, err := client.Post(
			server.URL+"/forms",
			"application/json",
			bytes.NewReader([]byte(`{"title":"Submit Invalid JSON Test"}`)),
		)
		if err != nil {
			t.Fatalf("create form request failed: %v", err)
		}
		defer createResp.Body.Close()

		var form models.Form
		if err := json.NewDecoder(createResp.Body).Decode(&form); err != nil {
			t.Fatalf("failed to decode create form response: %v", err)
		}

		addQuestionResp, err := client.Post(
			server.URL+"/forms/"+form.ID+"/questions",
			"application/json",
			bytes.NewReader([]byte(`{"text":"Question?","type":"text"}`)),
		)
		if err != nil {
			t.Fatalf("add question request failed: %v", err)
		}
		defer addQuestionResp.Body.Close()

		saveReq, err := http.NewRequest(
			http.MethodPost,
			server.URL+"/forms/"+form.ID+"/save",
			nil,
		)
		if err != nil {
			t.Fatalf("failed to build save request: %v", err)
		}

		saveResp, err := client.Do(saveReq)
		if err != nil {
			t.Fatalf("save form request failed: %v", err)
		}
		defer saveResp.Body.Close()

		var savedForm models.Form
		if err := json.NewDecoder(saveResp.Body).Decode(&savedForm); err != nil {
			t.Fatalf("failed to decode save form response: %v", err)
		}

		resp, err := client.Post(
			server.URL+"/form/"+savedForm.ShareSlug+"/responses",
			"application/json",
			bytes.NewReader([]byte(`{"answers":`)),
		)
		if err != nil {
			t.Fatalf("submit response request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}
