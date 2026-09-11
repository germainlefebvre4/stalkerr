package api

import "github.com/gin-gonic/gin"

// respondError writes a JSON ErrorResponse with the given status, code, and message.
func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{Error: code, Message: message})
}
