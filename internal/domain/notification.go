package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type NotificationChannel string

const (
	ChannelEmail   NotificationChannel = "email"
	ChannelPager   NotificationChannel = "pager"
	ChannelWebhook NotificationChannel = "webhook"
)

type Notification struct {
	ID, Recipient, Subject, Body, DeduplicationKey string
	Channel                                        NotificationChannel
	Priority                                       int
	CreatedAt                                      time.Time
}

func (n Notification) Validate() error {
	if n.ID == "" || strings.TrimSpace(n.Recipient) == "" || strings.TrimSpace(n.Subject) == "" {
		return FieldError{"notification", "id, recipient and subject are required"}
	}
	if n.Channel != ChannelEmail && n.Channel != ChannelPager && n.Channel != ChannelWebhook {
		return FieldError{"channel", "is unsupported"}
	}
	if n.Priority < 0 || n.Priority > 10 {
		return FieldError{"priority", "must be between zero and ten"}
	}
	return nil
}

func BuildRiskNotifications(alerts []Alert, recipients map[string][]NotificationChannel, now time.Time) []Notification {
	out := make([]Notification, 0)
	for recipient, channels := range recipients {
		for _, channel := range channels {
			for _, alert := range alerts {
				priority := 5
				if alert.Severity == "critical" {
					priority = 10
				}
				out = append(out, Notification{ID: NormalizeReference(recipient + "-" + alert.Code + "-" + string(channel)), Recipient: recipient, Subject: alert.Code, Body: alert.Message, Channel: channel, Priority: priority, DeduplicationKey: recipient + "/" + alert.Code + "/" + string(channel), CreatedAt: now.UTC()})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Priority > out[j].Priority })
	return out
}

func DeduplicateNotifications(items []Notification) []Notification {
	seen := map[string]struct{}{}
	out := make([]Notification, 0, len(items))
	for _, item := range items {
		key := item.DeduplicationKey
		if key == "" {
			key = fmt.Sprintf("%s/%s/%s", item.Recipient, item.Subject, item.Channel)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}
func PartitionNotifications(items []Notification, size int) [][]Notification {
	if size <= 0 {
		size = 50
	}
	out := make([][]Notification, 0)
	for len(items) > 0 {
		end := size
		if end > len(items) {
			end = len(items)
		}
		out = append(out, append([]Notification(nil), items[:end]...))
		items = items[end:]
	}
	return out
}
