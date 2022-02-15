package web

import (
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api-client/models"
	"github.com/resonatecoop/user-api/model"
)

// Profile user public profile
type Profile struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	LegacyID    int32  `json:"legacyID"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Credits     string `json:"credits"`
	// FullName               string                                 `json:"fullName"`
	// FirstName              string                                 `json:"firstName"`
	// LastName               string                                 `json:"lastName"`
	Country                string                                 `json:"country"`
	NewsletterNotification bool                                   `json:"newsletterNotification"`
	EmailConfirmed         bool                                   `json:"emailConfirmed"`
	Member                 bool                                   `json:"member"`
	Complete               bool                                   `json:"complete"`
	Usergroups             []*models.UserUserGroupPrivateResponse `json:"usergroups"`
}

// NewProfile
func NewProfile(
	user *model.User,
	usergroups []*models.UserUserGroupPrivateResponse,
	isUserAccountComplete bool,
	credits string,
	role string,
) *Profile {
	displayName := ""

	if len(usergroups) > 0 {
		displayName = usergroups[0].DisplayName
	}

	return &Profile{
		ID:                     user.ID.String(),
		Complete:               isUserAccountComplete,
		Country:                user.Country,
		Credits:                credits,
		DisplayName:            displayName,
		Role:                   role,
		Email:                  user.Username,
		EmailConfirmed:         user.EmailConfirmed,
		LegacyID:               user.LegacyID,
		Member:                 user.Member,
		NewsletterNotification: user.NewsletterNotification,
		Usergroups:             usergroups,
	}
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
	CountryList     []Country             `json:"countries"`
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
	countryList []Country,
) *InitialState {
	accessToken := ""

	if userSession != nil {
		accessToken = userSession.AccessToken
	}

	profile := NewProfile(
		user,
		usergroups,
		isUserAccountComplete,
		credits,
		userSession.Role,
	)

	if len(usergroups) > 0 {
		profile.DisplayName = usergroups[0].DisplayName
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
		CountryList:     countryList,
	}
}
