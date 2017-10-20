package controllers

import (
	"github.com/kataras/iris/context"
	q "github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/yoander/bolsillo/models"
)

// Tag define a TagMethods implementation
type Tag struct{}

// List list tags
func (tag *Tag) Read(ctx context.Context) {
}

// DumpAsJSON dumps tags
func (tag *Tag) DumpAsJSON(ctx context.Context) {
	// Load tags
	if tags, err := models.Tags(DB, q.Select("id", "tag"), q.OrderBy("tag ASC")).All(); err != nil {
		error500(ctx, err.Error(), f("Error loading tags as %s", "json"))
	} else {
		ctx.JSON(tags)
	}
}

// Tags actions
var Tags = Tag{}
