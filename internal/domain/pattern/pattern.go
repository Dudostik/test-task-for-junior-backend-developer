package pattern

import (
	"errors"
	"fmt"
	"time"
)

type PatternType string

const (
	Daily    PatternType = "daily"
	Monthly  PatternType = "monthly"
	Specific PatternType = "specific"
	EvenOdd  PatternType = "even_odd"
)

type EvenOddType string

const (
	Even EvenOddType = "even"
	Odd  EvenOddType = "odd"
)

type TaskPattern struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        PatternType `json:"type"`

	IntervalDays *int `json:"interval_days,omitempty"`

	DayOfMonth *int `json:"day_of_month,omitempty"`

	SpecificDates []time.Time `json:"specific_dates,omitempty"`

	EvenOddType *EvenOddType `json:"even_odd_type,omitempty"`

	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`

	IsActive bool `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *TaskPattern) Validate() error {
	if p.Name == "" {
		return errors.New("pattern name is required")
	}

	switch p.Type {
	case Daily:
		if p.IntervalDays == nil || *p.IntervalDays < 1 {
			return errors.New("daily pattern requires interval_days >= 1")
		}
	case Monthly:
		if p.DayOfMonth == nil || *p.DayOfMonth < 1 || *p.DayOfMonth > 30 {
			return errors.New("monthly pattern requires day_of_month between 1 and 30")
		}
	case Specific:
		if len(p.SpecificDates) == 0 {
			return errors.New("specific pattern requires at least one date")
		}
	case EvenOdd:
		if p.EvenOddType == nil || (*p.EvenOddType != Even && *p.EvenOddType != Odd) {
			return errors.New("even_odd pattern requires valid even_odd_type")
		}
	default:
		return fmt.Errorf("unknown pattern type: %s", p.Type)
	}

	return nil
}

func (p *TaskPattern) GenerateDates(from, to time.Time) []time.Time {
	if !p.IsActive {
		return nil
	}

	start := from
	if p.StartDate != nil && p.StartDate.After(start) {
		start = *p.StartDate
	}

	end := to
	if p.EndDate != nil && p.EndDate.Before(end) {
		end = *p.EndDate
	}

	if start.After(end) {
		return nil
	}

	switch p.Type {
	case Daily:
		return p.generateDaily(start, end)
	case Monthly:
		return p.generateMonthly(start, end)
	case Specific:
		return p.generateSpecific(start, end)
	case EvenOdd:
		return p.generateEvenOdd(start, end)
	default:
		return nil
	}
}

func (p *TaskPattern) generateDaily(start, end time.Time) []time.Time {
	dates := []time.Time{}
	interval := *p.IntervalDays
	current := start

	for !current.After(end) {
		dates = append(dates, current)
		current = current.AddDate(0, 0, interval)
	}

	return dates
}

func (p *TaskPattern) generateMonthly(start, end time.Time) []time.Time {
	dates := []time.Time{}
	day := *p.DayOfMonth
	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !current.After(end) {
		targetDay := day
		lastDayOfMonth := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

		if targetDay > lastDayOfMonth {
			targetDay = lastDayOfMonth
		}

		targetDate := time.Date(current.Year(), current.Month(), targetDay, 0, 0, 0, 0, time.UTC)

		if !targetDate.Before(start) && !targetDate.After(end) {
			dates = append(dates, targetDate)
		}

		current = current.AddDate(0, 1, 0)
	}

	return dates
}

func (p *TaskPattern) generateSpecific(start, end time.Time) []time.Time {
	dates := []time.Time{}

	for _, date := range p.SpecificDates {
		dateUTC := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		startUTC := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
		endUTC := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

		if (dateUTC.Equal(startUTC) || dateUTC.After(startUTC)) && (dateUTC.Equal(endUTC) || dateUTC.Before(endUTC)) {
			dates = append(dates, dateUTC)
		}
	}

	return dates
}

func (p *TaskPattern) generateEvenOdd(start, end time.Time) []time.Time {
	dates := []time.Time{}
	current := start

	for !current.After(end) {
		day := current.Day()
		isEven := day%2 == 0

		shouldInclude := (isEven && *p.EvenOddType == Even) || (!isEven && *p.EvenOddType == Odd)

		if shouldInclude {
			dates = append(dates, current)
		}

		current = current.AddDate(0, 0, 1)
	}

	return dates
}
