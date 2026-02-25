package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	notifKeyPrefix = "notifications:"
	maxPerUser     = 100
	ttl            = 7 * 24 * time.Hour
)

type Notification struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
	CreatedAt int64  `json:"createdAt"`
	Read      bool   `json:"read"`
}

type NotificationService struct {
	rdb *redis.Client
}

func NewNotificationService(rdb *redis.Client) *NotificationService {
	return &NotificationService{rdb: rdb}
}

func (s *NotificationService) Send(ctx context.Context, notif Notification) error {
	if notif.UserID == "" {
		return fmt.Errorf("userId requis")
	}
	if notif.Type == "" {
		return fmt.Errorf("type requis")
	}

	notif.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	notif.CreatedAt = time.Now().Unix()
	notif.Read = false

	data, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("erreur sérialisation: %w", err)
	}

	key := notifKeyPrefix + notif.UserID
	pipe := s.rdb.Pipeline()
	pipe.LPush(ctx, key, string(data))
	pipe.LTrim(ctx, key, 0, maxPerUser-1)
	pipe.Expire(ctx, key, ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *NotificationService) GetPending(ctx context.Context, userID string) ([]Notification, error) {
	if userID == "" {
		return nil, fmt.Errorf("userId requis")
	}

	key := notifKeyPrefix + userID
	items, err := s.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("erreur Redis: %w", err)
	}

	notifs := make([]Notification, 0, len(items))
	for _, item := range items {
		var n Notification
		if err := json.Unmarshal([]byte(item), &n); err != nil {
			continue
		}
		if !n.Read {
			notifs = append(notifs, n)
		}
	}
	return notifs, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("userId requis")
	}

	key := notifKeyPrefix + userID
	items, err := s.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("erreur Redis: %w", err)
	}

	pipe := s.rdb.Pipeline()
	for i, item := range items {
		var n Notification
		if err := json.Unmarshal([]byte(item), &n); err != nil {
			continue
		}
		n.Read = true
		data, _ := json.Marshal(n)
		pipe.LSet(ctx, key, int64(i), string(data))
	}
	_, err = pipe.Exec(ctx)
	return err
}