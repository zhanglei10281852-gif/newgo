package domain

import (
	"math"
	"sort"
	"time"
)

type CalibrationSample struct {
	Reference, StationID string
	At                   time.Time
	Expected, Observed   float64
	OperatorID           string
}
type CalibrationReport struct {
	StationID                                    string
	Samples                                      int
	MeanError, MaximumError, RootMeanSquareError float64
	Passed                                       bool
	CalibratedAt, DueAt                          time.Time
}

func BuildCalibrationReport(stationID string, samples []CalibrationSample, tolerance float64, calibratedAt time.Time, validity time.Duration) (CalibrationReport, error) {
	if stationID == "" || tolerance <= 0 || validity <= 0 {
		return CalibrationReport{}, FieldError{"calibration", "station, tolerance and validity are required"}
	}
	selected := make([]CalibrationSample, 0)
	for _, sample := range samples {
		if sample.StationID == stationID {
			selected = append(selected, sample)
		}
	}
	if len(selected) < 3 {
		return CalibrationReport{}, ConflictError{"calibration", "at least three samples are required"}
	}
	sum, square, maximum := 0.0, 0.0, 0.0
	for _, sample := range selected {
		delta := math.Abs(sample.Observed - sample.Expected)
		sum += delta
		square += delta * delta
		if delta > maximum {
			maximum = delta
		}
	}
	report := CalibrationReport{StationID: stationID, Samples: len(selected), MeanError: sum / float64(len(selected)), MaximumError: maximum, RootMeanSquareError: math.Sqrt(square / float64(len(selected))), CalibratedAt: calibratedAt.UTC(), DueAt: calibratedAt.Add(validity).UTC()}
	report.Passed = report.MaximumError <= tolerance && report.RootMeanSquareError <= tolerance*0.8
	return report, nil
}

func LatestCalibration(stationID string, reports []CalibrationReport) (CalibrationReport, error) {
	selected := make([]CalibrationReport, 0)
	for _, report := range reports {
		if report.StationID == stationID {
			selected = append(selected, report)
		}
	}
	if len(selected) == 0 {
		return CalibrationReport{}, ErrNotFound
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].CalibratedAt.After(selected[j].CalibratedAt) })
	return selected[0], nil
}
func CalibrationValid(report CalibrationReport, now time.Time) bool {
	return report.Passed && report.DueAt.After(now)
}
