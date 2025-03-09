package v1

import (
	"fcm/common/response"
	"fcm/services"

	"github.com/gin-gonic/gin"
)

type DevicesHandler struct {
	deviceService services.IDevices
}

func NewDevices(engine *gin.Engine, deviceService services.IDevices) {
	handler := &DevicesHandler{
		deviceService: deviceService,
	}

	group := engine.Group("v1/devices")
	{
		group.GET("devicecode", handler.DeviceCode)
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

	result, err := handler.deviceService.GetDeviceCode(c, clientId, scope, deviceEndpoint)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	c.JSON(200, result)
}
