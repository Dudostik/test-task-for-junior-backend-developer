package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	patterndomain "example.com/taskservice/internal/domain/pattern"
)

type PatternRepository struct {
	pool *pgxpool.Pool
}

func NewPatternRepository(pool *pgxpool.Pool) *PatternRepository {
	return &PatternRepository{pool: pool}
}

func (r *PatternRepository) Create(ctx context.Context, p *patterndomain.TaskPattern) (*patterndomain.TaskPattern, error) {
	tx, err := r.pool.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	query := `
        INSERT INTO task_patterns (name, description, type, interval_days, day_of_month, even_odd_type, start_date, end_date, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id, created_at, updated_at
    `

	now := time.Now().UTC()
	err = tx.QueryRow(ctx, query,
		p.Name, p.Description, p.Type,
		p.IntervalDays, p.DayOfMonth, p.EvenOddType,
		p.StartDate, p.EndDate, p.IsActive,
		now, now,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if p.Type == patterndomain.Specific && len(p.SpecificDates) > 0 {
		for _, date := range p.SpecificDates {
			_, err := tx.Exec(ctx, `
			    INSERT INTO pattern_specific_dates (pattern_id, specific_date)
                VALUES ($1, $2)
			`, p.ID, date)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *PatternRepository) GetByID(ctx context.Context, id int64) (*patterndomain.TaskPattern, error) {
	query := `
        SELECT id, name, description, type, interval_days, day_of_month, even_odd_type, 
               start_date, end_date, is_active, created_at, updated_at
        FROM task_patterns
        WHERE id = $1
    `

	var p patterndomain.TaskPattern
	var intervalDays, dayOfMonth sql.NullInt32
	var evenOddType sql.NullString
	var startDate, endDate sql.NullTime

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Type,
		&intervalDays, &dayOfMonth, &evenOddType,
		&startDate, &endDate, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, patterndomain.ErrNotFound
		}

		return nil, err
	}

	if intervalDays.Valid {
		val := int(intervalDays.Int32)
		p.IntervalDays = &val
	}
	if dayOfMonth.Valid {
		val := int(dayOfMonth.Int32)
		p.DayOfMonth = &val
	}
	if evenOddType.Valid {
		val := patterndomain.EvenOddType(evenOddType.String)
		p.EvenOddType = &val
	}
	if startDate.Valid {
		p.StartDate = &startDate.Time
	}
	if endDate.Valid {
		p.EndDate = &endDate.Time
	}

	if p.Type == patterndomain.Specific {
		rows, err := r.pool.Query(ctx, `
            SELECT specific_date FROM pattern_specific_dates WHERE pattern_id = $1 ORDER BY specific_date
        `, id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var date time.Time
			if err := rows.Scan(&date); err != nil {
				return nil, err
			}
			p.SpecificDates = append(p.SpecificDates, date)
		}
	}

	return &p, nil
}

func (r *PatternRepository) Update(ctx context.Context, p *patterndomain.TaskPattern) (*patterndomain.TaskPattern, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
        UPDATE task_patterns
        SET name = $1, description = $2, type = $3, 
            interval_days = $4, day_of_month = $5, even_odd_type = $6,
            start_date = $7, end_date = $8, is_active = $9,
            updated_at = $10
        WHERE id = $11
        RETURNING updated_at
    `

	now := time.Now().UTC()
	err = tx.QueryRow(ctx, query,
		p.Name, p.Description, p.Type,
		p.IntervalDays, p.DayOfMonth, p.EvenOddType,
		p.StartDate, p.EndDate, p.IsActive,
		now, p.ID,
	).Scan(&p.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, patterndomain.ErrNotFound
		}
		return nil, err
	}

	// Обновляем specific даты (удаляем старые, вставляем новые)
	if p.Type == patterndomain.Specific {
		_, err = tx.Exec(ctx, `DELETE FROM pattern_specific_dates WHERE pattern_id = $1`, p.ID)
		if err != nil {
			return nil, err
		}

		for _, date := range p.SpecificDates {
			_, err := tx.Exec(ctx, `
                INSERT INTO pattern_specific_dates (pattern_id, specific_date)
                VALUES ($1, $2)
            `, p.ID, date)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *PatternRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM task_patterns WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return patterndomain.ErrNotFound
	}
	return nil
}

func (r *PatternRepository) List(ctx context.Context) ([]patterndomain.TaskPattern, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, name, description, type, is_active, created_at, updated_at
        FROM task_patterns
        ORDER BY id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patterns := []patterndomain.TaskPattern{}
	for rows.Next() {
		var p patterndomain.TaskPattern
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Type, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}

	return patterns, nil
}
