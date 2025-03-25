package service

import (
	"time"

	"github.com/JennerWork/chatbot/internal/model"
	"gorm.io/gorm"
)

// MessageQueryParams 消息查询参数
type MessageQueryParams struct {
	SessionID  uint      `form:"session_id"`
	StartTime  time.Time `form:"start_time" time_format:"2006-01-02 15:04:05"`
	EndTime    time.Time `form:"end_time" time_format:"2006-01-02 15:04:05"`
	Page       int       `form:"page"`
	PageSize   int       `form:"page_size"`
	CustomerID uint      `form:"-"` // 从认证中间件获取，不从请求参数中获取
}

// MessageQueryResult 消息查询结果
type MessageQueryResult struct {
	Total    int64           `json:"total"`    // 总记录数
	Messages []MessageDetail `json:"messages"` // 消息列表
}

// MessageDetail 消息详情
type MessageDetail struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	Sender    string `json:"sender"`     // customer 或 bot
	Seq       uint   `json:"seq"`        // 消息序号
	CreatedAt string `json:"created_at"` // 格式化的时间字符串
	SessionID uint   `json:"session_id"`
}

// MessageQueryService 消息查询服务
type MessageQueryService interface {
	// GetMessageHistory 获取消息历史
	GetMessageHistory(params MessageQueryParams) (*MessageQueryResult, error)
}

type messageQueryService struct {
	db *gorm.DB
}

// NewMessageQueryService 创建消息查询服务实例
func NewMessageQueryService(db *gorm.DB) MessageQueryService {
	return &messageQueryService{
		db: db,
	}
}

// GetMessageHistory 获取消息历史实现
func (s *messageQueryService) GetMessageHistory(params MessageQueryParams) (*MessageQueryResult, error) {
	// 构建查询
	query := s.db.Model(&model.Message{}).Where("customer_id = ?", params.CustomerID)

	// 添加可选的查询条件
	if params.SessionID > 0 {
		query = query.Where("session_id = ?", params.SessionID)
	}
	if !params.StartTime.IsZero() {
		query = query.Where("created_at >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("created_at <= ?", params.EndTime)
	}

	// 获取总记录数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 计算分页
	offset := (params.Page - 1) * params.PageSize

	// 获取消息列表
	var messages []model.Message
	if err := query.Order("session_id asc, seq asc").
		Offset(offset).
		Limit(params.PageSize).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	// 转换为 MessageDetail
	details := make([]MessageDetail, len(messages))
	for i, msg := range messages {
		details[i] = MessageDetail{
			ID:        msg.ID,
			Content:   msg.Content,
			Sender:    string(msg.Sender),
			Seq:       msg.Seq,
			CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
			SessionID: msg.SessionID,
		}
	}

	return &MessageQueryResult{
		Total:    total,
		Messages: details,
	}, nil
}
