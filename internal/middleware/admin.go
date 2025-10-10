package middleware

// func AdminRequired(db *gorm.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		userID, _ := c.Get("user_id")
// 		var existingUser models.User
// 		result := db.Where("user_id = ?", userID).First(&existingUser)
// 		if result.Error != nil {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
// 			c.Abort()
// 			return
// 		}
// 		if !existingUser.IsAdmin {
// 			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
// 			c.Abort()
// 			return
// 		}
// 		c.Next()
// 	}
// }
