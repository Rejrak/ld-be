package trainingplan

import (
	"be/internal/database/db"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type TrainingPlan struct {
	ID          uuid.UUID
	Name        string
	Description *string
	StartDate   time.Time
	EndDate     time.Time
	UserID      uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime
}

type Repository struct {
	DB *sql.DB
}

func NewRepository() *Repository {
	return &Repository{DB: db.DB.LD}
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*TrainingPlan, error) {
	query := `SELECT id, name, description, start_date, end_date, user_id FROM training_plan WHERE id = $1 AND deleted_at IS NULL`

	row := r.DB.QueryRowContext(ctx, query, id)

	var tp TrainingPlan
	err := row.Scan(&tp.ID, &tp.Name, &tp.Description, &tp.StartDate, &tp.EndDate, &tp.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &tp, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int, startDate, userID *string) ([]TrainingPlan, error) {
	query := `
	SELECT 
		id, 
		name, 
		description, 
		start_date, 
		end_date, 
		user_id 
	FROM training_plan 
	WHERE 
		deleted_at IS NULL 
			AND 
		($3::uuid IS NULL OR user_id = $3)
			AND
		($4::timestamp IS NULL OR start_date >= $4)
	ORDER BY start_date 
	LIMIT $1 OFFSET $2
	`

	rows, err := r.DB.QueryContext(ctx, query, limit, offset, userID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []TrainingPlan
	for rows.Next() {
		var tp TrainingPlan
		err := rows.Scan(&tp.ID, &tp.Name, &tp.Description, &tp.StartDate, &tp.EndDate, &tp.UserID)
		if err != nil {
			return nil, err
		}
		plans = append(plans, tp)
	}
	return plans, nil
}

func (r *Repository) Save(ctx context.Context, tp TrainingPlan) (*TrainingPlan, error) {
	query := `INSERT INTO training_plan (id, name, description, start_date, end_date, user_id)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				start_date = EXCLUDED.start_date,
				end_date = EXCLUDED.end_date,
				user_id = EXCLUDED.user_id`

	if tp.ID == uuid.Nil {
		tp.ID = uuid.New()
	}

	_, err := r.DB.ExecContext(ctx, query,
		tp.ID, tp.Name, tp.Description, tp.StartDate, tp.EndDate, tp.UserID)

	if err != nil {
		return nil, err
	}

	return &tp, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE training_plan SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err == nil && count == 0 {
		return errors.New("training plan not found or already deleted")
	}

	return nil
}
