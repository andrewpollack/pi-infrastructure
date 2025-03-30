package meal_backend

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const EXPIRES_IN = 730 * time.Hour // 1 month
var signingMethod = jwt.SigningMethodHS256

func createToken(signingKey []byte) (string, error) {
	// TODO: Replace with actual username
	const USERNAME = "admin"
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(EXPIRES_IN)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "meals-go",
		Subject:   USERNAME,
		Audience:  getRole(USERNAME),
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	ss, err := token.SignedString(signingKey)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return ss, err
}

func getRole(username string) []string {
	return []string{"user"}
}

func verifyToken(tokenString string, signingKey []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})

	// Check for verification errors
	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func (c Config) authenticateMiddleware(ctx *gin.Context) {
	tokenString, err := ctx.Cookie("token")
	if err != nil {
		fmt.Println("Token missing in cookie")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing in cookie"})
		ctx.Abort()
		return
	}

	token, err := verifyToken(tokenString, c.JWTSigningKey)
	if err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed"})
		ctx.Abort()
		return
	}

	fmt.Printf("Token verified successfully. Claims: %+v\\n", token.Claims)

	ctx.Next()
}
