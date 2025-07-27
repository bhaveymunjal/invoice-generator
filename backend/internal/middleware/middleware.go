package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"invoice-generator/internal/auth"
	"invoice-generator/internal/constants"
	"invoice-generator/internal/models"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader(constants.AuthorizationHeader)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if len(tokenString) > constants.BearerPrefixLength && tokenString[:constants.BearerPrefixLength] == constants.BearerPrefix {
			tokenString = tokenString[constants.BearerPrefixLength:]
		}

		claims, err := auth.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}

// AdminMiddleware ensures the user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ValidateInvoiceData validates invoice data before creation
func ValidateInvoiceData() gin.HandlerFunc {
	return func(c *gin.Context) {
		var invoice models.Invoice
		if err := c.ShouldBindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice data: " + err.Error()})
			c.Abort()
			return
		}

		// Validate invoice type
		isValidType := false
		for _, t := range constants.ValidInvoiceTypes {
			if invoice.InvoiceType == t {
				isValidType = true
				break
			}
		}
		if !isValidType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice type"})
			c.Abort()
			return
		}

		// Validate line items
		if len(invoice.LineItems) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice must have at least one line item"})
			c.Abort()
			return
		}

		c.Set("invoice", invoice)
		c.Next()
	}
}
