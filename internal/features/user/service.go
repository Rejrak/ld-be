package user

import (
	"context"
	"errors"
	"os"

	userService "be/gen/user"
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

func (s *Service) OAuth2Auth(ctx context.Context, token string, schema *security.OAuth2Scheme) (context.Context, error) {
	// claims, err := middleware.ValidateToken(token)
	// if err != nil {
	// 	return ctx, err
	// }

	// // Aggiungi i claims nel context, così puoi usarli nei tuoi handler
	// ctx = context.WithValue(ctx, middleware.ClaimsKey, claims)

	return ctx, nil
}

// Create crea un nuovo utente sia in Keycloak che nel database
func (s *Service) Create(ctx context.Context, payload *userService.CreatePayload) (*userService.User, error) {
	// Creazione in Keycloak
	userModel := User{
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

// Get restituisce un utente dal database
func (s *Service) Get(ctx context.Context, payload *userService.GetPayload) (*userService.User, error) {
	user, err := s.Repository.FindByID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
			return nil, &userService.NotFound{Message: "Utente non trovato"}
		}
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, err
	}

	return &userService.User{
		ID:        user.ID.String(),
		KcID:      user.KcID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  &user.Nickname,
		Admin:     user.Admin,
	}, nil
}

// List restituisce un elenco di utenti con paginazione
func (s *Service) List(ctx context.Context, payload *userService.ListPayload) ([]*userService.User, error) {
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

// Update aggiorna un utente nel database e in Keycloak
func (s *Service) Update(ctx context.Context, payload *userService.UpdatePayload) (*userService.User, error) {
	// Recupera l'utente dal database
	user, err := s.Repository.FindByID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
			return nil, &userService.NotFound{Message: "Utente non trovato"}
		}
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, err
	}

	// Aggiorna i dati nel database
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

	// Aggiorna in Keycloak
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

// Delete elimina un utente sia dal database che da Keycloak
func (s *Service) Delete(ctx context.Context, payload *userService.DeletePayload) error {
	// Recupera l'utente dal database
	user, err := s.Repository.FindByID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
			return &userService.NotFound{Message: "Utente non trovato"}
		}
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return err
	}

	// Cancella da Keycloak
	// err = s.KcDelete(ctx, user.KcID.String())
	// if err != nil {
	// 	utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
	// 	return err
	// }

	// Cancella dal database
	err = s.Repository.DeleteUser(ctx, user.ID.String())
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return err
	}

	return nil
}
