package token_service

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"os"
	"strconv"
	"time"
)

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   uuid.UUID
	RefreshUuid  uuid.UUID
	AtExpires    int64
	RtExpires    int64
}

type AccessTokenClaims struct {
	AccessUUID string `json:"accessUuid"`
	UserID     int    `json:"userId"`
	Exp        int    `json:"exp"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	RefreshUUID string `json:"refreshUuid"`
	UserID      int    `json:"userId"`
	Exp         int    `json:"exp"`
	jwt.StandardClaims
}

type AccesTokenCache struct {
	UserID      int    `json:"userId"`
	RefreshUUID string `json:"refreshUuid"`
}

type Service struct {
	cache         *redis.Client
	accessSecret  string
	refreshSecret string
}

func NewService(cache *redis.Client, accessSecret, refreshSecret string) *Service {
	return &Service{
		cache:         cache,
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

func (s *Service) CreateToken(userID int) (*TokenDetails, error) {
	td := &TokenDetails{}

	accessExpMinutes, err := strconv.Atoi(os.Getenv("ACCESS_EXP_MINUTES"))
	if err != nil {
		return nil, fmt.Errorf("error getting ACCESS_EXP_MINUTES: %w", err)
	}

	refreshExpMinutes, err := strconv.Atoi(os.Getenv("REFRESH_EXP_MINUTES"))
	if err != nil {
		return nil, fmt.Errorf("error getting REFRESH_EXP_MINUTES: %w", err)
	}

	td.AtExpires = time.Now().Add(time.Minute * time.Duration(accessExpMinutes)).Unix()
	td.AccessUuid = uuid.New()

	td.RtExpires = time.Now().Add(time.Minute * time.Duration(refreshExpMinutes)).Unix()
	td.RefreshUuid = uuid.New()

	atClaims := jwt.MapClaims{}
	atClaims["accessUuid"] = td.AccessUuid.String()
	atClaims["userId"] = userID
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(s.accessSecret))
	if err != nil {
		return nil, fmt.Errorf("error signing access token: %w", err)
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["refreshUuid"] = td.RefreshUuid.String()
	rtClaims["userId"] = userID
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(s.refreshSecret))
	if err != nil {
		return nil, fmt.Errorf("error signing refresh token: %w", err)
	}

	return td, nil
}

func (s *Service) DecodeAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.accessSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while decoding access token: %w", err)
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token is not valid")
}

func (s *Service) DecodeRefreshToken(tokenStr string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.refreshSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while decoding refresh token: %w", err)
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token is not valid")
}

func (s *Service) CreateCacheKey(ctx context.Context, userID int, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) // convert to UTC
	rt := time.Unix(td.RtExpires, 0) // convert to UTC
	now := time.Now()

	cacheJSON, err := json.Marshal(AccesTokenCache{
		UserID:      userID,
		RefreshUUID: td.RefreshUuid.String(),
	})
	if err != nil {
		return fmt.Errorf("error while marshalling access token cache: %w", err)
	}

	if err = s.cache.Set(ctx, td.AccessUuid.String(), cacheJSON, at.Sub(now)).Err(); err != nil {
		return fmt.Errorf("error while setting access token to cache: %w", err)
	}

	if err = s.cache.Set(ctx, td.RefreshUuid.String(), strconv.Itoa(userID), rt.Sub(now)).Err(); err != nil {
		return fmt.Errorf("error while setting refresh token to cache: %w", err)
	}

	return nil
}

func (s *Service) DropCacheKey(ctx context.Context, Uuid string) error {
	err := s.cache.Del(ctx, Uuid).Err()
	if err != nil {
		return fmt.Errorf("error while deleting key from cache: %w", err)
	}

	return nil
}

func (s *Service) GetCacheValue(ctx context.Context, Uuid string) (*string, error) {
	value, err := s.cache.Get(ctx, Uuid).Result()
	if err != nil {
		return nil, fmt.Errorf("error while getting value from cache: %w", err)
	}

	return &value, nil
}

func (s *Service) DropCacheTokens(ctx context.Context, accessTokenClaims AccessTokenClaims) error {
	cacheJSON, err := s.GetCacheValue(ctx, accessTokenClaims.AccessUUID)
	if err != nil {
		return fmt.Errorf("error while getting cache value: %w", err)
	}

	accessTokenCache := new(AccesTokenCache)
	err = json.Unmarshal([]byte(*cacheJSON), accessTokenCache)
	if err != nil {
		return fmt.Errorf("error while unmarshalling access token cache: %w", err)
	}

	// drop refresh token from Redis cache
	err = s.DropCacheKey(ctx, accessTokenCache.RefreshUUID)
	if err != nil {
		return fmt.Errorf("error while dropping refresh token: %w", err)
	}

	// drop access token from Redis cache
	err = s.DropCacheKey(ctx, accessTokenClaims.AccessUUID)
	if err != nil {
		return fmt.Errorf("error while dropping access token: %w", err)
	}

	return nil
}
