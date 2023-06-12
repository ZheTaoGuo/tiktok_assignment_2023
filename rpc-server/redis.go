package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	cli *redis.Client
}

func (c *RedisClient) InitClient(ctx context.Context, address, password string) error {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	if err := r.Ping(ctx).Err(); err != nil {
		return err
	}

	c.cli = r
	return nil
}

func (c *RedisClient) SaveMessageToRedis(ctx context.Context, key string, message *Message) error {
	// Marshal the message struct to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Save the message to Redis using the key and JSON value
	err = c.cli.HSet(ctx, key, message.Timestamp, messageJSON).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *RedisClient) GetMessagesByGroupID(ctx context.Context, groupID string, start, end int64, reverse bool) ([]*Message, error) {
	var (
		rawMessages map[string]string
		messages    []*Message
		err         error
	)

	fmt.Printf("Retrieving messages from Redis with key %s\n", groupID)

	// Retrieve all fields and values of the hash
	rawMessages, err = c.cli.HGetAll(ctx, groupID).Result()
	if err != nil {
		return nil, err
	}

	// Create a slice to store the timestamps for sorting
	timestamps := make([]int64, 0, len(rawMessages))

	// Extract and sort the timestamps
	for timestampStr := range rawMessages {
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return nil, err
		}
		timestamps = append(timestamps, timestamp)
	}

	// Sort the timestamps based on the desired order
	if reverse {
		sort.Slice(timestamps, func(i, j int) bool {
			return timestamps[i] > timestamps[j]
		})
	} else {
		sort.Slice(timestamps, func(i, j int) bool {
			return timestamps[i] < timestamps[j]
		})
	}

	// Apply start and end indices
	startIndex := int(start)
	endIndex := int(end)
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex >= len(timestamps) {
		endIndex = len(timestamps) - 1
	}
	if endIndex < startIndex {
		return nil, fmt.Errorf("invalid start and end indices")
	}

	// Iterate over the selected range of timestamps and retrieve the corresponding messages
	for i := startIndex; i <= endIndex; i++ {
		timestamp := timestamps[i]
		timestampStr := strconv.FormatInt(timestamp, 10)
		msgJSON := rawMessages[timestampStr]
		temp := &Message{}
		err := json.Unmarshal([]byte(msgJSON), temp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, temp)
	}

	fmt.Println("Messages retrieved from Redis:", messages)

	return messages, nil
}

