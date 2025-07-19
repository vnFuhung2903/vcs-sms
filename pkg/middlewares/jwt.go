package middlewares

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vnFuhung2903/vcs-sms/interfaces"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
)

type IJWTMiddleware interface {
	GenerateJWT(context context.Context, userId string, username string, scope []string) error
	RequireScope(requiredScope string) gin.HandlerFunc
}

type jwtMiddleware struct {
	redisClient interfaces.IRedisClient
	jwtSecret   []byte
}

func NewJWTMiddleware(redisClient interfaces.IRedisClient, env env.AuthEnv) IJWTMiddleware {
	return &jwtMiddleware{
		redisClient: redisClient,
		jwtSecret:   []byte(env.JWTSecret),
	}
}

const jwtExpiry = time.Hour * 24 * 7

func (m *jwtMiddleware) GenerateJWT(ctx context.Context, userId string, username string, scope []string) error {
	claims := jwt.MapClaims{
		"sub":   userId,
		"name":  username,
		"scope": scope,
		"exp":   time.Now().Add(jwtExpiry).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString(m.jwtSecret)
	if err != nil {
		return err
	}

	if err := m.redisClient.Set(ctx, "token", jwtToken, time.Hour*24*7); err != nil {
		return err
	}
	return nil
}

func (m *jwtMiddleware) RequireScope(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		jwtToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})
		if err != nil || !jwtToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		rawScopes, ok := claims["scope"].([]interface{})
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid scope format"})
			c.Abort()
			return
		}

		tokens := make([]string, 0, len(rawScopes))
		for _, s := range rawScopes {
			if str, ok := s.(string); ok {
				tokens = append(tokens, str)
			}
		}

		if found := slices.Contains(tokens, requiredScope); !found {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient scope"})
			c.Abort()
			return
		}

		if sub, ok := claims["sub"].(string); ok {
			c.Set("userId", sub)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Insufficient userId"})
			return
		}
		c.Next()
	}
}
