package user

import (
	"be/internal/database/models"
	"be/internal/utils"
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (r *Repository) FindByID(ctx context.Context, userID string) (*models.User, error) {
	var userModel models.User
	if err := r.DB.WithContext(ctx).Where("id = ?", userID).First(&userModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("errore di comunicazione [DB-FU]")
	}
	return &userModel, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.DB.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) SaveUser(ctx context.Context, userModel models.User) (*models.User, error) {

	if err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(&userModel).Error; err != nil {
			utils.Log.Error(ctx, userModel, err)
			return err
		}
		return nil
	}); err != nil {
		utils.Log.Error(ctx, userModel, err)
		return nil, errors.New("errore di comunicazione [DB-UP]")
	}

	return &userModel, nil
}

func (r *Repository) FindUserWithMembershipsByDate(ctx context.Context, userID string, startDate *string, endDate *string) (*models.User, error) {
	var userModel models.User

	err := r.DB.WithContext(ctx).
		Where("id = ?", userID).
		Preload("Memberships", "start_date BETWEEN ? AND ?", startDate, endDate).
		First(&userModel).Error

	return &userModel, err
}

func (r *Repository) DeleteUser(ctx context.Context, userID string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		userModel, err := r.FindByID(ctx, userID)
		if err != nil {
			return err
		}

		if err := tx.WithContext(ctx).Delete(&userModel).Error; err != nil {
			return errors.New("errore durante l'eliminazione dell'utente")
		}

		return nil
	})
}
