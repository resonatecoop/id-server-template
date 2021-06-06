package web

import (
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/user-api/model"
)

type Profile struct {
	ID             uint64 `json:"id"`
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	Country        string `json:"country"`
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
	client *model.Client,
	profile *Profile,
) *InitialState {
	return &InitialState{
		ApplicationName: client.ApplicationName.String,
		ClientID:        client.Key,
		Clients:         cnf.Clients,
		Profile:         profile,
	}
}
