package sqlite

import "time"

func formatTime(value time.Time) string       { return value.UTC().Format(time.RFC3339Nano) }
func parseTime(raw string) (time.Time, error) { return time.Parse(time.RFC3339Nano, raw) }
func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}
func parseNullable(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	value, err := parseTime(*raw)
	return &value, err
}
