package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"shortlink/internal/config"
	"shortlink/internal/models"
)

const (
	urlKeyPrefix = "url:"
	urlCodesKey  = "url:codes"
)

// RedisRepository implements LinkRepository using Redis
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis repository instance
func NewRedisRepository(cfg config.RedisConfig) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection with retry
	ctx := context.Background()
	var lastErr error
	for i := 0; i < 30; i++ {
		_, err := client.Ping(ctx).Result()
		if err == nil {
			log.Println("Connected to Redis successfully")
			return &RedisRepository{client: client}, nil
		}
		lastErr = err
		log.Printf("Waiting for Redis... attempt %d/30", i+1)
		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("failed to connect to Redis after 30 attempts: %w", lastErr)
}

// Save stores a new link in Redis
func (r *RedisRepository) Save(ctx context.Context, link *models.Link) error {
	key := urlKeyPrefix + link.ShortCode

	err := r.client.HSet(ctx, key, map[string]interface{}{
		"original_url": link.OriginalURL,
		"created_at":   link.CreatedAt.Format(time.RFC3339),
		"clicks":       link.Clicks,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to save link: %w", err)
	}

	// Add to the set of all codes
	r.client.SAdd(ctx, urlCodesKey, link.ShortCode)

	return nil
}

// FindByCode retrieves a link by its short code
func (r *RedisRepository) FindByCode(ctx context.Context, code string) (*models.Link, error) {
	key := urlKeyPrefix + code

	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve link: %w", err)
	}

	if len(result) == 0 {
		return nil, ErrLinkNotFound
	}

	createdAt, _ := time.Parse(time.RFC3339, result["created_at"])
	clicks, _ := r.client.HGet(ctx, key, "clicks").Int64()

	return &models.Link{
		ShortCode:   code,
		OriginalURL: result["original_url"],
		CreatedAt:   createdAt,
		Clicks:      clicks,
	}, nil
}

// Exists checks if a short code already exists
func (r *RedisRepository) Exists(ctx context.Context, code string) (bool, error) {
	key := urlKeyPrefix + code
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return exists > 0, nil
}

// IncrementClicks increments the click counter for a link
func (r *RedisRepository) IncrementClicks(ctx context.Context, code string) error {
	key := urlKeyPrefix + code
	return r.client.HIncrBy(ctx, key, "clicks", 1).Err()
}

// FindAll retrieves all links
func (r *RedisRepository) FindAll(ctx context.Context) ([]*models.Link, error) {
	codes, err := r.client.SMembers(ctx, urlCodesKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve link codes: %w", err)
	}

	var links []*models.Link
	for _, code := range codes {
		link, err := r.FindByCode(ctx, code)
		if err != nil {
			continue // Skip invalid entries
		}
		links = append(links, link)
	}

	return links, nil
}

// Delete removes a link by its short code
func (r *RedisRepository) Delete(ctx context.Context, code string) error {
	key := urlKeyPrefix + code

	// Check if exists first
	exists, err := r.Exists(ctx, code)
	if err != nil {
		return err
	}
	if !exists {
		return ErrLinkNotFound
	}

	// Delete the hash and remove from set
	r.client.Del(ctx, key)
	r.client.SRem(ctx, urlCodesKey, code)

	return nil
}

// Close closes the Redis connection
func (r *RedisRepository) Close() error {
	return r.client.Close()
}
