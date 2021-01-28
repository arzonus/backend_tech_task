package external

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

func GetEngine() *gin.Engine {
	engine := gin.Default()
	
	engine.GET("/status/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		bsonId, err := primitive.ObjectIDFromHex(id)
		// For emulate time precessing
		time.Sleep(time.Second * 2)

		if err != nil || bsonId.IsZero() {
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		if bsonId.Timestamp().Second() % 2 == 0 {
			ctx.JSON(http.StatusOK, gin.H{
				"status": "Processed",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status": "Skipped",
		})
	})

	return engine
}