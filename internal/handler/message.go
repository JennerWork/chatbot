package handler

import (
	"net/http"

	"github.com/JennerWork/chatbot/internal/middleware"
	"github.com/JennerWork/chatbot/internal/service"
	"github.com/gin-gonic/gin"
)

// MessageHandler 消息相关的处理器
type MessageHandler struct {
	queryService service.MessageQueryService
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(queryService service.MessageQueryService) *MessageHandler {
	return &MessageHandler{
		queryService: queryService,
	}
}

// GetMessageHistory 获取消息历史
// @Summary 获取聊天历史
// @Description 获取当前认证用户的聊天历史记录
// @Tags messages
// @Accept json
// @Produce json
// @Param session_id query uint false "会话ID"
// @Param start_time query string false "开始时间 (格式: 2006-01-02 15:04:05)"
// @Param end_time query string false "结束时间 (格式: 2006-01-02 15:04:05)"
// @Param page query int false "页码 (默认: 1)"
// @Param page_size query int false "每页数量 (默认: 20)"
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
			Message: "无效的请求参数",
			Error:   err.Error(),
		})
		return
	}

	// 从认证中间件获取customer_id
	customerID := middleware.GetCustomerID(c)
	if customerID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    401,
			Message: "未认证的用户",
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
		params.PageSize = 100 // 限制最大页面大小
	}

	result, err := h.queryService.GetMessageHistory(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Message: "获取消息历史失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
