package ginmw

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/KKolyasik/url-shortify/internal/config"
	"github.com/KKolyasik/url-shortify/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidCookie = errors.New("cookie is invalid")
	ErrNoVisitorID   = errors.New("no visitor ID")
)

type VisitorClaims struct {
	jwt.RegisteredClaims
	VisitorID uuid.UUID `json:"vid"`
}

func GinAuthorization(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("Authorization")
		if err != nil {
			logger.Log.Sugar().Infow("Кука пришла с ошибкой. Создаем новую.")
			claims := VisitorClaims{
				VisitorID: uuid.New(),
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.TokenTTL)),
				},
			}
			token, err := generateToken(cfg, claims)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			ctx.Request = ctx.Request.WithContext(
				context.WithValue(ctx.Request.Context(), "vid", claims.VisitorID),
			)
			ctx.SetCookie("Authorization", token, int(cfg.TokenTTL.Seconds()), "/", "", false, true)
			ctx.Next()
			return
		}

		vid, err := getVisitorID(cfg, token)
		if err != nil {
			if errors.Is(err, ErrNoVisitorID) {
				logger.Log.Sugar().Infow("Отсутствует VisitorID")
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			claims := VisitorClaims{
				VisitorID: uuid.New(),
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.TokenTTL)),
				},
			}
			token, err := generateToken(cfg, claims)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			ctx.Request = ctx.Request.WithContext(
				context.WithValue(ctx.Request.Context(), "vid", claims.VisitorID),
			)
			ctx.SetCookie("Authorization", token, int(cfg.TokenTTL.Seconds()), "/", "", false, true)
			ctx.Next()
			return
		}

		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), "vid", vid))
		ctx.Next()
	}
}

func getVisitorID(cfg *config.Config, tokenString string) (uuid.UUID, error) {
	var claims VisitorClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) { return []byte(cfg.SecretKey), nil })

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, ErrInvalidCookie
	}

	if claims.VisitorID == uuid.Nil {
		return uuid.Nil, ErrNoVisitorID
	}

	return claims.VisitorID, err
}

func generateToken(cfg *config.Config, claims VisitorClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
