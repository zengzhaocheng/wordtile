package service

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/wordtile/backend/internal/apperr"
	"github.com/wordtile/backend/internal/auth"
	"github.com/wordtile/backend/internal/config"
	"github.com/wordtile/backend/internal/model"
	"github.com/wordtile/backend/internal/repository"
)

// AuthService 注册 / 登录 / 刷新令牌。
type AuthService struct {
	repo *repository.DB
	cfg  *config.Config
}

func NewAuthService(repo *repository.DB, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) issueTokens(u *model.User) (access, refresh string, err error) {
	access, err = auth.GenerateToken(u.ID, auth.TokenTypeAccess, s.cfg.JWTSecret, s.cfg.AccessTTL)
	if err != nil {
		return "", "", err
	}
	refresh, err = auth.GenerateToken(u.ID, auth.TokenTypeRefresh, s.cfg.JWTSecret, s.cfg.RefreshTTL)
	return access, refresh, err
}

func (s *AuthService) Register(username, email, password, displayName string) (*model.User, string, string, error) {
	if len(username) < 2 || len(username) > 64 {
		return nil, "", "", apperr.New(400, "用户名长度须为 2~64 个字符")
	}
	if len(password) < 6 {
		return nil, "", "", apperr.New(400, "密码至少 6 位")
	}
	if displayName == "" {
		displayName = username
	}
	if _, err := s.repo.GetUserByUsername(username); err == nil {
		return nil, "", "", apperr.New(409, "用户名已被占用")
	}
	if _, err := s.repo.GetUserByEmail(email); err == nil {
		return nil, "", "", apperr.New(409, "邮箱已被注册")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}
	u := &model.User{
		ID: uuid.New(), Username: username, Email: email,
		PasswordHash: string(hash), DisplayName: displayName, ThemePreference: "system",
	}
	if err := s.repo.CreateUser(u); err != nil {
		return nil, "", "", err
	}
	access, refresh, err := s.issueTokens(u)
	return u, access, refresh, err
}

func (s *AuthService) Login(identifier, password string) (*model.User, string, string, error) {
	u, err := s.repo.GetUserByUsername(identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u, err = s.repo.GetUserByEmail(identifier)
		}
		if err != nil {
			return nil, "", "", apperr.New(401, "用户名或密码错误")
		}
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, "", "", apperr.New(401, "用户名或密码错误")
	}
	access, refresh, err := s.issueTokens(u)
	return u, access, refresh, err
}

func (s *AuthService) Refresh(refreshToken string) (string, string, error) {
	claims, err := auth.ParseToken(refreshToken, s.cfg.JWTSecret)
	if err != nil || claims.Type != auth.TokenTypeRefresh {
		return "", "", apperr.New(401, "refresh token 无效或已过期")
	}
	u, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return "", "", apperr.ErrUnauthorized
	}
	return s.issueTokens(u)
}

// GetUser 按 ID 查询用户。
func (s *AuthService) GetUser(userID uuid.UUID) (*model.User, error) {
	u, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return u, nil
}

// UpdateProfile 更新资料 / 主题偏好。
func (s *AuthService) UpdateProfile(userID uuid.UUID, displayName, avatarURL, themePreference *string) (*model.User, error) {
	u, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if displayName != nil {
		u.DisplayName = *displayName
	}
	if avatarURL != nil {
		u.AvatarURL = *avatarURL
	}
	if themePreference != nil {
		switch *themePreference {
		case "light", "dark", "system":
			u.ThemePreference = *themePreference
		default:
			return nil, apperr.New(400, "theme_preference 仅支持 light/dark/system")
		}
	}
	if err := s.repo.UpdateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}
