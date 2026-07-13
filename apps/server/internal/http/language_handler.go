package http

import (
	"net/http"

	"github.com/acmhot100/server/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// listLanguages handles GET /api/v1/languages.
func listLanguages(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		languages, err := repository.ListEnabledLanguages(db)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list languages")
			return
		}

		items := make([]languageDTO, len(languages))
		for i, language := range languages {
			items[i] = newLanguageDTO(language)
		}
		c.JSON(http.StatusOK, items)
	}
}
