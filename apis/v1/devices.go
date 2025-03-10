package v1

import (
	"fcm/common/response"
	"fcm/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DevicesHandler struct {
	DeviceService services.IDevices
}

func NewDevices(engine *gin.Engine, deviceService services.IDevices) {
	handler := &DevicesHandler{
		DeviceService: deviceService,
	}

	group := engine.Group("v1/devices")
	{
		group.GET("device-code", handler.DeviceCode)
		group.GET("poll-for-token", handler.PollForToken)
	}
}

func (handler *DevicesHandler) DeviceCode(c *gin.Context) {
	clientId := c.Query("client_id")
	if len(clientId) < 1 {
		c.JSON(response.BadRequestMsg("client_id is empty"))
		return
	}

	scope := c.Query("scope")
	if len(scope) < 1 {
		c.JSON(response.BadRequestMsg("scope is empty"))
		return
	}

	deviceEndpoint := c.Query("device_endpoint")
	if len(deviceEndpoint) < 1 {
		c.JSON(response.BadRequestMsg("device_endpoint is empty"))
		return
	}

	result, err := handler.DeviceService.GetDeviceCode(c, clientId, scope, deviceEndpoint)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	c.JSON(response.OK(result))
}

func (handler *DevicesHandler) PollForToken(c *gin.Context) {
	clientId := c.Query("client_id")
	if len(clientId) < 1 {
		c.JSON(response.BadRequestMsg("client_id is empty"))
		return
	}

	deviceCode := c.Query("device_code")
	if len(deviceCode) < 1 {
		c.JSON(response.BadRequestMsg("device_code is empty"))
		return
	}

	tokenEndpoint := c.Query("token_endpoint")
	if len(tokenEndpoint) < 1 {
		c.JSON(response.BadRequestMsg("token_endpoint is empty"))
		return
	}

	intervalStr := c.Query("interval")
	if len(intervalStr) < 1 {
		c.JSON(response.BadRequestMsg("interval is empty"))
		return
	}
	interval, _ := strconv.Atoi(intervalStr)

	expiresInStr := c.Query("expires_in")
	if len(expiresInStr) < 1 {
		c.JSON(response.BadRequestMsg("expires_in is empty"))
		return
	}
	expiresIn, _ := strconv.Atoi(expiresInStr)

	result, err := handler.DeviceService.GetPollForToken(c, clientId, deviceCode, tokenEndpoint, interval, expiresIn)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	c.JSON(response.OK(result))
}
