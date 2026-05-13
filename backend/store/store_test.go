package store

import (
	"context"
	"os"
	"testing"
	"typeform-test/models"
)

func TestMemoryStore(t *testing.T) {
	s := NewMemoryStore()
	testStore(t, s)
}

func TestSQLiteStore(t *testing.T) {
	// Use a temporary file for SQLite test
	dbPath := "test_forms.db"
	os.Remove(dbPath) // Clean up before test
	defer os.Remove(dbPath)

	s, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite store: %v", err)
	}
	defer s.db.Close() // Close to release lock

	testStore(t, s)
}

func TestPostgresStore(t *testing.T) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://user:pass@localhost:5432/forms?sslmode=disable"
	}

	s, err := NewPostgresStore(connStr)
	if err != nil {
		t.Skipf("skipping postgres test: could not connect: %v", err)
	}
	defer s.db.Close()

	// Clean up for deterministic test
	s.db.Exec("DELETE FROM responses")
	s.db.Exec("DELETE FROM forms")

	testStore(t, s)
}

func testStore(t *testing.T, s Store) {
	ctx := context.Background()

	// 1. Create Form
	f := &models.Form{
		ID:    "form-1",
		Title: "Test Form",
	}
	if err := s.CreateForm(ctx, f); err != nil {
		t.Fatalf("failed to create form: %v", err)
	}

	// 2. Get Form
	got, err := s.GetForm(ctx, "form-1")
	if err != nil {
		t.Fatalf("failed to get form: %v", err)
	}
	if got.Title != "Test Form" {
		t.Errorf("expected title 'Test Form', got %q", got.Title)
	}

	// 3. Update Form (add questions)
	got.Questions = append(got.Questions, models.Question{ID: "q1", Text: "Question 1", Type: "text"})
	got.ShareSlug = "test-slug"
	if err := s.UpdateForm(ctx, got); err != nil {
		t.Fatalf("failed to update form: %v", err)
	}

	// 4. Get Form by Slug
	gotBySlug, err := s.GetFormBySlug(ctx, "test-slug")
	if err != nil {
		t.Fatalf("failed to get form by slug: %v", err)
	}
	if gotBySlug.ID != "form-1" {
		t.Errorf("expected form-1, got %q", gotBySlug.ID)
	}
	if len(gotBySlug.Questions) != 1 {
		t.Errorf("expected 1 question, got %d", len(gotBySlug.Questions))
	}

	// 5. List Forms
	forms, err := s.ListForms(ctx)
	if err != nil {
		t.Fatalf("failed to list forms: %v", err)
	}
	if len(forms) != 1 {
		t.Errorf("expected 1 form, got %d", len(forms))
	}

	// 6. Save Response
	resp := &models.Response{
		ID:     "resp-1",
		FormID: "form-1",
		Answers: []models.Answer{
			{QuestionID: "q1", Value: "Answer 1"},
		},
	}
	if err := s.SaveResponse(ctx, resp); err != nil {
		t.Fatalf("failed to save response: %v", err)
	}

	// 7. Error cases
	_, err = s.GetForm(ctx, "non-existent")
	if err != ErrFormNotFound {
		t.Errorf("expected ErrFormNotFound, got %v", err)
	}
}
