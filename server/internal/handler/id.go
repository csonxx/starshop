package handler

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PrimitiveFromParam(c *gin.Context) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(c.Param("id"))
}
