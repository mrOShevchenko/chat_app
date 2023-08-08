package handlers

import (
	"chat_app/internal/models"
	"chat_app/pkg/validators"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
)

type DeviceDetails struct {
	Type  string `json:"type" example:"web"`
	Token string `json:"token" example:"token"`
}

// AddDevice to user's devices.

// AddDevice godoc
// @Summary Add device
// @Description Set device Token(Firebase Cloud Messaging) for push notifications
// @Tags device
// @Accept  json
// @Produce application/json
// @Param password body DeviceDetails true "raw request body"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 409 {object} Response
// @Failure 500 {object} Response
// @Router /v1/devices [post]
func (h *BaseHandler) AddDevice(ctx echo.Context) error {
	user := ctx.Get("user").(*models.User)
	userAgent := ctx.Request().UserAgent()

	var deviceDetails DeviceDetails

	if err := ctx.Bind(&deviceDetails); err != nil {
		return ErrorResponse(ctx, http.StatusBadRequest, "invalid device data", err)
	}

	if err := validators.ValidateDeviceType(deviceDetails.Type); err != nil {
		return ErrorResponse(ctx, http.StatusBadRequest, "invalid device type", err)
	}

	for _, device := range user.Devices {
		if device.Token == deviceDetails.Token {
			return ErrorResponse(ctx, http.StatusConflict, "device with this token already exists", errors.New("device with this token already exists"))
		}
	}

	user.Devices = append(user.Devices, &models.Device{
		Type:  deviceDetails.Type,
		Token: deviceDetails.Token,
		Name:  userAgent,
	})

	if err := h.userRepo.UpdateWithAssociations(user); err != nil {
		return ErrorResponse(ctx, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(ctx, http.StatusOK, "successfully set device", user)
}
