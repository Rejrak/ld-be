package trainingplan

import (
	trainingplanService "be/gen/training_plan"
	"be/internal/database/db"
	"be/internal/database/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	Repository *Repository
}

func NewService() *Service {
	return &Service{
		Repository: NewTrainingPlanRepository(db.DB.LD),
	}
}

func parseDate(date string) (time.Time, error) {
	return time.Parse(time.RFC3339, date)

}

func (s *Service) Create(ctx context.Context, payload *trainingplanService.CreateTrainingPlanPayload) (*trainingplanService.TrainingPlan, error) {
	startDate, err := parseDate(payload.StartDate)
	if err != nil {
		return nil, errors.New("invalid startDate format")
	}

	endDate, err := parseDate(payload.StartDate)
	if err != nil {
		return nil, errors.New("invalid endDate format")
	}

	tp := models.TrainingPlan{
		Name:        payload.Name,
		Description: *payload.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		UserID:      uuid.MustParse(payload.UserID),
	}

	saved, err := s.Repository.Save(ctx, tp)
	if err != nil {
		return nil, err
	}

	return &trainingplanService.TrainingPlan{
		ID:          saved.ID.String(),
		Name:        saved.Name,
		Description: &saved.Description,
		StartDate:   saved.StartDate.Format(time.RFC3339),
		EndDate:     saved.EndDate.Format(time.RFC3339),
		UserID:      saved.UserID.String(),
	}, nil
}

func (s *Service) Get(ctx context.Context, payload *trainingplanService.GetPayload) (*trainingplanService.TrainingPlan, error) {
	tp, err := s.Repository.FindByID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &trainingplanService.NotFound{Message: "Piano non trovato"}
		}
		return nil, err
	}

	return &trainingplanService.TrainingPlan{
		ID:          tp.ID.String(),
		Name:        tp.Name,
		Description: &tp.Description,
		StartDate:   tp.StartDate.Format(time.RFC3339),
		EndDate:     tp.EndDate.Format(time.RFC3339),
		UserID:      tp.UserID.String(),
	}, nil
}

func (s *Service) List(ctx context.Context) ([]*trainingplanService.TrainingPlan, error) {
	tps, err := s.Repository.List(ctx, 100, 0)
	if err != nil {
		return nil, err
	}

	var res []*trainingplanService.TrainingPlan
	for _, tp := range tps {
		res = append(res, &trainingplanService.TrainingPlan{
			ID:          tp.ID.String(),
			Name:        tp.Name,
			Description: &tp.Description,
			StartDate:   tp.StartDate.Format(time.RFC3339),
			EndDate:     tp.EndDate.Format(time.RFC3339),
			UserID:      tp.UserID.String(),
		})
	}
	return res, nil
}

func (s *Service) Update(ctx context.Context, payload *trainingplanService.UpdatePayload) (*trainingplanService.TrainingPlan, error) {
	tp, err := s.Repository.FindByID(ctx, payload.ID)
	if err != nil {
		return nil, err
	}

	startDate, err := parseDate(payload.StartDate)
	if err != nil {
		return nil, errors.New("invalid startDate format")
	}

	endDate, err := parseDate(payload.StartDate)
	if err != nil {
		return nil, errors.New("invalid endDate format")
	}

	tp.Name = payload.Name
	tp.Description = *payload.Description
	tp.StartDate = startDate
	tp.EndDate = endDate
	tp.UserID = uuid.MustParse(payload.UserID)

	saved, err := s.Repository.Save(ctx, *tp)
	if err != nil {
		return nil, err
	}

	return &trainingplanService.TrainingPlan{
		ID:          saved.ID.String(),
		Name:        saved.Name,
		Description: &saved.Description,
		StartDate:   saved.StartDate.Format(time.RFC3339),
		EndDate:     saved.EndDate.Format(time.RFC3339),
		UserID:      saved.UserID.String(),
	}, nil
}

func (s *Service) Delete(ctx context.Context, payload *trainingplanService.DeletePayload) error {
	return s.Repository.Delete(ctx, payload.ID)
}
