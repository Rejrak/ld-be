package user

import (
	"be/internal/database/db"
	"be/internal/database/models"
	"context"
	"errors"
	"os"

	userService "be/gen/user"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	Repository   *Repository
	ctx          context.Context
	client       *gocloak.GoCloak
	clientID     string
	clientSecret string
	realm        string
}

func NewService() *Service {
	var (
		beDB     = db.DB.AionDB
		client   = gocloak.NewClient("https://auth.nekos.app")
		kcClient = os.Getenv("KC_CLIENT_ID")
		kcSecret = os.Getenv("KC_CLIENT_SECRET")
		kcRealm  = os.Getenv("KC_REALM")
	)
	return &Service{
		client:       client,
		clientID:     kcClient,
		clientSecret: kcSecret,
		realm:        kcRealm,
		Repository:   NewUserRepository(beDB),
	}
}

func (s *Service) GetToken(ctx context.Context) (*gocloak.JWT, error) {
	token, err := s.client.LoginClient(ctx, s.clientID, s.clientSecret, s.realm)
	if err != nil {
		rsp := errors.New("errore di comunicazione [DB-FU]")
		return nil, rsp
	}
	return token, nil
}

func (s *Service) KcCreate(ctx context.Context, userModel models.User, password string) (uuid *string, err error) {
	token, err := s.GetToken(ctx)
	if err != nil {
		return nil, errors.New("errore di comunicazione [KC-TK]")
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
		return nil, errors.New("user already exists")
	}

	if err := s.client.SetPassword(ctx, token.AccessToken, userID, s.realm, password, false); err != nil {
		return nil, errors.New("errore di comunicazione [KC-SP]")
	}

	uuid = &userID

	return
}

func (s *Service) KcUpdate(ctx context.Context, firstName, lastName *string, uuid string) (err error) {
	token, err := s.GetToken(ctx)
	if err != nil {
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
		return errors.New("errore di comunicazione [KC-UU]")
	}

	return nil
}

func (s *Service) KcDelete(ctx context.Context, uuid string) error {
	token, err := s.GetToken(ctx)
	if err != nil {
		return errors.New("errore di comunicazione [KC-TK]")
	}

	if err := s.client.DeleteUser(ctx, token.AccessToken, s.realm, uuid); err != nil {
		return errors.New("errore di comunicazione [KC-DU]")
	}
	return nil
}

// Create crea un nuovo utente sia in Keycloak che nel database
func (s *Service) Create(ctx context.Context, payload *userService.CreateUserPayload) (*userService.User, error) {
	// Creazione in Keycloak
	userModel := models.User{
		KCID:      uuid.MustParse(payload.KcID),
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Nickname:  *payload.Nickname,
		Admin:     payload.Admin,
	}

	userID, err := s.KcCreate(ctx, userModel, "defaultPassword123!")
	if err != nil {
		return nil, err
	}
	userModel.KCID = uuid.MustParse(*userID)

	// Salvataggio nel database
	savedModel, err := s.Repository.SaveUser(ctx, userModel)
	if err != nil {
		return nil, err
	}

	return &userService.User{
		ID:        savedModel.KCID.String(),
		KcID:      savedModel.KCID.String(),
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
			return nil, &userService.NotFound{Message: "Utente non trovato"}
		}
		return nil, err
	}

	return &userService.User{
		ID:        user.KCID.String(),
		KcID:      user.KCID.String(),
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
		return nil, err
	}

	var response []*userService.User
	for _, user := range users {
		response = append(response, &userService.User{
			ID:        user.KCID.String(),
			KcID:      user.KCID.String(),
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
			return nil, &userService.NotFound{Message: "Utente non trovato"}
		}
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
		return nil, err
	}

	// Aggiorna in Keycloak
	err = s.KcUpdate(ctx, &payload.FirstName, &payload.LastName, payload.KcID)
	if err != nil {
		return nil, err
	}

	return &userService.User{
		ID:        user.KCID.String(),
		KcID:      user.KCID.String(),
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
			return &userService.NotFound{Message: "Utente non trovato"}
		}
		return err
	}

	// Cancella da Keycloak
	err = s.KcDelete(ctx, user.KCID.String())
	if err != nil {
		return err
	}

	// Cancella dal database
	err = s.Repository.DeleteUser(ctx, user.KCID.String())
	if err != nil {
		return err
	}

	return nil
}
