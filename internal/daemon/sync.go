package daemon

import (
	"time"

	"github.com/gin-gonic/gin"
)

// TODO: create. a true sync endpoint

// getSync retrieves sync status
//
//	@Summary		Sync status
//	@Description	Get the current sync status and version information
//	@Tags			sync
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]any	"Sync status"
//	@Router			/sync [get]
//	@Security		BearerAuth
func (s *Server) getSync(c *gin.Context) {
	c.JSON(200, gin.H{
		"version":   s.GetVersion(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
