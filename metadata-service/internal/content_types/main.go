package content_types

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

var ScriptContentTypes = []string{"text/x-shellscript"}
var JsonContentTypes = []string{"application/json", "text/plain", "*/*"}
var YamlContentTypes = []string{"application/yaml", "text/yaml"}

func ValidateContentType(c *gin.Context, requested_content_type string, allowed_content_types [][]string) bool {
	var allowed []string
	for _, content_types := range allowed_content_types {
		allowed = append(allowed, content_types...)
	}

	if slices.Contains(allowed, requested_content_type) {
		return true
	}

	c.JSON(http.StatusNotAcceptable, gin.H{
		"error": "Unsupported content type",
		"message": "The requested content type is not supported. Please use one of the following: " + strings.Join(allowed, ", "),
		"requested_content_type": requested_content_type,
		"allowed_content_types": allowed,
	})

	return false
}

func IsJsonContentType(requested_content_type string) bool {
	return slices.Contains(JsonContentTypes, requested_content_type)
}
func IsYamlContentType(requested_content_type string) bool {
	return slices.Contains(YamlContentTypes, requested_content_type)
}
func IsScriptContentType(requested_content_type string) bool {
	return slices.Contains(ScriptContentTypes, requested_content_type)
}

