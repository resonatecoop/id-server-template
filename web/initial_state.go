package web

import (
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api-client/models"
	"github.com/resonatecoop/user-api/model"
)

// user public profile
type Profile struct {
	ID                     string                                 `json:"id"`
	LegacyID               int32                                  `json:"legacyID"`
	DisplayName            string                                 `json:"displayName"`
	Email                  string                                 `json:"email"`
	Credits                string                                 `json:"credits"`
	FullName               string                                 `json:"fullName"`
	FirstName              string                                 `json:"firstName"`
	LastName               string                                 `json:"lastName"`
	Country                string                                 `json:"country"`
	NewsletterNotification bool                                   `json:"newsletterNotification"`
	EmailConfirmed         bool                                   `json:"emailConfirmed"`
	Member                 bool                                   `json:"member"`
	Complete               bool                                   `json:"complete"`
	Usergroups             []*models.UserUserGroupPrivateResponse `json:"usergroups"`
}

type InitialState struct {
	ApplicationName string                `json:"applicationName"`
	ClientID        string                `json:"clientID"`
	UserGroup       string                `json:"usergroup"`
	Token           string                `json:"token"`
	Clients         []config.ClientConfig `json:"clients"`
	Profile         *Profile              `json:"profile"`
	Memberships     []Membership          `json:"memberships"`
	Shares          []Share               `json:"shares"`
	Products        []Product             `json:"products"`
	CSRFToken       string                `json:"csrfToken"`
}

func NewInitialState(
	cnf *config.Config,
	client *model.Client,
	user *model.User,
	userSession *session.UserSession,
	isUserAccountComplete bool,
	credits string,
	usergroups []*models.UserUserGroupPrivateResponse,
	memberships []Membership,
	shares []Share,
	products []Product,
	csrfToken string,
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
		ID:                     user.ID.String(),
		DisplayName:            displayName,
		Email:                  user.Username,
		Member:                 user.Member,
		FullName:               user.FullName,
		FirstName:              user.FirstName,
		LastName:               user.LastName,
		Credits:                credits,
		Country:                user.Country,
		NewsletterNotification: user.NewsletterNotification,
		EmailConfirmed:         user.EmailConfirmed,
		Complete:               isUserAccountComplete,
		Usergroups:             usergroups,
	}

	return &InitialState{
		ApplicationName: client.ApplicationName.String,
		ClientID:        client.Key,
		Clients:         cnf.Clients,
		Profile:         profile,
		Token:           accessToken,
		Memberships:     memberships,
		Shares:          shares,
		Products:        products,
		CSRFToken:       csrfToken,
	}
}
