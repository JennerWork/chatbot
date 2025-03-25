package dao

import (
	"errors"
	"time"

	"github.com/JennerWork/chatbot/internal/model"

	"gorm.io/gorm"
)

// MessageQuery 消息查询的参数
type MessageQuery struct {
	SessionID uint  `json:"session_id"`
	LastSeq   *uint `json:"last_seq,omitempty"` // 上次同步的最后一条消息序号
	Limit     int   `json:"limit,omitempty"`    // 每次获取的消息数量
}

// MessageDAO 消息数据访问对象
type MessageDAO struct {
	db *gorm.DB
}

// NewMessageDAO 创建消息DAO实例
func NewMessageDAO(db *gorm.DB) *MessageDAO {
	return &MessageDAO{db: db}
}

// GenerateNextSeq 生成下一个消息序号
func (dao *MessageDAO) GenerateNextSeq(message *model.Message) error {
	if message.SessionID == 0 {
		return errors.New("session id is required")
	}

	var maxSeq struct {
		MaxSeq uint
	}

	// 查询当前会话中最大的序号
	result := dao.db.Model(&model.Message{}).
		Select("COALESCE(MAX(seq), 0) as max_seq").
		Where("session_id = ?", message.SessionID).
		Scan(&maxSeq)

	if result.Error != nil {
		return result.Error
	}

	// 设置新消息的序号为最大序号+1
	message.Seq = maxSeq.MaxSeq + 1
	return nil
}

// ValidateSeq 验证消息序号的连续性
func (dao *MessageDAO) ValidateSeq(sessionID uint) error {
	var count int64
	result := dao.db.Model(&model.Message{}).
		Where("session_id = ?", sessionID).
		Where("seq != (SELECT COUNT(*) FROM (SELECT DISTINCT seq FROM messages m2 WHERE m2.session_id = ? AND m2.seq <= messages.seq) subq)", sessionID).
		Count(&count)

	if result.Error != nil {
		return result.Error
	}

	if count > 0 {
		return errors.New("message sequence is not continuous")
	}

	return nil
}

// GetMessages 获取消息列表
func (dao *MessageDAO) GetMessages(query MessageQuery) ([]model.Message, error) {
	if query.Limit <= 0 {
		query.Limit = 20 // 默认每次返回20条消息
	}

	db := dao.db.Model(&model.Message{}).Where("session_id = ?", query.SessionID)

	// 如果提供了LastSeq，则获取该序号之后的消息
	if query.LastSeq != nil {
		db = db.Where("seq > ?", *query.LastSeq)
	}

	var messages []model.Message
	result := db.Order("seq asc").Limit(query.Limit).Find(&messages)
	return messages, result.Error
}

// GetLatestMessages 获取最新的消息
func (dao *MessageDAO) GetLatestMessages(sessionID uint, limit int) ([]model.Message, error) {
	if limit <= 0 {
		limit = 20
	}

	var messages []model.Message
	result := dao.db.Model(&model.Message{}).
		Where("session_id = ?", sessionID).
		Order("seq desc").
		Limit(limit).
		Find(&messages)

	// 反转消息顺序，使其按seq升序排列
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, result.Error
}

// GetMessagesBySeqRange 获取指定序号范围内的消息
func (dao *MessageDAO) GetMessagesBySeqRange(sessionID uint, startSeq, endSeq uint) ([]model.Message, error) {
	var messages []model.Message
	result := dao.db.Model(&model.Message{}).
		Where("session_id = ? AND seq >= ? AND seq <= ?", sessionID, startSeq, endSeq).
		Order("seq asc").
		Find(&messages)

	return messages, result.Error
}

// GetLastMessageSeq 获取会话最新消息序号
func (dao *MessageDAO) GetLastMessageSeq(sessionID uint) (uint, error) {
	var message model.Message
	result := dao.db.Model(&model.Message{}).
		Where("session_id = ?", sessionID).
		Order("seq desc").
		First(&message)

	if result.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	return message.Seq, result.Error
}

// CreateMessage 创建新消息
func (dao *MessageDAO) CreateMessage(message *model.Message) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		// 生成消息序号
		if err := dao.GenerateNextSeq(message); err != nil {
			return err
		}

		// 保存消息
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		// 验证序号连续性
		return dao.ValidateSeq(message.SessionID)
	})
}

// GetMessagesByCustomer 获取客户的所有消息
func (dao *MessageDAO) GetMessagesByCustomer(customerID uint, limit int) ([]model.Message, error) {
	if limit <= 0 {
		limit = 20
	}

	var messages []model.Message
	result := dao.db.Model(&model.Message{}).
		Where("customer_id = ?", customerID).
		Order("created_at desc, session_id, seq asc").
		Limit(limit).
		Find(&messages)

	return messages, result.Error
}

// GetCustomerMessagesByTimeRange 获取指定时间范围内客户的消息
func (dao *MessageDAO) GetCustomerMessagesByTimeRange(customerID uint, startTime, endTime time.Time) ([]model.Message, error) {
	var messages []model.Message
	result := dao.db.Model(&model.Message{}).
		Where("customer_id = ? AND created_at BETWEEN ? AND ?", customerID, startTime, endTime).
		Order("created_at asc, session_id, seq asc").
		Find(&messages)

	return messages, result.Error
}
