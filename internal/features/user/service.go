package user

import (
	"context"
	"errors"

	userService "be/gen/user"
	common "be/internal/features/common"
	"be/internal/middleware"
	"be/internal/utils"

	"github.com/Nerzal/gocloak/v13"
	"goa.design/clue/log"
	"gorm.io/gorm"

	"goa.design/goa/v3/security"
)

type Service struct {
	Repository *Repository
	kc         *common.KcClient
}

func NewService() *Service {
	return &Service{
		Repository: NewRepository(),
		kc:         common.NewKcClient(),
	}
}

func (s *Service) parseUserAccess(groups []*gocloak.Group) {
	for _, group := range groups {
		paid := false
		if group.Attributes != nil {
			if val, ok := (*group.Attributes)["paid"]; ok {
				if val[0] == "1" {
					paid = true
				}
			}
		}

		switch *group.Name {
		case "pro":
			s.kc.UserAccess.Detail = true
			s.kc.UserAccess.List = true
			if paid {
				s.kc.UserAccess.Edit = true
			}
		case "base":
			s.kc.UserAccess.Detail = true
			s.kc.UserAccess.List = false
			if paid {
				s.kc.UserAccess.Edit = true
			}
		}
	}
}

func (s *Service) OAuth2Auth(ctx context.Context, token string, schema *security.OAuth2Scheme) (context.Context, error) {
	claims, err := middleware.ValidateToken(token)
	if err != nil {
		return ctx, err
	}
	for k, v := range claims {
		utils.Log.Debug(ctx, log.KV{K: k, V: v})
	}
	if claims["sub"] == nil {
		return ctx, errors.New("invalid token")
	}

	groups, err := s.kc.KcGetUserGroups(ctx, claims["sub"].(string))
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return ctx, &userService.InternalServerError{Message: "Internal Server error"}
	}
	s.parseUserAccess(groups)

	ctx = context.WithValue(ctx, middleware.ClaimsKey, claims)

	return ctx, nil
}

func (s *Service) Create(ctx context.Context, payload *userService.CreatePayload) (*userService.User, error) {
	// Creazione in Keycloak
	userModel := UserWithPlans{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Nickname:  *payload.Nickname,
		Admin:     payload.Admin,
	}

	// userID, err := s.KcCreate(ctx, userModel, *payload.Password)
	// if err != nil {
	// 	utils.Log.Error(ctx, log.KV{K: "KC-ER", V: err}, err)
	// 	return nil, err
	// }

	// userModel.KcID = uuid.MustParse(*userID)

	// Salvataggio nel database
	savedModel, err := s.Repository.SaveUser(ctx, userModel)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "DB-ERR", V: err}, err)
		return nil, err
	}

	return &userService.User{
		ID:        savedModel.ID.String(),
		KcID:      savedModel.KcID.String(),
		FirstName: savedModel.FirstName,
		LastName:  savedModel.LastName,
		Nickname:  &savedModel.Nickname,
		Admin:     savedModel.Admin,
	}, nil
}

func (s *Service) Get(ctx context.Context, payload *userService.GetPayload) (*userService.UserWithPlans, error) {
	user, err := s.Repository.FindByID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
			return nil, &userService.NotFound{Message: "User not found"}
		}
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, &userService.InternalServerError{Message: "Internal Server error"}
	}

	var trainingPlans []*userService.TrainingPlan
	for _, plan := range user.TrainingPlans {
		trainingPlans = append(trainingPlans, &userService.TrainingPlan{
			ID:          plan.ID.String(),
			Name:        plan.Name,
			Description: plan.Description,
			StartDate:   plan.StartDate.Format("2006-01-02T15:04:05Z"),
			EndDate:     plan.EndDate.Format("2006-01-02T15:04:05Z"),
			UserID:      plan.UserID.String(),
		})
	}

	return &userService.UserWithPlans{
		ID:            user.ID.String(),
		KcID:          user.KcID.String(),
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Nickname:      &user.Nickname,
		Admin:         user.Admin,
		TrainingPlans: trainingPlans,
	}, nil
}

func (s *Service) List(ctx context.Context, payload *userService.ListPayload) ([]*userService.User, error) {
	if !s.kc.UserAccess.List {
		return nil, &userService.Forbidden{Message: "Forbidden"}
	}

	users, err := s.Repository.List(ctx, payload.Limit, payload.Offset)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, err
	}

	var response []*userService.User
	for _, user := range users {
		response = append(response, &userService.User{
			ID:        user.ID.String(),
			KcID:      user.KcID.String(),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Nickname:  &user.Nickname,
			Admin:     user.Admin,
		})
	}
	return response, nil
}

func (s *Service) Update(ctx context.Context, payload *userService.UpdatePayload) (*userService.User, error) {
	user, err := s.Repository.FindByID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
			return nil, &userService.NotFound{Message: "Utente non trovato"}
		}
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, err
	}

	user.FirstName = payload.FirstName
	user.LastName = payload.LastName
	if payload.Nickname != nil {
		user.Nickname = *payload.Nickname
	}
	user.Admin = payload.Admin

	_, err = s.Repository.SaveUser(ctx, *user)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, err
	}

	// update in keycloak
	// err = s.KcUpdate(ctx, &payload.FirstName, &payload.LastName, payload.KcID)
	// if err != nil {
	// 	utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
	// 	return nil, err
	// }

	return &userService.User{
		ID:        user.KcID.String(),
		KcID:      user.KcID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  &user.Nickname,
		Admin:     user.Admin,
	}, nil
}

func (s *Service) Delete(ctx context.Context, payload *userService.DeletePayload) error {
	user, err := s.Repository.FindByID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
			return &userService.NotFound{Message: "Utente non trovato"}
		}
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return err
	}

	// delete from Keycloak
	// err = s.KcDelete(ctx, user.KcID.String())
	// if err != nil {
	// 	utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
	// 	return err
	// }

	err = s.Repository.DeleteUser(ctx, user.ID.String())
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return err
	}

	return nil
}
