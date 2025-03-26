package handler

import (
	"net/http"

	"github.com/JennerWork/chatbot/internal/middleware"
	"github.com/JennerWork/chatbot/internal/service"
	"github.com/gin-gonic/gin"
)

// MessageHandler message-related handler
type MessageHandler struct {
	queryService service.MessageQueryService
}

// NewMessageHandler create message handler
func NewMessageHandler(queryService service.MessageQueryService) *MessageHandler {
	return &MessageHandler{
		queryService: queryService,
	}
}

// GetMessageHistory get message history
// @Summary Get Chat History
// @Description Get chat history records of the currently authenticated user
// @Tags messages
// @Accept json
// @Produce json
// @Param session_id query uint false "Session ID"
// @Param start_time query string false "Start Time (format: 2006-01-02 15:04:05)"
// @Param end_time query string false "End Time (format: 2006-01-02 15:04:05)"
// @Param page query int false "Page Number (default: 1)"
// @Param page_size query int false "Page Size (default: 20)"
// @Success 200 {object} service.MessageQueryResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/message/list [get]
func (h *MessageHandler) GetMessageHistory(c *gin.Context) {
	var params service.MessageQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	// 从认证中间件获取customer_id
	customerID := middleware.GetCustomerID(c)
	if customerID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "Unauthorized user",
		})
		return
	}

	// 设置customer_id和默认值
	params.CustomerID = customerID
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100 // Limit maximum page size
	}

	result, err := h.queryService.GetMessageHistory(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Message: "Failed to get message history",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
