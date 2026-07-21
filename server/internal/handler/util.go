package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"star/server/pkg/response"
)

func objectIDParam(c *gin.Context) (primitive.ObjectID, bool) {
	s := c.Param("id")
	id, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return primitive.NilObjectID, false
	}
	return id, true
}

func intQuery(c *gin.Context, key string, fallback int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return v
}

func cleanString(v string) string {
	return strings.TrimSpace(v)
}

func cleanStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = cleanString(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
