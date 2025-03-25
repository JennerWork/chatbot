package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Customer 客户信息
type Customer struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"` // 密码哈希，json中不返回
	Salt      string         `gorm:"size:32;not null" json:"-"`  // 密码盐值
	Name      string         `gorm:"size:100" json:"name"`
	Status    string         `gorm:"size:20;default:active" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// SetPassword 设置密码
func (c *Customer) SetPassword(password string) error {
	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	c.Password = string(hashedPassword)
	return nil
}

// ValidatePassword 验证密码
func (c *Customer) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(password))
	return err == nil
}

// BeforeCreate GORM的钩子，在创建记录前执行
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.Status == "" {
		c.Status = "active"
	}
	return nil
}
