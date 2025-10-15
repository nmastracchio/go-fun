package filter

import (
	"fmt"
	"strconv"
	"time"
)

type FilterMode int

const (
	ModeInvalid FilterMode = iota
	ModeToday
	ModeOverdue
	ModeNextNDays
)

type TaskDueFilter struct {
	Mode FilterMode
	Days int // for ModeNextNDays
}

func NewTaskDueFilter(input string) (TaskDueFilter, error) {
	switch input {
	case "today":
		return TaskDueFilter{Mode: ModeToday}, nil
	case "overdue":
		return TaskDueFilter{Mode: ModeOverdue}, nil
	case "week":
		return TaskDueFilter{Mode: ModeNextNDays, Days: 7}, nil
	default:
		days, err := strconv.Atoi(input)
		if err != nil {
			return TaskDueFilter{}, fmt.Errorf("invalid filter flag: %q", input)
		}
		if days < 0 {
			return TaskDueFilter{}, fmt.Errorf("days cannot be negative: %d", days)
		}
		return TaskDueFilter{Mode: ModeNextNDays, Days: days}, nil
	}
}

func (f *TaskDueFilter) Matches(date time.Time) bool {
	switch f.Mode {
	case ModeToday:
		return date.Format(time.DateOnly) == time.Now().Format(time.DateOnly)
	case ModeOverdue:
		return date.Before(time.Now())
	case ModeNextNDays:
		return date.After(time.Now()) && date.Before(time.Now().AddDate(0, 0, f.Days))
	}
	return false
}
