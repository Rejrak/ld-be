package workout

import (
	"be/internal/database/db"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Workout struct {
	ID             uuid.UUID
	Name           string
	TrainingPlanID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      sql.NullTime
}

type Repository struct {
	DB *sql.DB
}

func NewRepository() *Repository {
	return &Repository{DB: db.DB.LD}
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Workout, error) {
	query := `
	SELECT id, name, training_plan_id, created_at, updated_at
	FROM workout
	WHERE id = $1 AND deleted_at IS NULL`

	var w Workout
	err := r.DB.QueryRowContext(ctx, query, id).
		Scan(&w.ID, &w.Name, &w.TrainingPlanID, &w.CreatedAt, &w.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("workout not found")
		}
		return nil, err
	}

	return &w, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int, trainingPlanID *string) ([]Workout, error) {
	query := `
	SELECT id, name, training_plan_id, created_at, updated_at
	FROM workout
	WHERE deleted_at IS NULL
	  AND ($3::uuid IS NULL OR training_plan_id = $3)
	ORDER BY created_at DESC
	LIMIT $1 OFFSET $2`

	rows, err := r.DB.QueryContext(ctx, query, limit, offset, trainingPlanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var w Workout
		err := rows.Scan(&w.ID, &w.Name, &w.TrainingPlanID, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, err
		}
		workouts = append(workouts, w)
	}
	return workouts, nil
}

func (r *Repository) Save(ctx context.Context, w Workout) (*Workout, error) {
	query := `
	INSERT INTO workout (id, name, training_plan_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		training_plan_id = EXCLUDED.training_plan_id,
		updated_at = NOW()`

	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}

	_, err := r.DB.ExecContext(ctx, query, w.ID, w.Name, w.TrainingPlanID)
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE workout SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err == nil && count == 0 {
		return errors.New("workout not found or already deleted")
	}

	return nil
}
