package model

import (
	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/web"
)

type Collection struct {
	ID               int64
	Name             string
	SourceLanguage   string
	TargetLanguage   string
	DefaultWord      string
	HTMLTemplateName web.TemplateName
	Template         apkg.Model
}

func CollectionSlug(source, target string) string {
	return source + "-" + target
}
