package course

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusShelf      Status = "na_prateleira"
	StatusWantToTake Status = "quero_fazer"
	StatusTaking     Status = "fazendo"
	StatusDone       Status = "concluido"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusShelf, StatusWantToTake, StatusTaking, StatusDone:
		return true
	default:
		return false
	}
}

type Course struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Title        string
	Description  string
	Link         string
	Status       Status
	TotalHours   *float64
	CurrentHours float64
	SortOrder    int
	Collection   *string
	CoverURL     *string
	StartedAt    *time.Time
	FinishedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func New(userID uuid.UUID, title, description, link string, totalHours *float64, status Status) *Course {
	return &Course{
		ID:           uuid.New(),
		UserID:       userID,
		Title:        title,
		Description:  description,
		Link:         link,
		Status:       status,
		TotalHours:   totalHours,
		CurrentHours: 0,
	}
}

func (c *Course) RegisterStudy(hoursLogged float64, now time.Time) error {
	if hoursLogged <= 0 {
		return ErrNoProgress
	}

	newHours := c.CurrentHours + hoursLogged
	if c.TotalHours != nil && newHours > *c.TotalHours {
		newHours = *c.TotalHours
	}
	c.CurrentHours = newHours

	if c.Status == StatusWantToTake {
		c.Status = StatusTaking
		c.StartedAt = &now
	}
	if c.TotalHours != nil && c.CurrentHours >= *c.TotalHours {
		c.Status = StatusDone
		c.FinishedAt = &now
	}

	return nil
}

// ChangeStatus applies the status transition and reports whether progress was
// reset (back to shelf/want-to-take), so the caller knows to also discard the
// course's study history — otherwise past course_logs would keep counting
// toward stats for a course that's no longer marked as done or in progress.
func (c *Course) ChangeStatus(newStatus Status, now time.Time) (resetProgress bool, err error) {
	if !newStatus.IsValid() {
		return false, ErrInvalidStatus
	}

	if newStatus == StatusTaking && c.StartedAt == nil {
		c.StartedAt = &now
	}
	if newStatus == StatusDone && c.FinishedAt == nil {
		c.FinishedAt = &now
	}
	if newStatus == StatusShelf || newStatus == StatusWantToTake {
		c.StartedAt = nil
		c.FinishedAt = nil
		c.CurrentHours = 0
		resetProgress = true
	}

	c.Status = newStatus
	return resetProgress, nil
}

func (c *Course) UpdateDetails(title, description, link *string, totalHours *float64, collection *string) {
	if title != nil && *title != "" {
		c.Title = *title
	}
	if description != nil {
		c.Description = *description
	}
	if link != nil {
		c.Link = *link
	}
	if totalHours != nil {
		c.TotalHours = totalHours
	}
	if collection != nil {
		if *collection == "" {
			c.Collection = nil
		} else {
			c.Collection = collection
		}
	}
}
