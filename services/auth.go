package services

import (
	"errors"
	"fmt"
	"time"

	"go-gin-demo/config"
	"go-gin-demo/database"
	"go-gin-demo/dto"
	"go-gin-demo/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var cfg *config.Config

// InitAuth 初始化认证服务，设置全局配置。
func InitAuth(c *config.Config) {
	cfg = c
}

// Register 创建新用户，初始化他们的 30 个农场块。
func Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// 检查用户名是否重复
	var existing models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errors.New("username already taken")
	}

	// 密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	user := models.User{
		Username: req.Username,
		Password: string(hash),
		Gold:     200,
		Level:    1,
		ExpVal:   0,
		ExpToLvl: 100,
	}

	// 在事务中创建用户和农场块
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		plots := make([]models.FarmPlot, 30)
		for i := 0; i < 30; i++ {
			plots[i] = models.FarmPlot{
				UserID:    user.ID,
				PlotIndex: i,
				Unlocked:  i == 0, // 第一个地块解锁
			}
		}
		return tx.Create(&plots).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	token, err := generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

// Login 验证凭证并返回一个 JWT 令牌。
func Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	token, err := generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

func generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func toUserResponse(u models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Gold:     u.Gold,
		Level:    u.Level,
		ExpVal:   u.ExpVal,
		ExpToLvl: u.ExpToLvl,
	}
}

// GetUserByID 根据 ID 返回用户信息。
func GetUserByID(userID uint) (*dto.UserResponse, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	resp := toUserResponse(user)
	return &resp, nil
}
