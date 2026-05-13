package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"typeform-test/models"
	"typeform-test/store"

	"github.com/google/uuid"
)

type Handler struct {
	Store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{Store: s}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) CreateForm(w http.ResponseWriter, r *http.Request) {
	var f models.Form
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f.ID = uuid.New().String()
	f.CreatedAt = time.Now()
	f.Questions = []models.Question{}

	if err := h.Store.CreateForm(r.Context(), &f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
}

func (h *Handler) AddQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	form, err := h.Store.GetForm(r.Context(), id)
	if err != nil {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}

	var q models.Question
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	q.ID = uuid.New().String()
	form.Questions = append(form.Questions, q)

	if err := h.Store.UpdateForm(r.Context(), form); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(q)
}

func (h *Handler) SaveForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	form, err := h.Store.GetForm(r.Context(), id)
	if err != nil {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}

	if len(form.Questions) == 0 {
		http.Error(w, "form must have at least one question", http.StatusBadRequest)
		return
	}

	// Simple slug generation: take first 8 chars of a new UUID
	form.ShareSlug = uuid.New().String()[:8]

	if err := h.Store.UpdateForm(r.Context(), form); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(form)
}

func (h *Handler) ListForms(w http.ResponseWriter, r *http.Request) {
	forms, err := h.Store.ListForms(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(forms)
}

func (h *Handler) GetForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	form, err := h.Store.GetForm(r.Context(), id)
	if err != nil {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(form)
}

func (h *Handler) GetFormBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "missing slug", http.StatusBadRequest)
		return
	}

	form, err := h.Store.GetFormBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(form)
}

func (h *Handler) SubmitResponse(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "missing slug", http.StatusBadRequest)
		return
	}

	form, err := h.Store.GetFormBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}

	var resp models.Response
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp.ID = uuid.New().String()
	resp.FormID = form.ID
	resp.CreatedAt = time.Now()

	if err := h.Store.SaveResponse(r.Context(), &resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
