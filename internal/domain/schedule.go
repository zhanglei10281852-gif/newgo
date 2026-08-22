package domain

import (
	"fmt"
	"sort"
	"time"
)

type OperationKind string

const (
	OperationMonitoring  OperationKind = "monitoring"
	OperationStimulation OperationKind = "stimulation"
	OperationMaintenance OperationKind = "maintenance"
)

type ScheduledOperation struct {
	ID, FieldID, WellID, EquipmentSlot, OwnerID string
	Kind                                        OperationKind
	Window                                      TimeWindow
	Priority                                    int
	State                                       string
	Version                                     int64
}

func (o ScheduledOperation) Validate() error {
	if o.ID == "" || o.FieldID == "" || o.WellID == "" || o.OwnerID == "" {
		return FieldError{"operation", "identity and ownership are required"}
	}
	if o.Kind == "" {
		return FieldError{"operation_kind", "is required"}
	}
	if o.Window.End.IsZero() || !o.Window.End.After(o.Window.Start) {
		return FieldError{"operation_window", "is invalid"}
	}
	if o.Kind == OperationStimulation && o.EquipmentSlot == "" {
		return FieldError{"equipment_slot", "is required for stimulation"}
	}
	return nil
}

func (o ScheduledOperation) Cancel() (ScheduledOperation, error) {
	if o.State == "completed" || o.State == "cancelled" {
		return o, TransitionError{"operation", o.State, "cancelled"}
	}
	o.State = "cancelled"
	o.Version++
	return o, nil
}

func (o ScheduledOperation) Start(now time.Time) (ScheduledOperation, error) {
	if o.State != "scheduled" {
		return o, TransitionError{"operation", o.State, "running"}
	}
	if now.Before(o.Window.Start) || !now.Before(o.Window.End) {
		return o, ConflictError{"operation", "outside scheduled window"}
	}
	o.State = "running"
	o.Version++
	return o, nil
}

func (o ScheduledOperation) Complete(now time.Time) (ScheduledOperation, error) {
	if o.State != "running" {
		return o, TransitionError{"operation", o.State, "completed"}
	}
	if now.Before(o.Window.Start) {
		return o, ConflictError{"operation", "completion precedes operation window"}
	}
	o.State = "completed"
	o.Version++
	return o, nil
}

func DetectScheduleConflicts(candidate ScheduledOperation, existing []ScheduledOperation) []ScheduledOperation {
	conflicts := make([]ScheduledOperation, 0)
	for _, operation := range existing {
		if operation.State == "cancelled" || operation.State == "completed" {
			continue
		}
		if !candidate.Window.Overlaps(operation.Window) {
			continue
		}
		if candidate.WellID == operation.WellID || (candidate.EquipmentSlot != "" && candidate.EquipmentSlot == operation.EquipmentSlot) {
			conflicts = append(conflicts, operation)
		}
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].Window.Start.Before(conflicts[j].Window.Start) })
	return conflicts
}

func FindAvailableWindow(day TimeWindow, duration time.Duration, existing []ScheduledOperation) (TimeWindow, error) {
	if duration <= 0 || duration > day.Duration() {
		return TimeWindow{}, FieldError{"duration", "does not fit business day"}
	}
	operations := append([]ScheduledOperation(nil), existing...)
	sort.Slice(operations, func(i, j int) bool { return operations[i].Window.Start.Before(operations[j].Window.Start) })
	cursor := day.Start
	for _, operation := range operations {
		if operation.State == "cancelled" {
			continue
		}
		if cursor.Add(duration).Before(operation.Window.Start) || cursor.Add(duration).Equal(operation.Window.Start) {
			return NewWindow(cursor, cursor.Add(duration), day.Location)
		}
		if operation.Window.End.After(cursor) {
			cursor = operation.Window.End
		}
	}
	if cursor.Add(duration).After(day.End) {
		return TimeWindow{}, ConflictError{"schedule", "no contiguous window available"}
	}
	return NewWindow(cursor, cursor.Add(duration), day.Location)
}

func ScheduleSummary(operations []ScheduledOperation) map[string]int {
	out := map[string]int{"scheduled": 0, "running": 0, "completed": 0, "cancelled": 0}
	for _, operation := range operations {
		out[operation.State]++
	}
	return out
}

func ValidateNoConflicts(operations []ScheduledOperation) error {
	for i, candidate := range operations {
		if conflicts := DetectScheduleConflicts(candidate, append(operations[:i], operations[i+1:]...)); len(conflicts) > 0 {
			return fmt.Errorf("operation %s conflicts with %s: %w", candidate.ID, conflicts[0].ID, ErrConflict)
		}
	}
	return nil
}
