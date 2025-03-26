package handler

import (
	"net/http"

	"github.com/JennerWork/chatbot/internal/middleware"
	"github.com/JennerWork/chatbot/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterRequest registration request parameters
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,min=2,max=50"`
}

// UpdatePasswordRequest update password request parameters
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateProfileRequest update profile request parameters
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=50"`
}

// CustomerHandler customer-related handler
type CustomerHandler struct {
	customerService service.CustomerService
}

// NewCustomerHandler create a customer handler
func NewCustomerHandler(customerService service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// Register handle customer registration
// @Summary Customer Registration
// @Description Register a new customer account
// @Tags customers
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration Information"
// @Success 200 {object} model.Customer
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/customers/register [post]
func (h *CustomerHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	customer, err := h.customerService.Register(req.Email, req.Password, req.Name)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Registration failed"

		if err == service.ErrEmailExists {
			status = http.StatusBadRequest
			message = "Email already registered"
		}

		c.JSON(status, ErrorResponse{
			Code:    status,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, customer)
}

// UpdatePassword handle password update
// @Summary Update Password
// @Description Update the password of the currently logged-in user
// @Tags customers
// @Accept json
// @Produce json
// @Param request body UpdatePasswordRequest true "Password Update Information"
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/customers/password [put]
func (h *CustomerHandler) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	customerID := middleware.GetCustomerID(c)
	if customerID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "Unauthorized user",
		})
		return
	}

	if err := h.customerService.UpdatePassword(customerID, req.OldPassword, req.NewPassword); err != nil {
		status := http.StatusInternalServerError
		message := "Failed to update password"

		if err == service.ErrInvalidCredentials {
			status = http.StatusBadRequest
			message = "Incorrect original password"
		}

		c.JSON(status, ErrorResponse{
			Code:    status,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
	})
}

// UpdateProfile handle profile update
// @Summary Update Profile
// @Description Update the profile of the currently logged-in user
// @Tags customers
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile Update Information"
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/customers/profile [put]
func (h *CustomerHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	customerID := middleware.GetCustomerID(c)
	if customerID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "Unauthorized user",
		})
		return
	}

	if err := h.customerService.UpdateProfile(customerID, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Message: "Failed to update profile",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}
