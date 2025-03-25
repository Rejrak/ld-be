package trainingplan

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

func NewTrainingPlanRepository(db *gorm.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (r *Repository) FindByID(ctx context.Context, id string) (*models.TrainingPlan, error) {
	var tp models.TrainingPlan
	if err := r.DB.WithContext(ctx).
		Preload("User").
		Preload("Workouts").
		Where("id = ?", id).
		First(&tp).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("training plan not found")
		}
		return nil, errors.New("errore di comunicazione [DB-FU]")
	}
	return &tp, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]models.TrainingPlan, error) {
	var tps []models.TrainingPlan
	err := r.DB.WithContext(ctx).
		Preload("User").
		Preload("Workouts").
		Limit(limit).Offset(offset).
		Find(&tps).Error
	if err != nil {
		return nil, err
	}
	return tps, nil
}

func (r *Repository) Save(ctx context.Context, tp models.TrainingPlan) (*models.TrainingPlan, error) {
	if err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(&tp).Error; err != nil {
			utils.Log.Error(ctx, tp, err)
			return err
		}
		return nil
	}); err != nil {
		utils.Log.Error(ctx, tp, err)
		return nil, errors.New("errore di comunicazione [DB-UP]")
	}
	return &tp, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		tp, err := r.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Delete(&tp).Error; err != nil {
			return errors.New("errore durante l'eliminazione del training plan")
		}
		return nil
	})
}
