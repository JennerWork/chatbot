package service

import (
	"errors"

	"github.com/JennerWork/chatbot/internal/model"
	"gorm.io/gorm"
)

var (
	ErrEmailExists = errors.New("邮箱已存在")
)

// CustomerService 客户服务接口
type CustomerService interface {
	// Register 注册新客户
	Register(email, password, name string) (*model.Customer, error)
	// UpdatePassword 更新密码
	UpdatePassword(customerID uint, oldPassword, newPassword string) error
	// GetByID 根据ID获取客户信息
	GetByID(id uint) (*model.Customer, error)
	// UpdateProfile 更新客户资料
	UpdateProfile(customerID uint, name string) error
}

type customerService struct {
	db *gorm.DB
}

// NewCustomerService 创建客户服务实例
func NewCustomerService(db *gorm.DB) CustomerService {
	return &customerService{
		db: db,
	}
}

// Register 实现客户注册
func (s *customerService) Register(email, password, name string) (*model.Customer, error) {
	// 检查邮箱是否已存在
	var count int64
	if err := s.db.Model(&model.Customer{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrEmailExists
	}

	// 创建新客户
	customer := &model.Customer{
		Email: email,
		Name:  name,
	}

	// 设置密码
	if err := customer.SetPassword(password); err != nil {
		return nil, err
	}

	// 保存到数据库
	if err := s.db.Create(customer).Error; err != nil {
		return nil, err
	}

	return customer, nil
}

// UpdatePassword 实现密码更新
func (s *customerService) UpdatePassword(customerID uint, oldPassword, newPassword string) error {
	var customer model.Customer
	if err := s.db.First(&customer, customerID).Error; err != nil {
		return err
	}

	// 验证旧密码
	if !customer.ValidatePassword(oldPassword) {
		return ErrInvalidCredentials
	}

	// 设置新密码
	if err := customer.SetPassword(newPassword); err != nil {
		return err
	}

	return s.db.Save(&customer).Error
}

// GetByID 实现获取客户信息
func (s *customerService) GetByID(id uint) (*model.Customer, error) {
	var customer model.Customer
	if err := s.db.First(&customer, id).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

// UpdateProfile 实现更新客户资料
func (s *customerService) UpdateProfile(customerID uint, name string) error {
	return s.db.Model(&model.Customer{}).Where("id = ?", customerID).Update("name", name).Error
}
