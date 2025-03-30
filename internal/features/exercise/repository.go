package exercise

import (
	"be/internal/database/db"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Exercise struct {
	ID        uuid.UUID
	Name      string
	WorkoutID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
}

type Repository struct {
	DB *sql.DB
}

func NewRepository() *Repository {
	return &Repository{DB: db.DB.LD}
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Exercise, error) {
	query := `
	SELECT id, name, workout_id, created_at, updated_at
	FROM exercise
	WHERE id = $1 AND deleted_at IS NULL`

	var ex Exercise
	err := r.DB.QueryRowContext(ctx, query, id).
		Scan(&ex.ID, &ex.Name, &ex.WorkoutID, &ex.CreatedAt, &ex.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("exercise not found")
		}
		return nil, err
	}

	return &ex, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int, workoutID *string) ([]Exercise, error) {
	query := `
	SELECT id, name, workout_id, created_at, updated_at
	FROM exercise
	WHERE deleted_at IS NULL
	  AND ($3::uuid IS NULL OR workout_id = $3)
	ORDER BY created_at DESC
	LIMIT $1 OFFSET $2`

	rows, err := r.DB.QueryContext(ctx, query, limit, offset, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []Exercise
	for rows.Next() {
		var ex Exercise
		err := rows.Scan(&ex.ID, &ex.Name, &ex.WorkoutID, &ex.CreatedAt, &ex.UpdatedAt)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, ex)
	}
	return exercises, nil
}

func (r *Repository) Save(ctx context.Context, ex Exercise) (*Exercise, error) {
	query := `
	INSERT INTO exercise (id, name, workout_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		workout_id = EXCLUDED.workout_id,
		updated_at = NOW()`

	if ex.ID == uuid.Nil {
		ex.ID = uuid.New()
	}

	_, err := r.DB.ExecContext(ctx, query, ex.ID, ex.Name, ex.WorkoutID)
	if err != nil {
		return nil, err
	}

	return &ex, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE exercise SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err == nil && count == 0 {
		return errors.New("exercise not found or already deleted")
	}

	return nil
}
