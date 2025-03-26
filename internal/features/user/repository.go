package user

import (
	"be/internal/database/db"
	"be/internal/utils"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	KcID      uuid.UUID
	FirstName string
	LastName  string
	Nickname  string
	Admin     bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository struct {
	DB *sql.DB
}

func NewRepository() *Repository {
	return &Repository{DB: db.DB.LD}
}

func (r *Repository) FindByID(ctx context.Context, userID string) (*User, error) {
	query := `SELECT id, kc_id, first_name, last_name, nickname, admin, created_at, updated_at 
		  FROM users WHERE id = $1 AND deleted_at IS NULL`

	var user User
	err := r.DB.QueryRowContext(ctx, query, userID).
		Scan(&user.ID, &user.KcID, &user.FirstName, &user.LastName, &user.Nickname, &user.Admin, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("errore di comunicazione [DB-FU]")
	}

	return &user, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]User, error) {
	query := `SELECT id, kc_id, first_name, last_name, nickname, admin, created_at, updated_at 
		  FROM users WHERE deleted_at IS NULL
		  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.KcID, &user.FirstName, &user.LastName, &user.Nickname, &user.Admin, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *Repository) SaveUser(ctx context.Context, user User) (*User, error) {
	query := `
		INSERT INTO users (id, kc_id, first_name, last_name, nickname, admin, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			kc_id = EXCLUDED.kc_id,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			nickname = EXCLUDED.nickname,
			admin = EXCLUDED.admin,
			updated_at = EXCLUDED.updated_at
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now()
	user.UpdatedAt = now
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}

	_, err := r.DB.ExecContext(ctx, query,
		user.ID, nil, user.FirstName, user.LastName, user.Nickname, user.Admin, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		utils.Log.Error(ctx, user, err)
		return nil, errors.New("errore di comunicazione [DB-UP]")
	}

	return &user, nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID string) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.DB.ExecContext(ctx, query, userID)
	if err != nil {
		utils.Log.Error(ctx, query, err)
		return err
	}

	count, err := res.RowsAffected()
	if err == nil && count == 0 {
		utils.Log.Debug(ctx, err)
		utils.Log.Error(ctx, res, err)
		return errors.New("user not found or already deleted")
	}

	return nil
}
