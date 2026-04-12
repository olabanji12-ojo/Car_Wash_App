package services

import (
	"context"
	"fmt"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/config"
	"github.com/olabanji12-ojo/CarWashApp/database"
)

type LockService interface {
	AcquireLock(key string, userID string) (bool, error)
	ReleaseLock(key string, userID string) error
	IsLockedByOther(key string, userID string) (bool, error)
}

type RedisLockService struct {
	ttl int
}

func NewRedisLockService() LockService {
	return &RedisLockService{
		ttl: config.Cfg.BookingLockTime,
	}
}

// AcquireLock attempts to set a key in Redis only if it doesn't exist
// This represents "Reserving" a slot for a specific user
func (s *RedisLockService) AcquireLock(key string, userID string) (bool, error) {
	if database.RedisClient == nil {
		return true, nil // Fail-open if Redis is down (MVP policy)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// SET key value NX EX ttl
	// NX: Only set if it doesn't exist
	// EX: Expire after ttl seconds
	success, err := database.RedisClient.SetNX(ctx, "lock:"+key, userID, time.Duration(s.ttl)*time.Second).Result()
	if err != nil {
		return false, err
	}

	return success, nil
}

// ReleaseLock removes the key if the userID matches (preventing a user from clearing someone else's lock)
func (s *RedisLockService) ReleaseLock(key string, userID string) error {
	if database.RedisClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Only delete if the value matches the userID
	val, err := database.RedisClient.Get(ctx, "lock:"+key).Result()
	if err == nil && val == userID {
		return database.RedisClient.Del(ctx, "lock:"+key).Err()
	}

	return nil
}

// IsLockedByOther checks if the key is held by anyone other than the current user
func (s *RedisLockService) IsLockedByOther(key string, userID string) (bool, error) {
	if database.RedisClient == nil {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	val, err := database.RedisClient.Get(ctx, "lock:"+key).Result()
	if err != nil {
		return false, nil // Not locked
	}

	return val != userID, nil
}

// GetLockKey generates a unique key for a carwash slot
func GetLockKey(carwashID, slotTime string) string {
	return fmt.Sprintf("slot:%s:%s", carwashID, slotTime)
}
