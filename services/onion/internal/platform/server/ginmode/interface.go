package ginmode

import (
	"context"

	"github.com/gin-gonic/gin"
)

type GinRunMode interface {
	Setup(ctx context.Context, r *gin.Engine)
}
