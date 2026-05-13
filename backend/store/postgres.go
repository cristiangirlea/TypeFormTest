package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"typeform-test/models"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Performance optimizations for high concurrency
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(0)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS forms (
			id TEXT PRIMARY KEY,
			title TEXT,
			questions TEXT,
			share_slug TEXT,
			created_at TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS responses (
			id TEXT PRIMARY KEY,
			form_id TEXT,
			answers TEXT,
			created_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_forms_share_slug ON forms(share_slug);
		CREATE INDEX IF NOT EXISTS idx_responses_form_id ON responses(form_id);
	`)
	if err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) CreateForm(ctx context.Context, form *models.Form) error {
	qs, _ := json.Marshal(form.Questions)
	_, err := s.db.ExecContext(ctx, "INSERT INTO forms (id, title, questions, share_slug, created_at) VALUES ($1, $2, $3, $4, $5)",
		form.ID, form.Title, string(qs), form.ShareSlug, form.CreatedAt)
	return err
}

func (s *PostgresStore) GetForm(ctx context.Context, id string) (*models.Form, error) {
	var f models.Form
	var qs string
	err := s.db.QueryRowContext(ctx, "SELECT id, title, questions, share_slug, created_at FROM forms WHERE id = $1", id).
		Scan(&f.ID, &f.Title, &qs, &f.ShareSlug, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrFormNotFound
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(qs), &f.Questions)
	return &f, nil
}

func (s *PostgresStore) GetFormBySlug(ctx context.Context, slug string) (*models.Form, error) {
	var f models.Form
	var qs string
	err := s.db.QueryRowContext(ctx, "SELECT id, title, questions, share_slug, created_at FROM forms WHERE share_slug = $1", slug).
		Scan(&f.ID, &f.Title, &qs, &f.ShareSlug, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrFormNotFound
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(qs), &f.Questions)
	return &f, nil
}

func (s *PostgresStore) UpdateForm(ctx context.Context, form *models.Form) error {
	qs, _ := json.Marshal(form.Questions)
	_, err := s.db.ExecContext(ctx, "UPDATE forms SET title = $1, questions = $2, share_slug = $3 WHERE id = $4",
		form.Title, string(qs), form.ShareSlug, form.ID)
	return err
}

func (s *PostgresStore) ListForms(ctx context.Context) ([]*models.Form, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, title, questions, share_slug, created_at FROM forms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	forms := []*models.Form{}
	for rows.Next() {
		var f models.Form
		var qs string
		if err := rows.Scan(&f.ID, &f.Title, &qs, &f.ShareSlug, &f.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(qs), &f.Questions)
		forms = append(forms, &f)
	}
	return forms, nil
}

func (s *PostgresStore) SaveResponse(ctx context.Context, resp *models.Response) error {
	ans, _ := json.Marshal(resp.Answers)
	_, err := s.db.ExecContext(ctx, "INSERT INTO responses (id, form_id, answers, created_at) VALUES ($1, $2, $3, $4)",
		resp.ID, resp.FormID, string(ans), resp.CreatedAt)
	return err
}
