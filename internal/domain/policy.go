package domain

import (
	"fmt"
	"strings"
	"time"
)

type PermitRule struct {
	Name                string
	MinimumStationCount int
	MaximumMagnitude    float64
	RequiresClassified  bool
	Window              time.Duration
}
type PermitDecision struct {
	Allowed  bool
	Reasons  []string
	Rule     PermitRule
	Envelope RiskEnvelope
}

func DefaultPermitRule() PermitRule {
	return PermitRule{Name: "hydraulic-stimulation-default", MinimumStationCount: 2, MaximumMagnitude: 2, RequiresClassified: true, Window: 2 * time.Hour}
}
func EvaluatePermit(rule PermitRule, envelope RiskEnvelope, stationCount int) PermitDecision {
	d := PermitDecision{Rule: rule, Envelope: envelope, Allowed: true, Reasons: make([]string, 0)}
	if stationCount < rule.MinimumStationCount {
		d.Allowed = false
		d.Reasons = append(d.Reasons, fmt.Sprintf("need %d calibrated stations", rule.MinimumStationCount))
	}
	if envelope.PeakMagnitude >= rule.MaximumMagnitude {
		d.Allowed = false
		d.Reasons = append(d.Reasons, "peak magnitude exceeds rule")
	}
	if rule.RequiresClassified && envelope.Unclassified > 0 {
		d.Allowed = false
		d.Reasons = append(d.Reasons, "unclassified events remain")
	}
	if envelope.Band == RiskRed || envelope.Band == RiskBlack {
		d.Allowed = false
		d.Reasons = append(d.Reasons, "risk band blocks stimulation")
	}
	return d
}
func (d PermitDecision) Message() string {
	if d.Allowed {
		return "permit eligible"
	}
	return strings.Join(d.Reasons, "; ")
}
func (d PermitDecision) HasReason(fragment string) bool {
	for _, reason := range d.Reasons {
		if strings.Contains(reason, fragment) {
			return true
		}
	}
	return false
}

type ClassificationRule struct {
	Label            string
	Minimum, Maximum float64
	Priority         int
}

func DefaultClassificationRules() []ClassificationRule {
	return []ClassificationRule{{Label: "background", Minimum: 0, Maximum: 0.8, Priority: 1}, {Label: "induced", Minimum: 0.8, Maximum: 2, Priority: 2}, {Label: "tectonic", Minimum: 2, Maximum: 10, Priority: 3}}
}
func ClassifyMagnitude(value float64, rules []ClassificationRule) (ClassificationRule, error) {
	for _, rule := range rules {
		if value >= rule.Minimum && value < rule.Maximum {
			return rule, nil
		}
	}
	return ClassificationRule{}, ConflictError{"event", "magnitude outside classification envelope"}
}
func SortRules(rules []ClassificationRule) []ClassificationRule {
	copyRules := append([]ClassificationRule(nil), rules...)
	for i := 0; i < len(copyRules); i++ {
		for j := i + 1; j < len(copyRules); j++ {
			if copyRules[j].Priority < copyRules[i].Priority {
				copyRules[i], copyRules[j] = copyRules[j], copyRules[i]
			}
		}
	}
	return copyRules
}

type Alert struct {
	Code, Severity, Message string
	EventID                 string
	At                      time.Time
}

func BuildAlerts(envelope RiskEnvelope, threshold float64) []Alert {
	alerts := make([]Alert, 0)
	if envelope.Unclassified > 0 {
		alerts = append(alerts, Alert{Code: "UNCLASSIFIED", Severity: "warning", Message: "event classification pending", At: envelope.WindowEnd})
	}
	if envelope.PeakMagnitude >= threshold {
		alerts = append(alerts, Alert{Code: "MAGNITUDE", Severity: "critical", Message: "peak magnitude above well threshold", At: envelope.WindowEnd})
	}
	if envelope.Band == RiskBlack {
		alerts = append(alerts, Alert{Code: "RISK_BLACK", Severity: "critical", Message: "stimulation must be suspended", At: envelope.WindowEnd})
	}
	return alerts
}
func HighestSeverity(alerts []Alert) string {
	rank := map[string]int{"info": 0, "warning": 1, "critical": 2}
	best := "info"
	for _, alert := range alerts {
		if rank[alert.Severity] > rank[best] {
			best = alert.Severity
		}
	}
	return best
}

type ReviewChecklist struct {
	Items     []string
	Completed map[string]bool
}

func NewChecklist(items []string) ReviewChecklist {
	completed := make(map[string]bool, len(items))
	for _, item := range items {
		completed[item] = false
	}
	return ReviewChecklist{Items: append([]string(nil), items...), Completed: completed}
}
func (c ReviewChecklist) Complete(item string) error {
	if _, ok := c.Completed[item]; !ok {
		return FieldError{"checklist", item + " is not a known item"}
	}
	c.Completed[item] = true
	return nil
}
func (c ReviewChecklist) Ready() bool {
	for _, item := range c.Items {
		if !c.Completed[item] {
			return false
		}
	}
	return true
}
func (c ReviewChecklist) Missing() []string {
	missing := make([]string, 0)
	for _, item := range c.Items {
		if !c.Completed[item] {
			missing = append(missing, item)
		}
	}
	return missing
}

type AuditQuery struct {
	ActorID, Action, EntityType string
	Since, Until                *time.Time
	Limit                       int
}

func (q AuditQuery) Validate() error {
	if q.Limit < 0 || q.Limit > 1000 {
		return FieldError{"limit", "must be between 0 and 1000"}
	}
	if q.Since != nil && q.Until != nil && !q.Until.After(*q.Since) {
		return FieldError{"time", "until must be after since"}
	}
	return nil
}
func (q AuditQuery) Matches(e AuditEvent) bool {
	if q.ActorID != "" && q.ActorID != e.ActorID {
		return false
	}
	if q.Action != "" && q.Action != e.Action {
		return false
	}
	if q.EntityType != "" && q.EntityType != e.EntityType {
		return false
	}
	if q.Since != nil && e.CreatedAt.Before(*q.Since) {
		return false
	}
	if q.Until != nil && !e.CreatedAt.Before(*q.Until) {
		return false
	}
	return true
}

type StateHistory struct {
	EntityID        string
	From, To        string
	At              time.Time
	ActorID, Reason string
}

func (h StateHistory) Valid() error {
	if h.EntityID == "" || h.From == "" || h.To == "" || h.ActorID == "" {
		return FieldError{"history", "entity and actors are required"}
	}
	if h.From == h.To {
		return ConflictError{"history", "state did not change"}
	}
	return nil
}
func HistoryFor(id string, history []StateHistory) []StateHistory {
	out := make([]StateHistory, 0)
	for _, item := range history {
		if item.EntityID == id {
			out = append(out, item)
		}
	}
	return out
}

type BatchCloseReport struct {
	BatchID                           string
	Closed                            bool
	EventCount, Classified, Escalated int
	Reasons                           []string
}

func BuildBatchCloseReport(batch MonitoringBatch, events []SeismicEvent) BatchCloseReport {
	r := BatchCloseReport{BatchID: batch.ID, EventCount: len(events), Reasons: make([]string, 0)}
	for _, event := range events {
		if event.Status == EventClassified {
			r.Classified++
		}
		if event.Status == EventEscalated {
			r.Escalated++
		}
	}
	if batch.State != BatchCollecting {
		r.Reasons = append(r.Reasons, "batch is not collecting")
	}
	if r.EventCount == 0 {
		r.Reasons = append(r.Reasons, "no events recorded")
	}
	if r.Classified+r.Escalated != r.EventCount {
		r.Reasons = append(r.Reasons, "events still unclassified")
	}
	r.Closed = len(r.Reasons) == 0
	return r
}

func RequireDistinct(a, b string) error {
	if strings.TrimSpace(a) == "" || strings.TrimSpace(b) == "" {
		return FieldError{"operator", "both operators are required"}
	}
	if a == b {
		return ConflictError{"operator", "requester and reviewer must differ"}
	}
	return nil
}
func RequireFuture(value, now time.Time) error {
	if !value.After(now) {
		return ErrExpired
	}
	return nil
}
func NormalizeReference(value string) string { return strings.ToUpper(strings.TrimSpace(value)) }
func NormalizeSerial(value string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), " ", "-"))
}
func SafeNote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 500 {
		return value[:500]
	}
	return value
}
