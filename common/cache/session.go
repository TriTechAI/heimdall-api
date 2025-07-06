package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
)

// SessionCache 会话缓存服务
type SessionCache struct {
	client redis.Cmdable
	prefix string
	logger logx.Logger
}

// SessionData 会话数据结构
type SessionData struct {
	UserID      string                 `json:"userId"`
	Username    string                 `json:"username"`
	Role        string                 `json:"role"`
	IPAddress   string                 `json:"ipAddress"`
	UserAgent   string                 `json:"userAgent"`
	LoginAt     time.Time              `json:"loginAt"`
	LastActive  time.Time              `json:"lastActive"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	TokenID     string                 `json:"tokenId,omitempty"`
}

// SessionInfo 会话信息（用于列表展示）
type SessionInfo struct {
	SessionID   string    `json:"sessionId"`
	UserID      string    `json:"userId"`
	Username    string    `json:"username"`
	IPAddress   string    `json:"ipAddress"`
	UserAgent   string    `json:"userAgent"`
	LoginAt     time.Time `json:"loginAt"`
	LastActive  time.Time `json:"lastActive"`
	ExpiresAt   time.Time `json:"expiresAt"`
	IsActive    bool      `json:"isActive"`
}

// NewSessionCache 创建会话缓存实例
func NewSessionCache(client redis.Cmdable, prefix string) *SessionCache {
	if client == nil {
		panic("redis client cannot be nil")
	}
	if prefix == "" {
		prefix = "session:"
	}
	
	return &SessionCache{
		client: client,
		prefix: prefix,
		logger: logx.WithContext(context.Background()),
	}
}

// CreateSession 创建新的用户会话
func (s *SessionCache) CreateSession(ctx context.Context, sessionID string, data *SessionData, expiration time.Duration) error {
	if err := s.validateSessionID(sessionID); err != nil {
		return err
	}
	if err := s.validateSessionData(data); err != nil {
		return err
	}
	
	// 设置创建时间和最后活动时间
	now := time.Now()
	data.LoginAt = now
	data.LastActive = now
	
	sessionKey := s.buildSessionKey(sessionID)
	userKey := s.buildUserSessionKey(data.UserID)
	
	// 序列化会话数据
	sessionJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	
	// 使用Pipeline确保原子性
	pipe := s.client.TxPipeline()
	
	// 存储会话数据
	pipe.Set(ctx, sessionKey, sessionJSON, expiration)
	
	// 在用户会话索引中添加此会话
	pipe.SAdd(ctx, userKey, sessionID)
	pipe.Expire(ctx, userKey, expiration+time.Hour) // 索引稍后过期，便于清理
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		s.logger.Errorf("Failed to create session %s: %v", sessionID, err)
		return fmt.Errorf("failed to create session: %w", err)
	}
	
	s.logger.Infof("Session %s created for user %s", sessionID, data.UserID)
	return nil
}

// GetSession 获取会话数据
func (s *SessionCache) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	if err := s.validateSessionID(sessionID); err != nil {
		return nil, err
	}
	
	sessionKey := s.buildSessionKey(sessionID)
	
	result := s.client.Get(ctx, sessionKey)
	if err := result.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // 会话不存在
		}
		s.logger.Errorf("Failed to get session %s: %v", sessionID, err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	var sessionData SessionData
	if err := json.Unmarshal([]byte(result.Val()), &sessionData); err != nil {
		s.logger.Errorf("Failed to unmarshal session data for %s: %v", sessionID, err)
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}
	
	return &sessionData, nil
}

// UpdateSession 更新会话数据
func (s *SessionCache) UpdateSession(ctx context.Context, sessionID string, data *SessionData) error {
	if err := s.validateSessionID(sessionID); err != nil {
		return err
	}
	if err := s.validateSessionData(data); err != nil {
		return err
	}
	
	sessionKey := s.buildSessionKey(sessionID)
	
	// 检查会话是否存在
	exists, err := s.client.Exists(ctx, sessionKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	if exists == 0 {
		return errors.New("session not found")
	}
	
	// 更新最后活动时间
	data.LastActive = time.Now()
	
	// 序列化更新后的数据
	sessionJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	
	// 获取剩余TTL
	ttl, err := s.client.TTL(ctx, sessionKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get session TTL: %w", err)
	}
	
	// 更新会话数据，保持原有的过期时间
	err = s.client.Set(ctx, sessionKey, sessionJSON, ttl).Err()
	if err != nil {
		s.logger.Errorf("Failed to update session %s: %v", sessionID, err)
		return fmt.Errorf("failed to update session: %w", err)
	}
	
	return nil
}

// TouchSession 更新会话最后活动时间
func (s *SessionCache) TouchSession(ctx context.Context, sessionID string) error {
	if err := s.validateSessionID(sessionID); err != nil {
		return err
	}
	
	sessionData, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if sessionData == nil {
		return errors.New("session not found")
	}
	
	// 只更新最后活动时间
	sessionData.LastActive = time.Now()
	
	return s.UpdateSession(ctx, sessionID, sessionData)
}

// DeleteSession 删除会话
func (s *SessionCache) DeleteSession(ctx context.Context, sessionID string) error {
	if err := s.validateSessionID(sessionID); err != nil {
		return err
	}
	
	// 先获取会话数据以获得用户ID
	sessionData, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if sessionData == nil {
		return nil // 会话已不存在
	}
	
	sessionKey := s.buildSessionKey(sessionID)
	userKey := s.buildUserSessionKey(sessionData.UserID)
	
	// 使用Pipeline确保原子性
	pipe := s.client.TxPipeline()
	
	// 删除会话数据
	pipe.Del(ctx, sessionKey)
	
	// 从用户会话索引中移除
	pipe.SRem(ctx, userKey, sessionID)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		s.logger.Errorf("Failed to delete session %s: %v", sessionID, err)
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	s.logger.Infof("Session %s deleted for user %s", sessionID, sessionData.UserID)
	return nil
}

// GetUserSessions 获取用户的所有活跃会话
func (s *SessionCache) GetUserSessions(ctx context.Context, userID string) ([]*SessionInfo, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}
	
	userKey := s.buildUserSessionKey(userID)
	
	// 获取用户的所有会话ID
	sessionIDs, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		s.logger.Errorf("Failed to get user sessions for %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	
	if len(sessionIDs) == 0 {
		return []*SessionInfo{}, nil
	}
	
	// 批量获取会话数据
	pipe := s.client.Pipeline()
	sessionCmds := make(map[string]*redis.StringCmd)
	ttlCmds := make(map[string]*redis.DurationCmd)
	
	for _, sessionID := range sessionIDs {
		sessionKey := s.buildSessionKey(sessionID)
		sessionCmds[sessionID] = pipe.Get(ctx, sessionKey)
		ttlCmds[sessionID] = pipe.TTL(ctx, sessionKey)
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		s.logger.Errorf("Failed to batch get sessions for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to batch get sessions: %w", err)
	}
	
	// 解析会话信息
	var sessions []*SessionInfo
	for _, sessionID := range sessionIDs {
		sessionInfo := s.parseSessionInfo(sessionID, sessionCmds[sessionID], ttlCmds[sessionID])
		if sessionInfo != nil {
			sessions = append(sessions, sessionInfo)
		}
	}
	
	return sessions, nil
}

// DeleteUserSessions 删除用户的所有会话
func (s *SessionCache) DeleteUserSessions(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("userID cannot be empty")
	}
	
	sessions, err := s.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}
	
	if len(sessions) == 0 {
		return nil
	}
	
	// 构建要删除的键名
	var keysToDelete []string
	userKey := s.buildUserSessionKey(userID)
	keysToDelete = append(keysToDelete, userKey)
	
	for _, session := range sessions {
		sessionKey := s.buildSessionKey(session.SessionID)
		keysToDelete = append(keysToDelete, sessionKey)
	}
	
	// 批量删除
	err = s.client.Del(ctx, keysToDelete...).Err()
	if err != nil {
		s.logger.Errorf("Failed to delete all sessions for user %s: %v", userID, err)
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	
	s.logger.Infof("Deleted %d sessions for user %s", len(sessions), userID)
	return nil
}

// ExtendSession 延长会话过期时间
func (s *SessionCache) ExtendSession(ctx context.Context, sessionID string, extension time.Duration) error {
	if err := s.validateSessionID(sessionID); err != nil {
		return err
	}
	
	sessionData, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if sessionData == nil {
		return errors.New("session not found")
	}
	
	sessionKey := s.buildSessionKey(sessionID)
	userKey := s.buildUserSessionKey(sessionData.UserID)
	
	// 使用Pipeline延长过期时间
	pipe := s.client.TxPipeline()
	pipe.Expire(ctx, sessionKey, extension)
	pipe.Expire(ctx, userKey, extension+time.Hour)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		s.logger.Errorf("Failed to extend session %s: %v", sessionID, err)
		return fmt.Errorf("failed to extend session: %w", err)
	}
	
	return nil
}

// CleanupExpiredSessions 清理过期的会话索引
func (s *SessionCache) CleanupExpiredSessions(ctx context.Context) error {
	pattern := s.prefix + "user:*"
	
	var cursor uint64
	var cleanedCount int64
	
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan user session keys: %w", err)
		}
		
		// 清理每个用户的过期会话
		for _, userKey := range keys {
			cleaned, err := s.cleanupUserSessions(ctx, userKey)
			if err != nil {
				s.logger.Errorf("Failed to cleanup sessions for key %s: %v", userKey, err)
				continue
			}
			cleanedCount += cleaned
		}
		
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	
	s.logger.Infof("Cleaned up %d expired sessions", cleanedCount)
	return nil
}

// parseSessionInfo 解析会话信息
func (s *SessionCache) parseSessionInfo(sessionID string, sessionCmd *redis.StringCmd, ttlCmd *redis.DurationCmd) *SessionInfo {
	if sessionCmd.Err() != nil {
		if !errors.Is(sessionCmd.Err(), redis.Nil) {
			s.logger.Errorf("Failed to get session %s: %v", sessionID, sessionCmd.Err())
		}
		return nil
	}
	
	var sessionData SessionData
	if err := json.Unmarshal([]byte(sessionCmd.Val()), &sessionData); err != nil {
		s.logger.Errorf("Failed to unmarshal session %s: %v", sessionID, err)
		return nil
	}
	
	ttl := ttlCmd.Val()
	expiresAt := time.Now().Add(ttl)
	isActive := ttl > 0
	
	return &SessionInfo{
		SessionID:  sessionID,
		UserID:     sessionData.UserID,
		Username:   sessionData.Username,
		IPAddress:  sessionData.IPAddress,
		UserAgent:  sessionData.UserAgent,
		LoginAt:    sessionData.LoginAt,
		LastActive: sessionData.LastActive,
		ExpiresAt:  expiresAt,
		IsActive:   isActive,
	}
}

// cleanupUserSessions 清理用户的过期会话
func (s *SessionCache) cleanupUserSessions(ctx context.Context, userKey string) (int64, error) {
	// 获取用户的所有会话ID
	sessionIDs, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return 0, err
	}
	
	var expiredSessions []string
	
	// 检查每个会话是否过期
	for _, sessionID := range sessionIDs {
		sessionKey := s.buildSessionKey(sessionID)
		exists, err := s.client.Exists(ctx, sessionKey).Result()
		if err != nil {
			continue
		}
		if exists == 0 {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}
	
	// 从用户会话索引中移除过期的会话
	if len(expiredSessions) > 0 {
		_, err := s.client.SRem(ctx, userKey, expiredSessions).Result()
		if err != nil {
			return 0, err
		}
	}
	
	return int64(len(expiredSessions)), nil
}

// validateSessionID 验证会话ID的有效性
func (s *SessionCache) validateSessionID(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID cannot be empty")
	}
	if len(sessionID) < 10 {
		return errors.New("session ID too short")
	}
	if len(sessionID) > 200 {
		return errors.New("session ID too long")
	}
	return nil
}

// validateSessionData 验证会话数据的有效性
func (s *SessionCache) validateSessionData(data *SessionData) error {
	if data == nil {
		return errors.New("session data cannot be nil")
	}
	if data.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	if data.Username == "" {
		return errors.New("username cannot be empty")
	}
	return nil
}

// buildSessionKey 构建会话键名
func (s *SessionCache) buildSessionKey(sessionID string) string {
	return s.prefix + "data:" + sessionID
}

// buildUserSessionKey 构建用户会话索引键名
func (s *SessionCache) buildUserSessionKey(userID string) string {
	return s.prefix + "user:" + userID
}