package user

import (
	"context"
	"errors"
	"os"

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
	Repository   *Repository
	client       *gocloak.GoCloak
	clientID     string
	clientSecret string
	realm        string
	access       *common.UserAccess
}

func NewService() *Service {
	var (
		client   = gocloak.NewClient("http://keycloak:8080")
		kcClient = os.Getenv("KC_CLIENT_ID")
		kcSecret = os.Getenv("KC_CLIENT_SECRET")
		kcRealm  = os.Getenv("KC_REALM")
	)
	return &Service{
		client:       client,
		clientID:     kcClient,
		clientSecret: kcSecret,
		realm:        kcRealm,
		Repository:   NewRepository(),
		access:       common.NewUserAccess(),
	}
}

func (s *Service) GetToken(ctx context.Context) (*gocloak.JWT, error) {
	token, err := s.client.LoginClient(ctx, s.clientID, s.clientSecret, s.realm)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		rsp := errors.New("errore di comunicazione [DB-FU]")
		return nil, rsp
	}
	return token, nil
}

func (s *Service) KcCreate(ctx context.Context, userModel User, password string) (uuid *string, err error) {
	token, err := s.GetToken(ctx)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, errors.New("communication error [KC-TK]")
	}

	kcUser := gocloak.User{
		Username:      gocloak.StringP(userModel.Nickname),
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(true),
		FirstName:     (*string)(&userModel.FirstName),
		LastName:      (*string)(&userModel.LastName),
	}

	userID, err := s.client.CreateUser(ctx, token.AccessToken, s.realm, kcUser)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "KC_UC", V: err}, err)
		return nil, errors.New("user already exists")
	}

	if err := s.client.SetPassword(ctx, token.AccessToken, userID, s.realm, password, false); err != nil {
		return nil, errors.New("communication error [KC-SP]")
	}

	uuid = &userID

	return
}

func (s *Service) KcUpdate(ctx context.Context, firstName, lastName *string, uuid string) (err error) {
	token, err := s.GetToken(ctx)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return err
	}

	kcUser := gocloak.User{ID: &uuid}
	if firstName != nil && *firstName != "" {
		kcUser.FirstName = firstName
	}
	if lastName != nil && *lastName != "" {
		kcUser.LastName = lastName
	}

	if err := s.client.UpdateUser(ctx, token.AccessToken, s.realm, kcUser); err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return errors.New("communication error [KC-UU]")
	}

	return nil
}

func (s *Service) KcDelete(ctx context.Context, uuid string) error {
	token, err := s.GetToken(ctx)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return errors.New("communication error [KC-TK]")
	}

	if err := s.client.DeleteUser(ctx, token.AccessToken, s.realm, uuid); err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return errors.New("communication error [KC-DU]")
	}
	return nil
}

func (s *Service) KcGetUser(ctx context.Context, uuid string) (*gocloak.User, error) {
	token, err := s.GetToken(ctx)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, errors.New("communication error [KC-TK]")
	}

	user, err := s.client.GetUserByID(ctx, token.AccessToken, s.realm, uuid)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, errors.New("communication error [KC-GU]")
	}

	return user, nil
}

func (s *Service) KcGetUserGroups(ctx context.Context, uuid string) ([]*gocloak.Group, error) {
	token, err := s.GetToken(ctx)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "KC-TK", V: err}, err)
		return nil, &userService.InternalServerError{Message: "Internal Server error"}
	}

	groups, err := s.client.GetUserGroups(ctx, token.AccessToken, s.realm, uuid, gocloak.GetGroupsParams{})
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "KC-FG", V: err}, err)
		return nil, &userService.InternalServerError{Message: "Internal Server error"}
	}

	for _, group := range groups {
		fullGroup, err := s.client.GetGroup(ctx, token.AccessToken, s.realm, *group.ID)
		if err != nil {
			utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
			continue
		}
		utils.Log.Info(ctx, log.KV{K: "group", V: *fullGroup.Name})
		utils.Log.Info(ctx, log.KV{K: "attributes", V: *fullGroup.Attributes})
	}
	return groups, nil
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
			s.access.Detail = true
			s.access.List = true
			if paid {
				s.access.Edit = true
			}
		case "base":
			s.access.Detail = true
			s.access.List = false
			if paid {
				s.access.Edit = true
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

	groups, err := s.KcGetUserGroups(ctx, claims["sub"].(string))
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
	if !s.access.List {
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
