package apidocs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Inject registers Swagger UI and OpenAPI YAML routes on the provided Gin router.
func Inject(router gin.IRoutes, options Options) {
	if router == nil {
		return
	}
	options = normalizeOptions(options)
	if options.Disabled {
		return
	}
	router.GET(options.Path, uiHandler(options))
	router.GET(options.SpecPath, specHandler(options))
}

func uiHandler(options Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, err := renderHTML(options)
		if err != nil {
			c.String(http.StatusInternalServerError, "render api docs")
			return
		}
		c.Data(http.StatusOK, htmlContentType, page)
	}
}

func specHandler(options Options) gin.HandlerFunc {
	spec := append([]byte(nil), options.Spec...)
	return func(c *gin.Context) {
		if len(spec) == 0 {
			c.String(http.StatusInternalServerError, "openapi document is empty")
			return
		}
		c.Data(http.StatusOK, openAPIYAMLContentType, spec)
	}
}
