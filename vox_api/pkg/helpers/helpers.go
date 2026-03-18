package helpers

import (
	"github.com/gin-gonic/gin"
)

func IsValString(ctx *gin.Context, key string) (str string, ok bool) {
	val, _ok := ctx.Get(key)
	if !_ok {
		return str, ok
	}

	switch v := val.(type) {
	case string:
		str = v
		ok = true
	default:
		return str, ok
	}

	return str, ok
}
