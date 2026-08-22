package domain

type EquipmentSlot struct {
	ID, Label string
	Capacity  int
	Reserved  int
	Version   int64
}

func (s EquipmentSlot) Available() int {
	value := s.Capacity - s.Reserved
	if value < 0 {
		return 0
	}
	return value
}
func (s EquipmentSlot) Reserve(units int) (EquipmentSlot, error) {
	if units <= 0 {
		return s, FieldError{"units", "must be positive"}
	}
	if units > s.Available() {
		return s, ConflictError{"equipment_slot", "capacity exhausted"}
	}
	s.Reserved += units
	s.Version++
	return s, nil
}
func (s EquipmentSlot) Release(units int) (EquipmentSlot, error) {
	if units <= 0 || units > s.Reserved {
		return s, ConflictError{"equipment_slot", "invalid release"}
	}
	s.Reserved -= units
	s.Version++
	return s, nil
}
func (s EquipmentSlot) Validate() error {
	if s.ID == "" || s.Label == "" {
		return FieldError{"slot", "id and label are required"}
	}
	if s.Capacity < 0 || s.Reserved < 0 || s.Reserved > s.Capacity {
		return FieldError{"capacity", "invalid reservation"}
	}
	return nil
}
