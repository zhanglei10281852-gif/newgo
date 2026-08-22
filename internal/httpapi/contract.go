package httpapi

import (
	"encoding/json"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"net/http"
)

type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}
type ErrorDetail struct {
	Code, Message, RequestID string
	Status                   int
}

func DecodeStrict(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
func StatusFor(err error) int {
	switch {
	case domain.IsNotFound(err):
		return 404
	case IsConflict(err):
		return 409
	case IsForbidden(err):
		return 403
	case IsInvalid(err):
		return 400
	default:
		return 500
	}
}
func IsConflict(err error) bool  { return err != nil && contains(err.Error(), "conflict") }
func IsForbidden(err error) bool { return err != nil && contains(err.Error(), "forbidden") }
func IsInvalid(err error) bool   { return err != nil && contains(err.Error(), "required") }
func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
