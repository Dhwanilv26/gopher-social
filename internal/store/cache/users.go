package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Dhwanilv26/gopher-social/internal/store"
	"github.com/redis/go-redis/v9"
)

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute

func (s *UserStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%d", userID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()

	if err != nil {
		fmt.Printf("REDIS GET: key=%s, error=%v\n", cacheKey, err)
		return nil, err
	}

	fmt.Printf("REDIS GET: key=%s, data=%s\n", cacheKey, data)

	var user store.User

	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	if err := s.rdb.Set(ctx, cacheKey, json, UserExpTime).Err(); err != nil {
		return err
	}
	return nil
}
