package web

import (
	"github.com/RichardKnop/go-oauth2-server/config"
	"github.com/RichardKnop/go-oauth2-server/models"
)

type Profile struct {
	ID             uint64 `json:"id"`
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	EmailConfirmed bool   `json:"emailConfirmed"`
}

type InitialState struct {
	ApplicationName string                `json:"applicationName"`
	ClientID        string                `json:"clientID"`
	Clients         []config.ClientConfig `json:"clients"`
	Profile         *Profile              `json:"profile"`
}

func NewInitialState(
	cnf *config.Config,
	client *models.OauthClient,
	profile *Profile,
) *InitialState {
	return &InitialState{
		ApplicationName: client.ApplicationName.String,
		ClientID:        client.Key,
		Clients:         cnf.Clients,
		Profile:         profile,
	}
}
