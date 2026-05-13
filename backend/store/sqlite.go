package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"typeform-test/models"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Performance optimizations for high concurrency
	db.SetMaxOpenConns(25) // SQLite can handle multiple readers in WAL mode
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(0) // Keep connections open

	_, err = db.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA synchronous=NORMAL;
		PRAGMA foreign_keys=ON;
		PRAGMA busy_timeout=5000;
		CREATE TABLE IF NOT EXISTS forms (
			id TEXT PRIMARY KEY,
			title TEXT,
			questions TEXT,
			share_slug TEXT,
			created_at DATETIME
		);
		CREATE TABLE IF NOT EXISTS responses (
			id TEXT PRIMARY KEY,
			form_id TEXT,
			answers TEXT,
			created_at DATETIME
		);
		CREATE INDEX IF NOT EXISTS idx_forms_share_slug ON forms(share_slug);
		CREATE INDEX IF NOT EXISTS idx_responses_form_id ON responses(form_id);
	`)
	if err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) CreateForm(ctx context.Context, form *models.Form) error {
	qs, _ := json.Marshal(form.Questions)
	_, err := s.db.ExecContext(ctx, "INSERT INTO forms (id, title, questions, share_slug, created_at) VALUES (?, ?, ?, ?, ?)",
		form.ID, form.Title, string(qs), form.ShareSlug, form.CreatedAt)
	return err
}

func (s *SQLiteStore) GetForm(ctx context.Context, id string) (*models.Form, error) {
	var f models.Form
	var qs string
	err := s.db.QueryRowContext(ctx, "SELECT id, title, questions, share_slug, created_at FROM forms WHERE id = ?", id).
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

func (s *SQLiteStore) GetFormBySlug(ctx context.Context, slug string) (*models.Form, error) {
	var f models.Form
	var qs string
	err := s.db.QueryRowContext(ctx, "SELECT id, title, questions, share_slug, created_at FROM forms WHERE share_slug = ?", slug).
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

func (s *SQLiteStore) UpdateForm(ctx context.Context, form *models.Form) error {
	qs, _ := json.Marshal(form.Questions)
	_, err := s.db.ExecContext(ctx, "UPDATE forms SET title = ?, questions = ?, share_slug = ? WHERE id = ?",
		form.Title, string(qs), form.ShareSlug, form.ID)
	return err
}

func (s *SQLiteStore) ListForms(ctx context.Context) ([]*models.Form, error) {
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

func (s *SQLiteStore) SaveResponse(ctx context.Context, resp *models.Response) error {
	ans, _ := json.Marshal(resp.Answers)
	_, err := s.db.ExecContext(ctx, "INSERT INTO responses (id, form_id, answers, created_at) VALUES (?, ?, ?, ?)",
		resp.ID, resp.FormID, string(ans), resp.CreatedAt)
	return err
}
