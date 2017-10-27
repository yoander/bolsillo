package controllers

import (
	"strconv"

	"github.com/kataras/iris/context"
	q "github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/yoander/bolsillo/models"
)

// tag unexported
type tag struct{}

// Read one tag
func (*tag) Read(ctx context.Context) {
	ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		error500(ctx, err.Error(), "Editing tag")
	}
	ctx.ViewData("Title", "Edit tag")
	id := uint16(ID)
	if id > 0 {
		tag, err := models.FindTag(DB, id)
		if err != nil {
			error500(ctx, err.Error(), "Finding tag: "+string(id))
		}
		ctx.ViewData("action", "Edit")
		ctx.ViewData("inv", tag)
	} else {
		ctx.ViewData("action", "New")
	}
	ctx.View("tag-form.gohtml")
}

// List tags
func (*tag) List(ctx context.Context) {
	// Eager loading
	tags, err := models.Tags(DB).All()
	if err != nil {
		error500(ctx, err.Error(), "Loading tags")
	}
	ctx.ViewData("Title", "Tags")
	ctx.ViewData("Tags", tags)
	ctx.View("tags.gohtml")
}

// Save one transaction
func (*tag) Save(ctx context.Context) {
	if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err != nil {
		error500(ctx, err.Error(), f("Error saving tag %d", ID))
	} else {
		var tag models.Tag
		tag.ID = uint16(ID)
		tag.Tag = ctx.PostValue("Tag")

		if tag.ID > 0 {
			if err := tag.Update(DB); err != nil {
				error500(ctx, err.Error(), f("Error saving tag %d", ID))
			}
		} else if err := tag.Insert(DB); err != nil {
			error500(ctx, err.Error(), f("Error saving tag %d", ID))
		}
	}
	ctx.Redirect(ReverseRouter.Path("ListTags"))
}

// Delete mark an tag as deleted
func (*tag) Delete(ctx context.Context) {
	if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err != nil {
		error500(ctx, err.Error(), f("Error deleting transactions %d", ID))
	} else if ID > 0 {
		if tag, err := models.FindTag(DB, uint16(ID)); err != nil {
			error500(ctx, err.Error(), f("Error deleting transactions %d", ID))
		} else {
			tag.Delete(DB)
		}
	}

	ctx.Redirect(ReverseRouter.Path("ListTags"))
}

// DumpAsJSON format
func (*tag) DumpAsJSON(ctx context.Context) {
	// Load tags
	if tags, err := models.Tags(DB, q.Select("id", "tag"), q.OrderBy("tag ASC")).All(); err != nil {
		error500(ctx, err.Error(), f("Error loading tags as %s", "json"))
	} else {
		ctx.JSON(tags)
	}
}

// Tags actions
var Tags = tag{}
