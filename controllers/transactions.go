package controllers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/context"
	q "github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/yoander/bolsillo/models"
)

// transaction unexported
type transaction struct{}

// Read one transaction
func (*transaction) Read(ctx context.Context) {
	id := ctx.Params().Get("id")
	if ID, err := strconv.ParseUint(id, 10, 64); err != nil {
		error500(ctx, err.Error(), f("Error editing transaction %s", ID))
	} else if ID > 0 {
		ctx.ViewData("action", "Edit")
		if tx, err := models.FindTransaction(DB, uint(ID)); err != nil {
			error500(ctx, err.Error(), f("Error editing transaction %s", ID))
		} else {
			ctx.ViewData("tx", tx)
			if txTags, err := tx.Tags(DB, q.Select("id, tag")).All(); err != nil {
				error500(ctx, err.Error(), f("Error editing transaction %s", ID))
			} else {
				ctx.ViewData("txTags", txTags)
			}
		}
	} else {
		ctx.ViewData("action", "New")
	}

	if invoices, err := models.Invoices(DB, q.Select("id", "code", "date", "note"), q.Where("deleted = 0"), q.OrderBy("date DESC")).All(); err != nil {
		error500(ctx, err.Error(), f("Error editing transaction %s", id))
	} else {
		ctx.ViewData("invoices", invoices)
	}
	// Load tags
	if tags, err := models.Tags(DB, q.Select("id", "tag"), q.OrderBy("tag ASC")).All(); err != nil {
		error500(ctx, err.Error(), f("Error editing transaction %s", id))
	} else {
		ctx.ViewData("Tags", tags)
	}

	if units, err := models.Units(DB, q.Select("id", "name", "symbol"), q.OrderBy("name ASC")).All(); err != nil {
		error500(ctx, err.Error(), f("Error editing transaction %s", id))
	} else {
		ctx.ViewData("Units", units)
	}

	ctx.View("transaction-form.gohtml")
}

// List transactions
func (*transaction) GetFilteredTransactions(startDate string, endDate string, keyword string) (models.TransactionSlice, error) {
	sDate, err := time.Parse("02.01.2006", startDate)
	eDate, err := time.Parse("02.01.2006", endDate)

	if err != nil {
		now := time.Now().Local()
		sDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		eDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	}

	// Eager loading
	queries := []q.QueryMod{}
	queries = append(queries, q.Where("deleted = ?", 0))
	queries = append(queries, q.And("date BETWEEN ? AND ?", sDate, eDate))
	if keyword != "" {
		queries = append(queries, q.And("CONCAT_WS(':', description, note) LIKE ?", "%"+keyword+"%"))
	}
	queries = append(queries, q.OrderBy("date DESC, id DESC"))
	queries = append(queries, q.Load("Tags"))

	transactions, err := models.Transactions(DB, queries...).All()

	return transactions, err
}

// Clone transactions
func (*transaction) Clone(ctx context.Context) {
	if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err != nil {
		error500(ctx, err.Error(), "Error Loading Transactions")
	} else if id := uint(ID); id > 0 {
		if tx, err := models.FindTransaction(DB, id); err != nil {
			error500(ctx, err.Error(), "Error Loading Transactions")
		} else if txTags, err := tx.Tags(DB, q.Select("ID")).All(); err != nil {
			error500(ctx, err.Error(), "Error Loading Transactions")
		} else {
			now := time.Now().Local()
			tx.ID = 0
			tx.Date = now
			tx.CreatedAt = now
			tx.UpdatedAt = now
			if err := tx.Insert(DB); err != nil {
				error500(ctx, err.Error(), "Error Loading Transactions")
			}
			if len(txTags) > 0 {
				tx.SetTags(DB, false, txTags...)
			}
		}
	}
	ctx.Redirect(ReverseRouter.Path("ListTransactions"))
}

// Save one transaction
func (*transaction) Save(ctx context.Context) {
	if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err != nil {
		error500(ctx, err.Error(), f("Error saving transactions %d", ID))
	} else {
		var tx models.Transaction
		tx.ID = uint(ID)
		tx.PersonID = 1
		tx.Type = ctx.PostValue("Type")
		tx.Description = ctx.PostValue("Description")
		tx.Note = ctx.PostValue("Note")
		tx.TotalPrice = ctx.PostValue("TotalPrice")
		tx.Quantity = ctx.PostValue("Quantity")
		tx.Price = ctx.PostValue("UnitPrice")
		tx.Status = ctx.PostValue("Status")
		if date, err := time.Parse("02.01.2006", ctx.PostValue("Date")); err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", ID))
		} else {
			tx.Date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)
		}

		if invID, err := strconv.ParseUint(ctx.PostValue("Invoice"), 10, 8); err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", ID))
		} else if invID > 0 {
			tx.InvoiceID.SetValid(uint(invID))
		}

		if unitID, err := strconv.ParseUint(ctx.PostValue("Unit"), 10, 8); err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", ID))
		} else if unitID > 0 {
			tx.UnitID.SetValid(uint8(unitID))
		}

		if isExpensive, err := strconv.ParseUint(ctx.PostValue("IsExpensive"), 10, 8); err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", ID))
		} else {
			tx.IsExpensive = int8(isExpensive)
		}

		updTags := true

		if tx.ID > 0 {
			if err := tx.Update(DB); err != nil {
				updTags = false
				error500(ctx, err.Error(), f("Error saving transactions %d", ID))
			}
		} else if err := tx.Insert(DB); err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", ID))
		}

		if updTags {
			var IDs []interface{}
			if err := json.NewDecoder(strings.NewReader(ctx.PostValue("Tags"))).Decode(&IDs); err != nil {
				error500(ctx, err.Error(), f("Error saving transactions %d", ID))
			}

			if len(IDs) > 0 {
				if tags, err := models.Tags(DB, q.Select("id"), q.WhereIn("id IN ?", IDs...)).All(); err != nil {
					error500(ctx, err.Error(), f("Error saving transactions %d", ID))
				} else if err := tx.SetTags(DB, false, tags...); err != nil {
					error500(ctx, err.Error(), f("Error saving transactions %d", ID))
				}
			}
		}
	}
	ctx.Redirect(ReverseRouter.Path("ListTransactions"))
}

// SoftDelete mark an transaction as deleted
func (*transaction) SoftDelete(ctx context.Context) {
	if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err != nil {
		error500(ctx, err.Error(), f("Error deleting transactions %d", ID))
	} else if ID > 0 {
		if tx, err := models.FindTransaction(DB, uint(ID)); err != nil {
			error500(ctx, err.Error(), f("Error deleting transactions %d", ID))
		} else {
			tx.Deleted = 1
			if err := tx.Update(DB); err != nil {
				error500(ctx, err.Error(), f("Error deleting transactions %d", ID))
			}
		}
	}

	ctx.Redirect(ReverseRouter.Path("ListTransactions"))
}

// DumpAsJSON format
func (*transaction) DumpAsJSON(ctx context.Context) {

}

// Transactions actions
var Transactions = transaction{}
