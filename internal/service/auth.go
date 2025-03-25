package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/JennerWork/chatbot/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("无效的凭证")
	ErrTokenExpired       = errors.New("token已过期")
	ErrInvalidToken       = errors.New("无效的token")
)

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey     string        // JWT密钥
	TokenExpiry   time.Duration // Token过期时间
	RefreshExpiry time.Duration // 刷新Token过期时间
}

// AuthService 认证服务接口
type AuthService interface {
	// Login 登录并返回token
	Login(email, password string) (string, error)
	// ValidateToken 验证token
	ValidateToken(tokenString string) (*Claims, error)
	// RefreshToken 刷新token
	RefreshToken(tokenString string) (string, error)
}

// Claims 定义JWT的payload
type Claims struct {
	jwt.RegisteredClaims
	CustomerID uint   `json:"customer_id"`
	Email      string `json:"email"`
}

type authService struct {
	db     *gorm.DB
	config JWTConfig
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB, config JWTConfig) AuthService {
	return &authService{
		db:     db,
		config: config,
	}
}

// Login 登录实现
func (s *authService) Login(email, password string) (string, error) {
	var customer model.Customer
	if err := s.db.Where("email = ?", email).First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// TODO: 实现密码验证
	// if !customer.ValidatePassword(password) {
	//     return "", ErrInvalidCredentials
	// }

	// 生成token
	return s.generateToken(customer)
}

// ValidateToken 验证token
func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken 刷新token
func (s *authService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil && !errors.Is(err, ErrTokenExpired) {
		return "", err
	}

	// 检查用户是否存在
	var customer model.Customer
	if err := s.db.First(&customer, claims.CustomerID).Error; err != nil {
		return "", ErrInvalidToken
	}

	// 生成新token
	return s.generateToken(customer)
}

// generateToken 生成JWT token
func (s *authService) generateToken(customer model.Customer) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		CustomerID: customer.ID,
		Email:      customer.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}
