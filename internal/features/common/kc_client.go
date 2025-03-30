package common

import (
	"be/internal/utils"

	"context"
	"errors"
	"os"

	"github.com/Nerzal/gocloak/v13"
	"goa.design/clue/log"
)

type KcClient struct {
	clientID     string
	clientSecret string
	realm        string
	client       *gocloak.GoCloak
	UserAccess   *UserAccess
}

func NewKcClient() *KcClient {
	var (
		client   = gocloak.NewClient("http://keycloak:8080")
		kcClient = os.Getenv("KC_CLIENT_ID")
		kcSecret = os.Getenv("KC_CLIENT_SECRET")
		kcRealm  = os.Getenv("KC_REALM")
	)
	return &KcClient{
		client:       client,
		clientID:     kcClient,
		clientSecret: kcSecret,
		realm:        kcRealm,
		UserAccess:   NewUserAccess(),
	}
}

func (s *KcClient) GetToken(ctx context.Context) (*gocloak.JWT, error) {
	token, err := s.client.LoginClient(ctx, s.clientID, s.clientSecret, s.realm)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		rsp := errors.New("errore di comunicazione [DB-FU]")
		return nil, rsp
	}
	return token, nil
}

func (s *KcClient) KcCreate(ctx context.Context, firstName, lastName, nickName, password string) (uuid *string, err error) {
	token, err := s.GetToken(ctx)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "error", V: err}, err)
		return nil, errors.New("communication error [KC-TK]")
	}

	kcUser := gocloak.User{
		Username:      gocloak.StringP(nickName),
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(true),
		FirstName:     gocloak.StringP(firstName),
		LastName:      gocloak.StringP(lastName),
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

func (s *KcClient) KcUpdate(ctx context.Context, firstName, lastName *string, uuid string) (err error) {
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

func (s *KcClient) KcDelete(ctx context.Context, uuid string) error {
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

func (s *KcClient) KcGetUser(ctx context.Context, uuid string) (*gocloak.User, error) {
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

func (s *KcClient) parseUserAccess(groups []*gocloak.Group) {
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
			s.UserAccess.Detail = true
			s.UserAccess.List = true
			if paid {
				s.UserAccess.Edit = true
			}
		case "base":
			s.UserAccess.Detail = true
			s.UserAccess.List = false
			if paid {
				s.UserAccess.Edit = true
			}
		}
	}
}

func (s *KcClient) KcGetUserGroups(ctx context.Context, uuid string) ([]*gocloak.Group, error) {
	token, err := s.GetToken(ctx)
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "KC-TK", V: err}, err)
		return nil, errors.New("communication error [KC-GT]")
	}

	groups, err := s.client.GetUserGroups(ctx, token.AccessToken, s.realm, uuid, gocloak.GetGroupsParams{})
	if err != nil {
		utils.Log.Error(ctx, log.KV{K: "KC-FG", V: err}, err)
		return nil, errors.New("communication error [KC-GG]")
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
