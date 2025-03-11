package v1

import (
	"fcm/common/response"
	"fcm/middlewares/oauth2"
	"fcm/models"
	"fcm/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DevicesNotificationHandler struct {
	DeviceNotificationService services.IDevicesNotification
}

// NewDevicesNotification sets up the routes for handling device notifications.
// It initializes a DevicesNotificationHandler with the provided service and registers
// the following endpoints:
// - GET /v1/devices_notification/get-notifications: Fetches device notifications.
// - GET /v1/devices_notification/get-total-unread-notifications: Retrieves the total count of unread notifications.
// - PUT /v1/devices_notification/put-bulk-notifications: Performs bulk actions on notifications.
// All routes are protected with OAuth2 middleware.
func NewDevicesNotification(engine *gin.Engine, deviceNotificationService services.IDevicesNotification) {
	// Initialize the handler with the provided device notification service
	handler := &DevicesNotificationHandler{
		DeviceNotificationService: deviceNotificationService,
	}

	// Define a new route group for device notifications with OAuth2 protection
	group := engine.Group("v1/devices_notification").Use(oauth2.NewOAuth2Middleware())
	{
		// Register endpoints with their respective handlers
		group.GET("get-notifications", handler.GetDeviceNotifications)
		group.GET("get-total-unread-notifications", handler.GetTotalUnreadNotifications)
		group.PUT("put-bulk-notifications", handler.PutBulkDeviceNotification)
	}
}

// GetDeviceNotifications fetches device notifications based on the request query params.
// It queries the service to get the total count of notifications and the actual entries.
// The result is wrapped in a Pagination struct and returned as JSON.
func (handler *DevicesNotificationHandler) GetDeviceNotifications(c *gin.Context) {
	user, err := oauth2.GetUser(c)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	// Get the limit and offset from the query params
	limitStr := c.Query("limit")
	if len(limitStr) < 1 {
		c.JSON(response.BadRequestMsg("limit is empty"))
		return
	}
	limit, _ := strconv.Atoi(limitStr)

	offsetStr := c.Query("offset")
	if len(offsetStr) < 1 {
		c.JSON(response.BadRequestMsg("offset is empty"))
		return
	}
	offset, _ := strconv.Atoi(offsetStr)

	// Query the service for the total count and the actual entries
	total, result, err := handler.DeviceNotificationService.GetDeviceNotifications(c, user, models.DevicesNotificationQueryParam{
		// TODO: add filters
	}, int64(limit), int64(offset))
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	// Wrap the result in a Pagination struct and return it as JSON
	c.JSON(response.Pagination(total, result, limit, offset))
}

// GetTotalUnreadNotifications retrieves the total count of unread notifications for the current user.
// It queries the service to get the total count of unread notifications and returns it as JSON.
func (handler *DevicesNotificationHandler) GetTotalUnreadNotifications(c *gin.Context) {
	user, err := oauth2.GetUser(c)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	// Query the service for the total count of unread notifications
	result, err := handler.DeviceNotificationService.GetTotalUnreadNotifications(c, user, models.DevicesNotificationQueryParam{
		// TODO: add filters
	})
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	// Return the result as JSON
	c.JSON(response.OK(models.DevicesNotificationStatsResponse{
		TotalUnread: result,
	}))
}

// PutBulkDeviceNotification handles the PUT request to /v1/devices/notifications to
// perform a bulk action on device notifications for the current user.
// It queries the service to get the total count of notifications and the actual
// entries, and then performs the action on the notifications.
func (handler *DevicesNotificationHandler) PutBulkDeviceNotification(c *gin.Context) {
	user, err := oauth2.GetUser(c)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	body := &models.DevicesNotificationBulkRequest{}
	if err := c.ShouldBindJSON(body); err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	if err := body.Validate(); err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	// Query the service for the total count and the actual entries
	err = handler.DeviceNotificationService.PutBulkDeviceNotification(c, user, models.DevicesNotificationQueryParam{
		// TODO: add filters
	}, body)
	if err != nil {
		c.JSON(response.BadRequestMsg(err.Error()))
		return
	}

	// Return a success response
	c.JSON(response.OKResponse())
}
