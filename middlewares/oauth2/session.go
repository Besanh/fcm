package oauth2

import (
	"fcm/models"

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
)

func SessionMiddleware(sessionManager *scs.SessionManager, data models.MiddlwareOauth2) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
