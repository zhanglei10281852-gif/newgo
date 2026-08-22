package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type IdempotencyRecord struct {
	Scope, Key, Method, Path, RequestHash string
	StatusCode                            int
	Response                              []byte
	ExpiresAt, CreatedAt                  time.Time
}

func IdempotencyHash(value any) (string, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode idempotency request: %w", err)
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func NewIdempotencyRecord(scope, key, method, path, hash string, status int, response []byte, now, expires time.Time) (IdempotencyRecord, error) {
	record := IdempotencyRecord{Scope: strings.TrimSpace(scope), Key: strings.TrimSpace(key), Method: strings.ToUpper(strings.TrimSpace(method)), Path: strings.TrimSpace(path), RequestHash: hash, StatusCode: status, Response: append([]byte(nil), response...), CreatedAt: now.UTC(), ExpiresAt: expires.UTC()}
	if err := record.Validate(); err != nil {
		return IdempotencyRecord{}, err
	}
	return record, nil
}

func (r IdempotencyRecord) Validate() error {
	if r.Scope == "" || r.Key == "" || r.Method == "" || r.Path == "" || r.RequestHash == "" {
		return FieldError{"idempotency", "scope, key, request identity and hash are required"}
	}
	if r.StatusCode < 100 || r.StatusCode > 599 {
		return FieldError{"status_code", "is invalid"}
	}
	if !r.ExpiresAt.After(r.CreatedAt) {
		return FieldError{"expires_at", "must follow creation"}
	}
	return nil
}

func (r IdempotencyRecord) Match(scope, key, method, path, hash string, now time.Time) error {
	if !r.ExpiresAt.After(now) {
		return ErrExpired
	}
	if r.Scope != strings.TrimSpace(scope) || r.Key != strings.TrimSpace(key) {
		return ErrNotFound
	}
	if r.Method != strings.ToUpper(strings.TrimSpace(method)) || r.Path != strings.TrimSpace(path) {
		return ConflictError{"idempotency", "key belongs to another operation"}
	}
	if r.RequestHash != hash {
		return ConflictError{"idempotency", "request payload changed"}
	}
	return nil
}

func (r IdempotencyRecord) Replay() []byte { return append([]byte(nil), r.Response...) }
func IdempotencyIdentity(method, path, tenant string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + "\x00" + strings.TrimSpace(path) + "\x00" + strings.TrimSpace(tenant)
}
