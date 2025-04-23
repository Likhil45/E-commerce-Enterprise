package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var JWTSecret = []byte("one_piece")

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			ctx.Abort()
			return
		}
		log.Println("Authorization Header:", authHeader)

		// Ensure proper format and extract the token safely
		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
			log.Println("Invalid Authorization header format and length of parts:", len(parts))
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			ctx.Abort()
			return
		}
		tokenString := parts[1]
		log.Println("Token String:", tokenString)

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC for security
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Println("Invalid signing method")
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token signing method"})
				return nil, jwt.ErrSignatureInvalid
			}
			return JWTSecret, nil
		})

		if err != nil || token == nil || !token.Valid {
			log.Println("Token Validity:", token.Valid)
			log.Printf("Token parsing failed: %v", err)

			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "Invalid token", "error": err.Error()})
			ctx.Abort()
			return
		}

		// Extract claims safely
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			log.Printf("JWT Claims: %v", claims)
		}
		// if !ok {
		// 	ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		// 	ctx.Abort()
		// 	return
		// }

		// Extract user ID and store it in context
		userID, exists := claims["user_id"]
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User ID missing in token"})
			ctx.Abort()
			return
		}

		ctx.Set("userID", userID)
		log.Printf("userID set in context: %v", userID)
		log.Println(ctx.Get("userID"))
		ctx.Next()
	}
}
