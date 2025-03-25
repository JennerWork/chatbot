package handler

import (
	"net/http"

	"github.com/JennerWork/chatbot/internal/middleware"
	"github.com/JennerWork/chatbot/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,min=2,max=50"`
}

// UpdatePasswordRequest 更新密码请求参数
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateProfileRequest 更新资料请求参数
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=50"`
}

// CustomerHandler 客户相关的处理器
type CustomerHandler struct {
	customerService service.CustomerService
}

// NewCustomerHandler 创建客户处理器
func NewCustomerHandler(customerService service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// Register 处理客户注册
// @Summary 客户注册
// @Description 注册新客户账号
// @Tags customers
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 200 {object} model.Customer
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/customers/register [post]
func (h *CustomerHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "无效的请求参数",
			Error:   err.Error(),
		})
		return
	}

	customer, err := h.customerService.Register(req.Email, req.Password, req.Name)
	if err != nil {
		status := http.StatusInternalServerError
		message := "注册失败"

		if err == service.ErrEmailExists {
			status = http.StatusBadRequest
			message = "邮箱已被注册"
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

// UpdatePassword 处理密码更新
// @Summary 更新密码
// @Description 更新当前登录用户的密码
// @Tags customers
// @Accept json
// @Produce json
// @Param request body UpdatePasswordRequest true "密码更新信息"
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
			Message: "无效的请求参数",
			Error:   err.Error(),
		})
		return
	}

	customerID := middleware.GetCustomerID(c)
	if customerID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "未认证的用户",
		})
		return
	}

	if err := h.customerService.UpdatePassword(customerID, req.OldPassword, req.NewPassword); err != nil {
		status := http.StatusInternalServerError
		message := "更新密码失败"

		if err == service.ErrInvalidCredentials {
			status = http.StatusBadRequest
			message = "原密码错误"
		}

		c.JSON(status, ErrorResponse{
			Code:    status,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密码更新成功",
	})
}

// UpdateProfile 处理资料更新
// @Summary 更新资料
// @Description 更新当前登录用户的资料
// @Tags customers
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "资料更新信息"
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
			Message: "无效的请求参数",
			Error:   err.Error(),
		})
		return
	}

	customerID := middleware.GetCustomerID(c)
	if customerID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "未认证的用户",
		})
		return
	}

	if err := h.customerService.UpdateProfile(customerID, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Message: "更新资料失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "资料更新成功",
	})
}
