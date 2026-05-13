package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"typeform-test/models"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func NewCache(url string) *Cache {
	if url == "" {
		return nil
	}
	client := redis.NewClient(&redis.Options{
		Addr: url,
	})
	return &Cache{client: client}
}

func (c *Cache) GetFormBySlug(ctx context.Context, slug string) (*models.Form, error) {
	if c == nil {
		return nil, nil
	}
	val, err := c.client.Get(ctx, fmt.Sprintf("form:slug:%s", slug)).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var form models.Form
	if err := json.Unmarshal([]byte(val), &form); err != nil {
		return nil, err
	}
	return &form, nil
}

func (c *Cache) SetFormBySlug(ctx context.Context, slug string, form *models.Form) error {
	if c == nil {
		return nil
	}
	data, err := json.Marshal(form)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fmt.Sprintf("form:slug:%s", slug), data, 10*time.Minute).Err()
}

func (c *Cache) InvalidateFormBySlug(ctx context.Context, slug string) error {
	if c == nil || slug == "" {
		return nil
	}
	return c.client.Del(ctx, fmt.Sprintf("form:slug:%s", slug)).Err()
}

func (c *Cache) InvalidateFormByID(ctx context.Context, id string) error {
	if c == nil {
		return nil
	}
	return nil
}
