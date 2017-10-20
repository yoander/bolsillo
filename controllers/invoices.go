package controllers

import (
	"log"
	"math/rand"
	"strconv"
	"time"

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

// Clone invoices
func (*invoice) Clone(ctx context.Context) {
	id := ctx.Params().Get("id")
	log.Print(f("Clonning invoice %s", id))
	if ID, err := strconv.ParseUint(id, 10, 64); err != nil {
		error500(ctx, err.Error(), f("Error clonning invoice %s", ID))
	} else if ID > 0 {
		if inv, err := models.FindInvoice(DB, uint(ID)); err != nil {
			error500(ctx, err.Error(), f("Error clonning invoice %s", ID))
		} else {
			now := time.Now().Local()
			inv.ID = 0
			inv.Code = strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Int())
			inv.Date = now
			inv.CreatedAt = now
			inv.UpdatedAt = now
			if err := inv.Insert(DB); err != nil {
				error500(ctx, err.Error(), f("Error clonning invoice %s", ID))
			}
		}
	} else {
		error500(ctx, f("Invoice %s could not be clonned", ID), f("Error clonning invoice %s", ID))
	}
	ctx.Redirect(ReverseRouter.Path("ListInvoices"))
}

func (*invoice) Save(ctx context.Context) {
	if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err != nil {
		error500(ctx, err.Error(), f("Error saving invoice %d", ID))
	} else {
		var inv models.Invoice
		inv.ID = uint(ID)
		//inv.PersonID = 1
		inv.Status = ctx.PostValue("Status")
		inv.Code = ctx.PostValue("Code")
		inv.Note = ctx.PostValue("Note")
		if date, err := time.Parse("02.01.2006", ctx.PostValue("Date")); err != nil {
			error500(ctx, err.Error(), f("Error saving invoice %d", ID))
		} else {
			inv.Date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)
		}

		if inv.ID > 0 {
			if err := inv.Update(DB); err != nil {
				error500(ctx, err.Error(), f("Error saving invoice %d", ID))
			}
		} else if err := inv.Insert(DB); err != nil {
			error500(ctx, err.Error(), f("Error saving invoice %d", ID))
		}
	}
	ctx.Redirect(ReverseRouter.Path("ListInvoices"))
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
