package oauth2

import (
	"fcm/common/cache"
	"fcm/models"
	"fcm/services"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewOAuth2Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		headerValue := c.GetHeader("Authorization")
		token := parseTokenFromAuthorization(headerValue)
		user, err := validateToken(token)
		if err != nil {
			c.Set("UNAUTHORIZED", true)
			c.Abort()
			return
		}
		c.Set("USER", user)
		c.Next()
	}
}

func parseTokenFromAuthorization(authorizationHeader string) string {
	return strings.Replace(authorizationHeader, "Bearer ", "", 1)
}

func validateToken(tokenString string) (userInfo *models.User, err error) {
	// Because the token is stored in redis, so I need to get the user info from it
	dataCache := cache.RCache.Get(fmt.Sprintf("%s:%s", services.OAUTH2_TOKEN, tokenString))
	if dataCache == nil {
		err = fmt.Errorf("invalid token")
		return
	}

	userInfo = dataCache.(*models.User)

	return
}
