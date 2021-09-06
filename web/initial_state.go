package web

import (
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api/model"
)

// user public profile
type Profile struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	FullName       string `json:"fullName"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Country        string `json:"country"`
	EmailConfirmed bool   `json:"emailConfirmed"`
}

type InitialState struct {
	ApplicationName string                `json:"applicationName"`
	ClientID        string                `json:"clientID"`
	UserGroup       string                `json:"usergroup"`
	Token           string                `json:"token"`
	Clients         []config.ClientConfig `json:"clients"`
	Profile         *Profile              `json:"profile"`
}

func NewInitialState(
	cnf *config.Config,
	client *model.Client,
	user *model.User,
	userSession *session.UserSession,
	usergroup string,
) *InitialState {
	accessToken := ""

	if userSession != nil {
		accessToken = userSession.AccessToken
	}

	profile := &Profile{
		ID:             user.ID.String(),
		Email:          user.Username,
		FullName:       user.FullName,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Country:        user.Country,
		EmailConfirmed: user.EmailConfirmed,
	}

	return &InitialState{
		ApplicationName: client.ApplicationName.String,
		ClientID:        client.Key,
		Clients:         cnf.Clients,
		Profile:         profile,
		UserGroup:       usergroup,
		Token:           accessToken,
	}
}
