package pattern

import (
	"context"
	"fmt"
	"time"

	patterndomain "example.com/taskservice/internal/domain/pattern"
)

type Repository interface {
	Create(ctx context.Context, p *patterndomain.TaskPattern) (*patterndomain.TaskPattern, error)
	GetByID(ctx context.Context, id int64) (*patterndomain.TaskPattern, error)
	Update(ctx context.Context, p *patterndomain.TaskPattern) (*patterndomain.TaskPattern, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]patterndomain.TaskPattern, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreatePatternInput) (*patterndomain.TaskPattern, error)
	GetByID(ctx context.Context, id int64) (*patterndomain.TaskPattern, error)
	Update(ctx context.Context, id int64, input UpdatePatternInput) (*patterndomain.TaskPattern, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]patterndomain.TaskPattern, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreatePatternInput struct {
	Name          string
	Description   string
	Type          patterndomain.PatternType
	IntervalDays  *int
	DayOfMonth    *int
	SpecificDates []time.Time
	EvenOddType   *patterndomain.EvenOddType
	StartDate     *time.Time
	EndDate       *time.Time
	IsActive      bool
}

func (s *Service) Create(ctx context.Context, input CreatePatternInput) (*patterndomain.TaskPattern, error) {
	pattern := &patterndomain.TaskPattern{
		Name:          input.Name,
		Description:   input.Description,
		Type:          input.Type,
		IntervalDays:  input.IntervalDays,
		DayOfMonth:    input.DayOfMonth,
		SpecificDates: input.SpecificDates,
		EvenOddType:   input.EvenOddType,
		StartDate:     input.StartDate,
		EndDate:       input.EndDate,
		IsActive:      input.IsActive,
	}

	if err := pattern.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	return s.repo.Create(ctx, pattern)
}

type UpdatePatternInput struct {
	Name          *string
	Description   *string
	Type          *patterndomain.PatternType
	IntervalDays  *int
	DayOfMonth    *int
	SpecificDates []time.Time
	EvenOddType   *patterndomain.EvenOddType
	StartDate     *time.Time
	EndDate       *time.Time
	IsActive      *bool
}

func (s *Service) Update(ctx context.Context, id int64, input UpdatePatternInput) (*patterndomain.TaskPattern, error) {
	existing, err := s.repo.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}
	if input.Type != nil {
		existing.Type = *input.Type
	}
	if input.IntervalDays != nil {
		existing.IntervalDays = input.IntervalDays
	}
	if input.DayOfMonth != nil {
		existing.DayOfMonth = input.DayOfMonth
	}
	if input.SpecificDates != nil {
		existing.SpecificDates = input.SpecificDates
	}
	if input.EvenOddType != nil {
		existing.EvenOddType = input.EvenOddType
	}
	if input.StartDate != nil {
		existing.StartDate = input.StartDate
	}
	if input.EndDate != nil {
		existing.EndDate = input.EndDate
	}
	if input.IsActive != nil {
		existing.IsActive = *input.IsActive
	}

	if err := existing.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	return s.repo.Update(ctx, existing)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*patterndomain.TaskPattern, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid pattern id", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid pattern id", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]patterndomain.TaskPattern, error) {
	return s.repo.List(ctx)
}
