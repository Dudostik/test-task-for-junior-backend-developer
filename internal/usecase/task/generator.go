package task

import (
	"context"
	"fmt"
	"time"

	patterndomain "example.com/taskservice/internal/domain/pattern"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type PatternRepository interface {
	GetByID(ctx context.Context, id int64) (*patterndomain.TaskPattern, error)
	List(ctx context.Context) ([]patterndomain.TaskPattern, error)
}

type TaskGenerator struct {
	taskRepo    Repository
	patternRepo PatternRepository
	now         func() time.Time
}

func NewTaskGenerator(taskRepo Repository, patternRepo PatternRepository) *TaskGenerator {
	return &TaskGenerator{
		taskRepo:    taskRepo,
		patternRepo: patternRepo,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func (g *TaskGenerator) GenerateForPattern(ctx context.Context, patternID int64, from, to time.Time) ([]*taskdomain.Task, error) {
	pattern, err := g.patternRepo.GetByID(ctx, patternID)

	if err != nil {
		return nil, err
	}

	if !pattern.IsActive {
		return nil, nil
	}

	dates := pattern.GenerateDates(from, to)
	created := []*taskdomain.Task{}

	for _, date := range dates {
		task := &taskdomain.Task{
			Title:        fmt.Sprintf("%s (%s)", pattern.Name, date.Format("2006-01-02")),
			Description:  pattern.Description,
			Status:       taskdomain.StatusNew,
			CreatedAt:    g.now(),
			UpdatedAt:    g.now(),
			PatternID:    &patternID,
			IsGenerated:  true,
			OriginalDate: &date,
		}

		createdTask, err := g.taskRepo.Create(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("failed to create task for date %s: %w", date, err)
		}
		created = append(created, createdTask)
	}

	return created, nil
}

func (g *TaskGenerator) GenerateAllActive(ctx context.Context, from, to time.Time) (map[int64][]*taskdomain.Task, error) {
	patterns, err := g.patternRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	results := make(map[int64][]*taskdomain.Task)
	for _, pattern := range patterns {
		if !pattern.IsActive {
			continue
		}
		tasks, err := g.GenerateForPattern(ctx, pattern.ID, from, to)
		if err != nil {
			// logging
			continue
		}
		if len(tasks) > 0 {
			results[pattern.ID] = tasks
		}
	}

	return results, nil
}
