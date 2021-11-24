package web

import (
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api-client/models"
	"github.com/resonatecoop/user-api/model"
)

// user public profile
type Profile struct {
	ID             string                                 `json:"id"`
	LegacyID       int32                                  `json:"legacyID"`
	DisplayName    string                                 `json:"displayName"`
	Email          string                                 `json:"email"`
	FullName       string                                 `json:"fullName"`
	FirstName      string                                 `json:"firstName"`
	LastName       string                                 `json:"lastName"`
	Country        string                                 `json:"country"`
	EmailConfirmed bool                                   `json:"emailConfirmed"`
	Complete       bool                                   `json:"complete"`
	Usergroups     []*models.UserUserGroupPrivateResponse `json:"usergroups"`
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
	isUserAccountComplete bool,
	usergroups []*models.UserUserGroupPrivateResponse,
) *InitialState {
	accessToken := ""
	displayName := ""

	if userSession != nil {
		accessToken = userSession.AccessToken
	}

	if len(usergroups) > 0 {
		displayName = usergroups[0].DisplayName
	}

	profile := &Profile{
		ID:             user.ID.String(),
		DisplayName:    displayName,
		Email:          user.Username,
		FullName:       user.FullName,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Country:        user.Country,
		EmailConfirmed: user.EmailConfirmed,
		Complete:       isUserAccountComplete,
		Usergroups:     usergroups,
	}

	return &InitialState{
		ApplicationName: client.ApplicationName.String,
		ClientID:        client.Key,
		Clients:         cnf.Clients,
		Profile:         profile,
		Token:           accessToken,
	}
}
