package v1

import (
	"fcm/common/response"
	"fcm/models"
	"fcm/services"
	"fmt"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService    services.IUsers
	SessionManager *scs.SessionManager
}

func NewUsers(engine *gin.Engine, sessionManager *scs.SessionManager, userService services.IUsers) {
	handler := &UserHandler{
		UserService:    userService,
		SessionManager: sessionManager,
	}

	engine.GET("login", handler.Login)
	engine.GET("oauth2callback", handler.OAuth2Callback)

	group := engine.Group("v1/user")
	{
		group.GET("me", handler.Me)
	}
}

func (handler *UserHandler) Login(c *gin.Context) {
	url := handler.UserService.Login(c)
	fmt.Println(url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (handler *UserHandler) OAuth2Callback(c *gin.Context) {
	code := c.Query("code")
	if len(code) < 1 {
		c.JSON(response.BadRequestMsg("code is empty"))
		return
	}

	state := c.Query("state")
	if len(state) < 1 {
		c.JSON(response.BadRequestMsg("state is empty"))
		return
	}

	scope := c.Query("scope")
	if len(scope) < 1 {
		c.JSON(response.BadRequestMsg("scope is empty"))
		return
	}

	callbackData := &models.OAuth2Callback{
		Code:  code,
		State: state,
		Scope: scope,
	}

	token, err := handler.UserService.OAuth2Callback(c, callbackData)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}
	c.JSON(response.Created(map[string]any{
		"token": token,
	}))
}

func (handler *UserHandler) Me(c *gin.Context) {

}
