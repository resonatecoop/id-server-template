package web

import (
	"net/http"

	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api/model"
)

func (s *Service) profileCommon(r *http.Request) (
	session.ServiceInterface,
	*model.Client,
	*model.User,
	bool,
	*session.UserSession,
	error,
) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		return nil, nil, nil, false, nil, err
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		return nil, nil, nil, false, nil, err
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		return nil, nil, nil, false, nil, err
	}

	// Fetch the user
	user, err := s.oauthService.FindUserByUsername(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, false, nil, err
	}

	// Check if user account is complete
	isUserAccountComplete := s.isUserAccountComplete(userSession)

	return sessionService, client, user, isUserAccountComplete, userSession, nil
}

// isUserAccountComplete checks if user account completeness (email confirmation, ...)
func (s *Service) isUserAccountComplete(userSession *session.UserSession) bool {
	user, err := s.oauthService.FindUserByUsername(userSession.Username)

	if err != nil {
		return false
	}

	// is email address confirmed
	if !user.EmailConfirmed {
		return false
	}

	result, err := s.getUserGroupList(user, userSession.AccessToken)

	if err != nil {
		return false
	}

	if len(result.Usergroup) == 0 {
		return false
	}

	// listeners only need to confirm their email address
	if user.RoleID == int32(model.UserRole) {
		return true
	}

	// if user.FirstName == "" || user.LastName == "" || user.FullName == "" {
	// 	return false
	// }

	if user.Country == "" {
		return false
	}

	return true
}
