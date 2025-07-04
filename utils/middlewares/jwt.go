package middlewares

import (
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vnFuhung2903/vcs-sms/entities"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

const jwtExpiry = time.Hour * 24 * 7 // 1 week

func GenerateJWT(userID string, username string, scope []string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"name":  username,
		"scope": scope,
		"exp":   time.Now().Add(jwtExpiry).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func RequireScope(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		jwtToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
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

func UserRoleToDefaultScopes(role entities.UserRole) []string {
	switch role {
	case entities.Developer:
		{
			return []string{"user:modify", "container:create", "container:view", "container:update", "container:delete"}
		}
	case entities.Manager:
		{
			return []string{"user:modify", "user:manager", "container:view"}
		}
	default:
		{
			return []string{"user:modify", "container:view"}
		}
	}
}

var scopeHashMap = []string{"user:modify", "user:manager", "container:create", "container:view", "container:update", "container:delete"}

func ScopeHashMap(userScopes []string) int64 {
	res := int64(0)
	for i, scope := range scopeHashMap {
		if found := slices.Contains(userScopes, scope); found {
			res |= (1 << i)
		}
	}
	return res
}
