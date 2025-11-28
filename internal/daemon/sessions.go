package daemon

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
	"github.com/thand-io/agent/internal/models"
	"github.com/thand-io/agent/internal/sessions"
)

// postSession creates a new session
//
//	@Summary		Create a new session
//	@Description	Create a new session with the provided session token
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Param			session	body		models.SessionCreateRequest	true	"Session creation request"
//	@Success		200		{object}	map[string]any		"Session created successfully"
//	@Failure		400		{object}	map[string]any		"Bad request"
//	@Failure		500		{object}	map[string]any		"Internal server error"
//	@Router			/sessions [post]
func (s *Server) postSession(c *gin.Context) {

	// This is an un-authenticated endpoint to create a session
	// but only allowed in agent mode. The code provided must match
	// the code we issued when starting the agent. This will prevent
	// unauthorised session creation.

	if !s.Config.IsAgent() && !s.Config.IsClient() {
		s.getErrorPage(c, http.StatusBadRequest, "Session creation can only be called in agent mode")
		return
	}

	// Get the post JSON Body as a Session create request
	// which is a struct with fields for session creation
	var sessionCreateRequest models.SessionCreateRequest
	if err := c.ShouldBindJSON(&sessionCreateRequest); err != nil {
		s.getErrorPage(c, http.StatusBadRequest, "Failed to parse request body", err)
		return
	}

	// Validate the code we sent matches the expected code
	if len(sessionCreateRequest.Code) == 0 {
		s.getErrorPage(c, http.StatusBadRequest, "Session code is required")
		return
	}

	// We need to decrypt the code to check we issued it.
	if !s.Config.GetServices().HasEncryption() {
		s.getErrorPage(c, http.StatusInternalServerError, "Encryption service is not configured")
		return
	}

	sessionCode := sessionCreateRequest.Code

	// If the code decrypts then we're all good.
	codeResponse, err := models.EncodingWrapper{
		Type: models.ENCODED_SESSION_CODE,
	}.DecodeAndDecrypt(
		sessionCode,
		s.Config.GetServices().GetEncryption(),
	)

	if err != nil {
		s.getErrorPage(c, http.StatusBadRequest, "Failed to decrypt session code", err)
		return
	}

	codeWrapper := models.CodeWrapper{}
	err = common.ConvertInterfaceToInterface(codeResponse.Data, &codeWrapper)

	if err != nil {
		s.getErrorPage(c, http.StatusBadRequest, "Invalid session code data")
		return
	}

	// Validate the code is still valid
	if !codeWrapper.IsValid(s.Config.GetLoginServerUrl()) {
		s.getErrorPage(c, http.StatusBadRequest, "Session code is invalid or expired")
		return
	}

	sessionToken := sessionCreateRequest.Session

	// The session token is an encoded local session
	// The payload is encrypted - however, the decode
	// call does not require decryption as the data is
	// already encrypted within the session token.
	sessionData, err := models.EncodingWrapper{
		Type: models.ENCODED_SESSION_LOCAL,
	}.Decode(sessionToken)

	if err != nil {
		s.getErrorPage(c, http.StatusBadRequest, "Failed to decode session token", err)
		return
	}

	var session models.LocalSession
	err = common.ConvertMapToInterface(sessionData.Data.(map[string]any), &session)

	if err != nil {
		s.getErrorPage(c, http.StatusBadRequest, "Failed to convert session data", err)
		return
	}

	// Now lets store the session in the users local session manager.
	sessionManager := sessions.GetSessionManager()
	err = sessionManager.AddSession(
		s.Config.GetLoginServerHostname(),
		sessionCreateRequest.Provider,
		session,
	)

	if err != nil {
		s.getErrorPage(c, http.StatusInternalServerError, "Failed to store session", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session created successfully",
		"expiry":  session.Expiry.UTC(),
	})
}

// getSessions retrieves all sessions
//
//	@Summary		Get all sessions
//	@Description	Retrieve all active sessions for the user
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	sessions.LoginServer		"List of sessions"
//	@Failure		400	{object}	map[string]any	"Bad request"
//	@Failure		500	{object}	map[string]any	"Internal server error"
//	@Router			/sessions [get]
func (s *Server) getSessions(c *gin.Context) {

	if s.Config.IsServer() {

		remoteSessions, err := s.getUserSessions(c)

		if err != nil {
			s.getErrorPage(c, http.StatusBadRequest, "Failed to get user sessions", err)
			return
		}

		foundSessions := map[string]models.LocalSession{}

		// Convert to response format
		for providerName, session := range remoteSessions {
			foundSessions[providerName] = models.LocalSession{
				Version: 1,
				Expiry:  session.Expiry,
			}
		}

		sessionsList := sessions.LoginServer{
			Version:   "1",
			Timestamp: time.Now(),
			Sessions:  foundSessions,
		}

		c.JSON(http.StatusOK, sessionsList)
		return

	} else if s.Config.IsAgent() {

		loginServer := s.Config.GetLoginServerHostname()

		logrus.WithFields(logrus.Fields{
			"loginServer": loginServer,
		}).Debugln("Fetching sessions")

		sessionManager := sessions.GetSessionManager()
		sessionManager.Load(loginServer)
		sessionsList, err := sessionManager.GetLoginServer(loginServer)

		if err != nil {
			s.getErrorPage(c, http.StatusInternalServerError, "Failed to list sessions", err)
			return
		}

		c.JSON(http.StatusOK, sessionsList)
		return

	} else {

		s.getErrorPage(c, http.StatusBadRequest, "Get sessions can only be called in agent or server mode")
		return
	}
}

// getSessionByProvider retrieves a session for a specific provider
//
//	@Summary		Get session by provider
//	@Description	Retrieve session information for a specific provider
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string					true	"Provider name"
//	@Success		200			{object}	map[string]any	"Session information"
//	@Failure		400			{object}	map[string]any	"Bad request"
//	@Failure		404			{object}	map[string]any	"Session not found"
//	@Failure		500			{object}	map[string]any	"Internal server error"
//	@Router			/session/{provider} [get]
func (s *Server) getSessionByProvider(c *gin.Context) {

	provider := c.Param("provider")
	if len(provider) == 0 {
		s.getErrorPage(c, http.StatusBadRequest, "Provider is required")
		return
	}

	loginServer := s.Config.GetLoginServerHostname()

	logrus.WithFields(logrus.Fields{
		"loginServer": loginServer,
		"provider":    provider,
	}).Debugln("Fetching session for provider")

	sessionManager := sessions.GetSessionManager()
	sessionManager.Load(loginServer)
	session, err := sessionManager.GetSession(loginServer, provider)

	if err != nil {
		s.getErrorPage(c, http.StatusInternalServerError, "Failed to get session", err)
		return
	}

	if session == nil {
		s.getErrorPage(c, http.StatusNotFound, "Session not found for provider")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session": session,
	})
}

// deleteSession removes a session
//
//	@Summary		Delete session
//	@Description	Remove a session for a specific provider
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string					true	"Provider name"
//	@Success		200			{object}	map[string]any	"Session deleted successfully"
//	@Failure		400			{object}	map[string]any	"Bad request"
//	@Failure		500			{object}	map[string]any	"Internal server error"
//	@Router			/session/{provider} [delete]
func (s *Server) deleteSession(c *gin.Context) {

	provider := c.Param("provider")
	if len(provider) == 0 {
		s.getErrorPage(c, http.StatusBadRequest, "Provider is required")
		return
	}

	sessionManager := sessions.GetSessionManager()
	err := sessionManager.RemoveSession(s.Config.GetLoginServerHostname(), provider)

	if err != nil {
		s.getErrorPage(c, http.StatusInternalServerError, "Failed to delete session", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session deleted successfully",
	})
}
