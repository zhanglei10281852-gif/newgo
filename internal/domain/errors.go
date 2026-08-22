package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("business conflict")
	ErrForbidden = errors.New("forbidden")
	ErrInvalid   = errors.New("invalid input")
	ErrExpired   = errors.New("expired")
	ErrCanceled  = errors.New("operation canceled")
)

type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string { return fmt.Sprintf("%s: %s", e.Field, e.Message) }
func (e FieldError) Unwrap() error { return ErrInvalid }

type ConflictError struct{ Resource, Reason string }

func (e ConflictError) Error() string { return fmt.Sprintf("%s conflict: %s", e.Resource, e.Reason) }
func (e ConflictError) Unwrap() error { return ErrConflict }

type TransitionError struct{ Entity, From, To string }

func (e TransitionError) Error() string {
	return fmt.Sprintf("%s cannot transition from %s to %s", e.Entity, e.From, e.To)
}
func (e TransitionError) Unwrap() error { return ErrConflict }
func IsNotFound(err error) bool         { return errors.Is(err, ErrNotFound) }
