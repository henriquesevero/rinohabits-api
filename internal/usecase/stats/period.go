package stats

import (
	"errors"
	"time"
)

type PeriodType string

const (
	PeriodWeek  PeriodType = "week"
	PeriodMonth PeriodType = "month"
	PeriodYear  PeriodType = "year"
)

var ErrInvalidPeriodType = errors.New("invalid period type")

func periodRange(today time.Time, periodType PeriodType, offset int) (time.Time, time.Time, error) {
	switch periodType {
	case PeriodWeek:
		start := mondayOf(today).AddDate(0, 0, 7*offset)
		end := start.AddDate(0, 0, 6)
		return start, capEnd(end, today, offset), nil
	case PeriodMonth:
		start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, offset, 0)
		end := start.AddDate(0, 1, -1)
		return start, capEnd(end, today, offset), nil
	case PeriodYear:
		start := time.Date(today.Year(), 1, 1, 0, 0, 0, 0, time.UTC).AddDate(offset, 0, 0)
		end := time.Date(start.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
		return start, capEnd(end, today, offset), nil
	default:
		return time.Time{}, time.Time{}, ErrInvalidPeriodType
	}
}

func mondayOf(day time.Time) time.Time {
	offset := int(day.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}
	return day.AddDate(0, 0, -offset)
}

func capEnd(end, today time.Time, offset int) time.Time {
	if offset == 0 && end.After(today) {
		return today
	}
	return end
}

func formatPeriodLabel(periodType PeriodType, start time.Time) string {
	switch periodType {
	case PeriodWeek:
		return start.Format("02/01")
	case PeriodMonth:
		return start.Format("2006-01")
	case PeriodYear:
		return start.Format("2006")
	default:
		return start.Format("2006-01-02")
	}
}

func percentageOf(completed, required int) float64 {
	if required == 0 {
		return 100
	}
	return float64(completed) / float64(required) * 100
}
