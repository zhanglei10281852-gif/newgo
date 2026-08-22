package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type EvidenceItem struct {
	Kind, Identifier, Checksum string
	Size                       int64
}

func Checksum(body []byte) string { sum := sha256.Sum256(body); return hex.EncodeToString(sum[:]) }
func (e EvidenceItem) Validate() error {
	if strings.TrimSpace(e.Kind) == "" || strings.TrimSpace(e.Identifier) == "" || len(e.Checksum) != 64 {
		return FieldError{"evidence", "kind, identifier and sha256 are required"}
	}
	if e.Size < 0 {
		return FieldError{"size", "cannot be negative"}
	}
	return nil
}
func EvidenceFingerprint(items []EvidenceItem) (string, error) {
	copyItems := append([]EvidenceItem(nil), items...)
	sort.Slice(copyItems, func(i, j int) bool {
		if copyItems[i].Kind == copyItems[j].Kind {
			return copyItems[i].Identifier < copyItems[j].Identifier
		}
		return copyItems[i].Kind < copyItems[j].Kind
	})
	var builder strings.Builder
	for _, item := range copyItems {
		if err := item.Validate(); err != nil {
			return "", err
		}
		fmt.Fprintf(&builder, "%s\x00%s\x00%s\x00%d\n", item.Kind, item.Identifier, item.Checksum, item.Size)
	}
	return Checksum([]byte(builder.String())), nil
}
func EvidenceComplete(required []string, items []EvidenceItem) ([]string, bool) {
	present := map[string]bool{}
	for _, item := range items {
		present[item.Kind] = true
	}
	missing := make([]string, 0)
	for _, kind := range required {
		if !present[kind] {
			missing = append(missing, kind)
		}
	}
	return missing, len(missing) == 0
}
