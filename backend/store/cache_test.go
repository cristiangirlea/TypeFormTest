package store

import (
	"context"
	"os"
	"testing"
	"typeform-test/models"
)

func TestCache(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	c := NewCache(redisURL)
	if c == nil {
		t.Skip("Redis URL not provided")
	}

	// Smoke test to see if redis is reachable
	ctx := context.Background()
	err := c.client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("skipping redis test: could not connect: %v", err)
	}

	slug := "test-cache-slug"
	form := &models.Form{
		ID:    "form-1",
		Title: "Cached Form",
	}

	// 1. Set
	err = c.SetFormBySlug(ctx, slug, form)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// 2. Get
	got, err := c.GetFormBySlug(ctx, slug)
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}
	if got == nil || got.Title != "Cached Form" {
		t.Errorf("expected title 'Cached Form', got %v", got)
	}

	// 3. Invalidate
	err = c.InvalidateFormBySlug(ctx, slug)
	if err != nil {
		t.Fatalf("failed to invalidate cache: %v", err)
	}

	// 4. Get after invalidation
	got, err = c.GetFormBySlug(ctx, slug)
	if err != nil {
		t.Fatalf("failed to get cache after invalidation: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil after invalidation, got %v", got)
	}
}

func TestNilCache(t *testing.T) {
	var c *Cache = nil
	ctx := context.Background()

	// Should not panic
	got, err := c.GetFormBySlug(ctx, "slug")
	if got != nil || err != nil {
		t.Error("expected nil, nil from nil cache")
	}

	err = c.SetFormBySlug(ctx, "slug", &models.Form{})
	if err != nil {
		t.Error("expected nil error from nil cache")
	}

	err = c.InvalidateFormBySlug(ctx, "slug")
	if err != nil {
		t.Error("expected nil error from nil cache")
	}
}
