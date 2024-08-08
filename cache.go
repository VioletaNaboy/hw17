package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Cache struct {
	client *redis.Client
}

func NewCache() *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return &Cache{client: rdb}
}

func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *Cache) Get(key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (c *Cache) Del(key string) error {
	return c.client.Del(ctx, key).Err()
}
