package store

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"
	"typeform-test/models"
)

var (
	ErrFormNotFound = errors.New("form not found")
)

type Store interface {
	CreateForm(ctx context.Context, form *models.Form) error
	GetForm(ctx context.Context, id string) (*models.Form, error)
	GetFormBySlug(ctx context.Context, slug string) (*models.Form, error)
	UpdateForm(ctx context.Context, form *models.Form) error
	ListForms(ctx context.Context) ([]*models.Form, error)
	SaveResponse(ctx context.Context, resp *models.Response) error
}

type MemoryStore struct {
	sync.RWMutex
	forms     map[string]*models.Form
	responses []*models.Response
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		forms: make(map[string]*models.Form),
	}
}

func (s *MemoryStore) CreateForm(ctx context.Context, form *models.Form) error {
	s.Lock()
	defer s.Unlock()
	s.forms[form.ID] = form
	return nil
}

func (s *MemoryStore) GetForm(ctx context.Context, id string) (*models.Form, error) {
	s.RLock()
	defer s.RUnlock()
	form, ok := s.forms[id]
	if !ok {
		return nil, ErrFormNotFound
	}
	return form, nil
}

func (s *MemoryStore) GetFormBySlug(ctx context.Context, slug string) (*models.Form, error) {
	s.RLock()
	defer s.RUnlock()
	for f := range maps.Values(s.forms) {
		if f.ShareSlug == slug {
			return f, nil
		}
	}
	return nil, ErrFormNotFound
}

func (s *MemoryStore) UpdateForm(ctx context.Context, form *models.Form) error {
	s.Lock()
	defer s.Unlock()
	s.forms[form.ID] = form
	return nil
}

func (s *MemoryStore) ListForms(ctx context.Context) ([]*models.Form, error) {
	s.RLock()
	defer s.RUnlock()
	return slices.Collect(maps.Values(s.forms)), nil
}

func (s *MemoryStore) SaveResponse(ctx context.Context, resp *models.Response) error {
	s.Lock()
	defer s.Unlock()
	s.responses = append(s.responses, resp)
	return nil
}
