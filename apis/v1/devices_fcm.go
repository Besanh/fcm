package v1

import (
	"fcm/common/response"
	"fcm/middlewares/oauth2"
	"fcm/models"
	"fcm/services"

	"github.com/gin-gonic/gin"
)

type DevicesFcmHandler struct {
	DevicesFcm services.IDevicesFcm
}

func NewDevicesFcm(engine *gin.Engine, devicesFcm services.IDevicesFcm) {
	handler := &DevicesFcmHandler{
		DevicesFcm: devicesFcm,
	}

	group := engine.Group("v1/fcm").Use(oauth2.NewOAuth2Middleware())
	{
		group.POST("register-device-token", handler.RegisterDeviceToken)
		group.POST("unregister-device-token", handler.UnregisterDeviceToken)
	}
}

/*
 * Fcm register device token
 */
func (h *DevicesFcmHandler) RegisterDeviceToken(c *gin.Context) {
	body := &models.RegisterDeviceTokenRequest{}
	if err := c.ShouldBindJSON(body); err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	err := h.DevicesFcm.RegisterDeviceToken(c, *body)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	c.JSON(response.OKResponse())
}

/*
 * Fcm unregister device token
 */
func (h *DevicesFcmHandler) UnregisterDeviceToken(c *gin.Context) {
	body := &models.RegisterDeviceTokenRequest{}
	if err := c.ShouldBindJSON(body); err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	err := h.DevicesFcm.UnRegisterDeviceToken(c, *body)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	c.JSON(response.OKResponse())
}
