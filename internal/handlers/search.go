package handlers

import (
	"chat_app/internal/models"
	"chat_app/pkg/validators"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

const (
	defaultSearchType  = "user"
	defaultOrder       = "asc"
	defaultSearchLimit = 10
	maxSearchLimit     = 1000
)

// Search godoc
// @Summary search users
// @Description Search users by username with autocomplete
// @Tags search
// @Accept  json
// @Produce application/json
// @Param q searchQuery string true "searchQuery string for search by username, minimum 1 character, maximum 40 characters"
// @Param type searchQuery string false "type of search, default: 'user', available: 'user', 'friend', 'blacklist'"
// @Param order searchQuery string false "order of search, default: 'asc', available: 'asc', 'desc'"
// @Param searchLimit searchQuery int false "limit of search, default: '10', available: '1-1000'"
// @Security BearerAuth
// @Success 200 {object} Response{data=[]models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/search [get]
func (h *BaseHandler) Search(ctx echo.Context) error {
	user := ctx.Get("user").(*models.User)
	searchQuery := ctx.QueryParam("q")
	if err := validators.ValidateQuery(searchQuery); err != nil {
		return ErrorResponse(ctx, http.StatusBadRequest, "invalid searchQuery. "+err.Error(), err)
	}

	searchType := ctx.QueryParam("type")
	if searchType == "" {
		searchType = defaultSearchType
	}

	if err := validateSearchType(searchType); err != nil {
		return ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err)
	}

	order := ctx.QueryParam("order")
	if order == "" {
		order = defaultOrder
	}

	if err := validateOrder(order); err != nil {
		return ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err)
	}

	searchLimit := defaultSearchLimit
	limitStr := ctx.QueryParam("limit")
	if limitStr != "" {
		var err error
		searchLimit, err = strconv.Atoi(limitStr)
		if err != nil {
			return ErrorResponse(ctx, http.StatusBadRequest, "Invalid limit value. Limit must be an integer", err)
		}
	}

	if err := validateSearchLimit(searchLimit); err != nil {
		return ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err)
	}

	users, err := h.userRepo.FindArrayByPartUsername(searchQuery, order, searchLimit)
	if err != nil {
		return ErrorResponse(ctx, http.StatusInternalServerError, "Error occurred while finding users", err)
	}

	users = filterUsers(users, user.ID)

	if users == nil || len(*users) == 0 {
		err := fmt.Errorf("no users found with username %s", searchQuery)
		return ErrorResponse(ctx, http.StatusNotFound, "users not found", err)
	}

	return SuccessResponse(ctx, http.StatusOK, "users found successfully", users)
}

// filterUsers function exclude the current user from the search results
func filterUsers(users *[]models.User, userID int) *[]models.User {
	for i, item := range *users {
		if item.ID == userID {
			*users = append((*users)[:i], (*users)[i+1:]...)
			break
		}
	}
	return users
}

func validateSearchType(searchType string) error {
	if searchType != "user" && searchType != "friend" && searchType != "blacklist" {
		return errors.New("Invalid search type. Allowed types: 'user', 'friend', 'blacklist'")
	}
	return nil
}

func validateOrder(order string) error {
	if order != "asc" && order != "desc" {
		return errors.New("Invalid order parameter. Allowed orders: 'asc', 'desc'")
	}
	return nil
}

func validateSearchLimit(limit int) error {
	if limit < 1 || limit > maxSearchLimit {
		return errors.New("Invalid limit value. Allowed limit values: 1-1000")
	}
	return nil
}
