package domain

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type EventImportRow struct {
	ExternalID, OccurredAt string
	Magnitude, DepthKm     float64
}
type EventImportResult struct {
	Rows, Accepted, Rejected int
	Errors                   []string
}

func ParseEventCSV(input io.Reader) ([]EventImportRow, error) {
	reader := csv.NewReader(bufio.NewReader(input))
	reader.FieldsPerRecord = 4
	header, e := reader.Read()
	if e != nil {
		return nil, e
	}
	for i, v := range header {
		header[i] = strings.ToLower(strings.TrimSpace(v))
	}
	required := []string{"external_id", "occurred_at", "magnitude", "depth_km"}
	for i, v := range required {
		if header[i] != v {
			return nil, fmt.Errorf("column %d is %s, want %s", i, header[i], v)
		}
	}
	out := make([]EventImportRow, 0)
	for {
		record, e := reader.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		m, e := strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
		if e != nil {
			return nil, fmt.Errorf("magnitude: %w", e)
		}
		d, e := strconv.ParseFloat(strings.TrimSpace(record[3]), 64)
		if e != nil {
			return nil, fmt.Errorf("depth: %w", e)
		}
		if _, e = time.Parse(time.RFC3339, strings.TrimSpace(record[1])); e != nil {
			return nil, fmt.Errorf("occurred_at: %w", e)
		}
		out = append(out, EventImportRow{ExternalID: strings.TrimSpace(record[0]), OccurredAt: strings.TrimSpace(record[1]), Magnitude: m, DepthKm: d})
	}
	return out, nil
}
func ValidateImport(rows []EventImportRow) EventImportResult {
	result := EventImportResult{Rows: len(rows)}
	seen := map[string]struct{}{}
	for i, row := range rows {
		if row.ExternalID == "" {
			result.Rejected++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: external_id required", i+1))
			continue
		}
		if _, ok := seen[row.ExternalID]; ok {
			result.Rejected++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: duplicate external_id", i+1))
			continue
		}
		seen[row.ExternalID] = struct{}{}
		if row.Magnitude < 0 || row.DepthKm < 0 {
			result.Rejected++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: measurement negative", i+1))
			continue
		}
		result.Accepted++
	}
	return result
}
