package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Shortifyer interface {
	Shortify(http.ResponseWriter, *http.Request)
}

type Redirector interface {
	Redirect(http.ResponseWriter, *http.Request, string)
}

type JSONShortifyer interface {
	ShortifyJSON(http.ResponseWriter, *http.Request)
}

type BatchShortifyer interface {
	ShortenBatch(w http.ResponseWriter, r *http.Request)
}

type UserResolver interface {
	UserURLS(w http.ResponseWriter, r *http.Request)
}

func GinShortify(s Shortifyer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.Shortify(ctx.Writer, ctx.Request)
	}
}

func GinRedirect(r Redirector) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		r.Redirect(ctx.Writer, ctx.Request, id)
	}
}

func GinShortifyJSON(s JSONShortifyer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.ShortifyJSON(ctx.Writer, ctx.Request)
	}
}

func GinShortifyBatch(s BatchShortifyer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.ShortenBatch(ctx.Writer, ctx.Request)
	}
}

func GinUserURLS(u UserResolver) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		u.UserURLS(ctx.Writer, ctx.Request)
	}
}
