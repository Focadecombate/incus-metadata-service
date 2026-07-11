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

// AcceptsYAML reports whether the Accept header permits a YAML response.
// An empty Accept, a "*/*" wildcard (including inside a compound header such
// as "application/json, */*"), or an explicit YAML media type all qualify.
// cloud-init's NoCloud datasource sends "Accept: */*", so it must pass here.
func AcceptsYAML(accept string) bool {
	if accept == "" {
		return true
	}
	for _, part := range strings.Split(accept, ",") {
		media := strings.TrimSpace(part)
		if i := strings.IndexByte(media, ';'); i >= 0 {
			media = strings.TrimSpace(media[:i])
		}
		if media == "*/*" || slices.Contains(YamlContentTypes, media) {
			return true
		}
	}
	return false
}

// WantsJSON reports whether the client explicitly requested JSON. A wildcard
// ("*/*") or missing Accept is deliberately NOT treated as JSON so that the
// default YAML representation is served for cloud-init clients.
func WantsJSON(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		media := strings.TrimSpace(part)
		if i := strings.IndexByte(media, ';'); i >= 0 {
			media = strings.TrimSpace(media[:i])
		}
		if media == "application/json" {
			return true
		}
	}
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

