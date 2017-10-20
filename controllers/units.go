package controllers

import (
	"github.com/kataras/iris/context"
	q "github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/yoander/bolsillo/models"
)

// Unit define a TagMethods implementation
type Unit struct{}

// DumpAsJSON dumps tags
func (tag *Unit) DumpAsJSON(ctx context.Context) {
	// Load tags
	if tags, err := models.Tags(DB, q.Select("id", "tag"), q.OrderBy("tag ASC")).All(); err != nil {
		error500(ctx, err.Error(), f("Error loading tags as %s", "json"))
	} else {
		ctx.JSON(tags)
	}
}

// Units actions
var Units = Unit{}
