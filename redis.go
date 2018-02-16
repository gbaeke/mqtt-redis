package main

import (
	"github.com/go-redis/redis"
)

// RedisClient struct
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient returns a point to RedisClient
func NewRedisClient(client *redis.Client) (*RedisClient, error) {
	// try to ping the Redis server
	err := client.Ping().Err()

	// return *RedisClient and error
	// even with error, connection can come back up later
	return &RedisClient{Client: client}, err
}

// Publish a message to a channel and return error if any
func (redis *RedisClient) Publish(channel, message string) error {
	return redis.Client.Publish(channel, message).Err()
}
