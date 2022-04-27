package slug

import (
	"github.com/gosimple/unidecode"
	"regexp"
	"strings"
)

var regexpNonAuthorizedChars = regexp.MustCompile("[^a-zA-Z0-9-_]")
var regexpMultipleDashes = regexp.MustCompile("-+")

func Slugify(s string) (slug string) {
	slug = strings.TrimSpace(s)
	slug = unidecode.Unidecode(slug)
	slug = strings.ToLower(slug)
	slug = regexpNonAuthorizedChars.ReplaceAllString(slug, "-")
	slug = regexpMultipleDashes.ReplaceAllString(slug, "-")
	return slug
}
