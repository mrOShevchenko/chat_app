package handlers

import (
	"chat_app/internal/models"
	"chat_app/pkg/validators"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

type UpdateUsernameRequestBody struct {
	Username string `json:"username" example:"username"`
}

type UpdatePasswordRequestBody struct {
	NewPassword string `json:"newPassword" example:"NewP@ssw0rd"`
	OldPassword string `json:"oldPassword" example:"OldP@ssw0rd"`
}

// User godoc
// @Summary Get user data
// @Description Get user data
// @Tags user
// @Accept  json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user [get]
func (h *BaseHandler) User(c echo.Context) error {
	user := c.Get("user").(*models.User)

	return SuccessResponse(c, http.StatusOK, "successfully found user data", user)
}

// UploadImage godoc
// @Summary Upload user image
// @Description Uploading user image as file by form-data "image"
// @Tags user
// @Param image formData file true "User image file. The preferred size is 315x315px because the image will resize to 315x315px. Max size: 2MB, Allowed types: 'jpg', 'jpeg', 'png', 'gif'"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/upload-image [post]
func (h *BaseHandler) UploadImage(c echo.Context) error {
	user := c.Get("user").(*models.User)
	src, err := validateAndGetFile(c, "imageAvatar")
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "error: can't get file image avatar", err)
	}
	defer closeFile(src, "imageAvatar")

	inputImage, err := decodeAndResizeImage(src)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error: can't resize image avatar", err)
	}

	dst, err := prepareDestination(user)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error: can't prepare destination", err)
	}
	defer closeFile(dst, "dst")

	if err = saveImage(dst, src, inputImage); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error with saving image", err)
	}

	return finalizeImageUpload(c, h, user)
}

func validateAndGetFile(c echo.Context, key string) (multipart.File, error) {
	file, err := c.FormFile(key)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("%s is required", key))
	}
	if file.Size > 2*1024*1024 {
		return nil, errors.Wrapf(err, fmt.Sprintf("%s size is to big", key))
	}
	src, err := file.Open()
	return src, nil
}

func closeFile(file multipart.File, name string) {
	err := file.Close()
	if err != nil {
		log.Printf("error closing %s: %s", name, err)
	}
}

func decodeAndResizeImage(src multipart.File) (image.Image, error) {
	inputImage, _, err := image.Decode(src)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding image")
	}
	width := inputImage.Bounds().Size().X
	height := inputImage.Bounds().Size().Y
	if width >= height {
		return imaging.Resize(inputImage, 0, 315, imaging.Lanczos), nil
	}
	return imaging.Resize(inputImage, 315, 0, imaging.Lanczos), nil
}

func prepareDestination(user *models.User) (*os.File, error) {
	path := os.Getenv("IMAGES_DIRECTORY_PATH")
	if path == "" {
		return nil, errors.New("images directory path is not set")
	}
	imageName := uuid.New().String() + ".png"
	user.Image = os.Getenv("URL_PREFIX_IMAGES") + imageName
	return os.Create(path + imageName)
}

func saveImage(dst *os.File, src multipart.File, img image.Image) error {
	err := imaging.Encode(dst, img, imaging.PNG)
	if err != nil {
		return errors.New("error encoding image")
	}
	_, err = io.Copy(dst, src)
	return err
}

func finalizeImageUpload(c echo.Context, h *BaseHandler, user *models.User) error {
	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error saving image", err)
	}
	oldImageName := user.Image
	if oldImageName != "" {
		path := os.Getenv("IMAGES_DIRECTORY_PATH")
		err := os.Remove(path + oldImageName)
		if err != nil {
			c.Logger().Error(err)
		}
	}
	return SuccessResponse(c, http.StatusOK, "Image uploaded successfully", nil)
}

// UpdateUsername godoc
// @Summary update username
// @Description Username should contain:
// @Description - lower, upper case latin letters and digits
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Tags user
// @Accept  json
// @Produce application/json
// @Param username body UpdateUsernameRequestBody true "raw request body"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/update-username [put]
func (h *BaseHandler) UpdateUsername(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	var requestPayload UpdateUsernameRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := validators.ValidateUsername(requestPayload.Username); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "username is incorrect", err)
	}

	if tmpUser, err := h.userRepo.FindByUsername(requestPayload.Username); tmpUser != nil {
		return ErrorResponse(c, http.StatusBadRequest, "username is already taken", err)
	}

	user.Username = requestPayload.Username

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error with updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("username successfuly updated to %s", user.Username), nil)
}

// UpdatePassword godoc
// @Summary update password
// @Description OldPassword/NewPassword should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Tags user
// @Accept  json
// @Produce application/json
// @Param password body UpdatePasswordRequestBody true "raw request body"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/update-password [put]
func (h *BaseHandler) UpdatePassword(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	var requestPayload UpdatePasswordRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := validators.ValidatePassword(requestPayload.OldPassword); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "old password is incorrect", err)
	}

	if err := validators.ValidatePassword(requestPayload.NewPassword); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "new password incorrect", err)
	}

	ok, err := h.userRepo.PasswordMatches(user, requestPayload.OldPassword)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error matching passwords", err)
	}
	if !ok {
		err := errors.New("old password is incorrect")
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}

	if err := h.userRepo.ResetPassword(user, requestPayload.NewPassword); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating password", err)
	}

	return SuccessResponse(c, http.StatusOK, "password successfully updated", nil)
}

// Block godoc
// @Summary Block user
// @Tags user
// @Accept  json
// @Produce application/json
// @Param	user_id	path	int	true	"Blocked User ID"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/{user_id}/block [post]
func (h *BaseHandler) Block(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("user_id")
	blockedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "incorrect userId", err)
	}

	blockedUser, err := h.userRepo.FindByID(blockedUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid blockedUserId", err)
	}
	if blockedUser == nil {
		err := errors.New(fmt.Sprintf("user with id %d not found", blockedUserID))
		return ErrorResponse(c, http.StatusBadRequest, "blockedUser not found ", err)
	}
	if blockedUser.ID == user.ID {
		err := errors.New(fmt.Sprintf("user with id %d can't block himself", blockedUserID))
		return ErrorResponse(c, http.StatusBadRequest, "you can't block yourself", err)
	}
	for _, item := range user.BlockedUsers {
		if item.ID == blockedUser.ID {
			err := errors.New(fmt.Sprintf("user with id %d is already blocked", blockedUserID))
			return ErrorResponse(c, http.StatusBadRequest, "user is already blocked", err)
		}
	}

	user.BlockedUsers = append(user.BlockedUsers, blockedUser)

	if err := h.userRepo.UpdateWithAssociations(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("successfully blacklisted user %s (id: %d)", blockedUser.Username, blockedUser.ID), user)
}

// Unblock godoc
// @Summary Unblock user
// @Tags user
// @Accept  json
// @Produce application/json
// @Param	user_id	path	int	true	"Unblocked User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/{user_id}/block [delete]
func (h *BaseHandler) Unblock(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("user_id")
	blockedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid blocked userId", err)
	}

	for i, item := range user.BlockedUsers {
		if item.ID == blockedUserID {
			followedUsers := append(user.BlockedUsers[:i], user.BlockedUsers[i+1:]...)
			err := h.userRepo.ReplaceBlockedUsers(user, followedUsers)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to unblock user", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully unblocked user", user)
		}
	}
	err = errors.New(fmt.Sprintf("user with id %d is not blocked", blockedUserID))
	return ErrorResponse(c, http.StatusNotFound, "user is not blocked", err)
}

// Follow godoc
// @Summary Follow user
// @Tags user
// @Accept  json
// @Produce application/json
// @Param	user_id	path	int	true	"Followed User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/{user_id}/follow [post]
func (h *BaseHandler) Follow(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("user_id")
	followedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followedUserId", err)
	}

	followedUser, err := h.userRepo.FindByID(followedUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if followedUser == nil {
		err := errors.New(fmt.Sprintf("user with id %d not found", followedUserID))
		return ErrorResponse(c, http.StatusBadRequest, "can't found followed userId", err)
	}
	if followedUser.ID == user.ID {
		err := errors.New(fmt.Sprintf("user with id %d can't follow himself", followedUserID))
		return ErrorResponse(c, http.StatusBadRequest, "you can't follow yourself", err)
	}
	for _, item := range user.FollowedUsers {
		if item.ID == followedUser.ID {
			err := errors.New(fmt.Sprintf("user with id %d is already followed", followedUserID))
			return ErrorResponse(c, http.StatusBadRequest, "user is already followed", err)
		}
	}

	user.FollowedUsers = append(user.FollowedUsers, followedUser)

	if err := h.userRepo.UpdateWithAssociations(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("successfully followed user %s (id: %d)", followedUser.Username, followedUser.ID), user)
}

// Unfollow godoc
// @Summary Unfollow user
// @Tags user
// @Accept  json
// @Produce application/json
// @Param	user_id	path	int	true	"Followed User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/{user_id}/follow [delete]
func (h *BaseHandler) Unfollow(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("user_id")
	followedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid unfollowed userId", err)
	}

	for i, item := range user.FollowedUsers {
		if item.ID == followedUserID {
			followedUsers := append(user.FollowedUsers[:i], user.FollowedUsers[i+1:]...)
			err = h.userRepo.ReplaceFollowedUsers(user, followedUsers)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to unfollow user", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully unfollowed user", user)
		}
	}
	err = errors.New(fmt.Sprintf("user with id %d is not followed", followedUserID))
	return ErrorResponse(c, http.StatusNotFound, "user is not followed", err)
}
