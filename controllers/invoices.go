package controllers

import (
	"strconv"

	"github.com/kataras/iris/context"
	q "github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/yoander/bolsillo/models"
)

// Invoice unexported
type invoice struct{}

// Read one invoice
func (*invoice) Read(ctx context.Context) {
	ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
	if err != nil {
		error500(ctx, err.Error(), "Editing invoice")
	}
	ctx.ViewData("Title", "Edit Invoice")
	id := uint(ID)
	if id > 0 {
		inv, err := models.FindInvoice(DB, id)
		if err != nil {
			error500(ctx, err.Error(), "Finding invoice: "+string(id))
		}
		ctx.ViewData("action", "Edit")
		ctx.ViewData("inv", inv)
	} else {
		ctx.ViewData("action", "New")
	}
	ctx.View("invoice-form.gohtml")
}

// List invoices
func (*invoice) List(ctx context.Context) {
	// Eager loading
	inv, err := models.Invoices(DB, q.Where("deleted = ?", 0), q.OrderBy("date DESC, id DESC")).All()
	if err != nil {
		error500(ctx, err.Error(), "Loading invoices")
	}
	ctx.ViewData("Title", "Invoices")
	ctx.ViewData("Invoices", inv)
	ctx.View("invoices.gohtml")
}

// DumpAsJSON format
func (*invoice) DumpAsJSON(ctx context.Context) {
	// Load tags
	if tags, err := models.Tags(DB, q.Select("id", "tag"), q.OrderBy("tag ASC")).All(); err != nil {
		error500(ctx, err.Error(), f("Error loading tags as %s", "json"))
	} else {
		ctx.JSON(tags)
	}
}

// Invoices actions
var Invoices = invoice{}
