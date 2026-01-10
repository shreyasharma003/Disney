package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// RedisKeyPrefix is the prefix for recently viewed keys
	RedisKeyPrefix = "recently_viewed:user:"
	// MaxRecentlyViewedCount is the maximum number of items to store
	MaxRecentlyViewedCount = 5
	// TTL is the time-to-live for the Redis key (24 hours)
	TTL = 24 * time.Hour
)

var redisClient *redis.Client

// SetRedisClient sets the Redis client for the service
func SetRedisClient(client *redis.Client) {
	redisClient = client
}

// AddRecentlyViewed adds a cartoon to the user's recently viewed list
// Logic:
// 1. Remove the cartoon ID if it already exists (LREM)
// 2. Push the cartoon ID to the front of the list (LPUSH)
// 3. Trim the list to keep only the latest 5 items (LTRIM)
// 4. Set TTL on the key to 24 hours (EXPIRE)
func AddRecentlyViewed(userId int, cartoonId int) error {
	// Check if Redis is available
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("%s%d", RedisKeyPrefix, userId)
	cartoonIDStr := strconv.Itoa(cartoonId)

	// Step 1: Remove the cartoon ID if it already exists in the list
	// This ensures we don't have duplicates and maintains the "most recent" logic
	err := redisClient.LRem(ctx, key, 0, cartoonIDStr).Err()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to remove cartoon from recently viewed: %w", err)
	}

	// Step 2: Push the cartoon ID to the front of the list (most recent first)
	err = redisClient.LPush(ctx, key, cartoonIDStr).Err()
	if err != nil {
		return fmt.Errorf("failed to add cartoon to recently viewed: %w", err)
	}

	// Step 3: Trim the list to keep only the latest 5 items
	// LTRIM keeps elements from index 0 to 4 (5 elements total)
	err = redisClient.LTrim(ctx, key, 0, MaxRecentlyViewedCount-1).Err()
	if err != nil {
		return fmt.Errorf("failed to trim recently viewed list: %w", err)
	}

	// Step 4: Set expiration time (TTL) to 24 hours
	err = redisClient.Expire(ctx, key, TTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL on recently viewed key: %w", err)
	}

	return nil
}

// GetRecentlyViewed retrieves the list of recently viewed cartoon IDs for a user
// Returns the IDs in order from most recent to oldest
func GetRecentlyViewed(userId int) ([]int, error) {
	// Check if Redis is available
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("%s%d", RedisKeyPrefix, userId)

	// Get all items from the list (index 0 to -1)
	val, err := redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to retrieve recently viewed list: %w", err)
	}

	// If key doesn't exist or list is empty, return empty slice
	if len(val) == 0 {
		return []int{}, nil
	}

	// Convert string IDs to integers
	cartoonIds := make([]int, len(val))
	for i, idStr := range val {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse cartoon ID: %w", err)
		}
		cartoonIds[i] = id
	}

	return cartoonIds, nil
}
