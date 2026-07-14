package http

import (
	"net/http"

	"github.com/acmhot100/server/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getProfileSummary(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		summary, err := repository.GetProfileProgressSummary(db, getUserID(c))
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load profile summary")
			return
		}
		c.JSON(http.StatusOK, summary)
	}
}

func getProfileProgressByStage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stages, err := repository.GetProfileProgressByStage(db, getUserID(c))
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load stage progress")
			return
		}
		if stages == nil {
			stages = []repository.StageProgress{}
		}
		c.JSON(http.StatusOK, stages)
	}
}
