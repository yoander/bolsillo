package controllers

import (
	"fmt"
	"strconv"
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
func (*transaction) GetFilteredTransactions(startDate time.Time,
	endDate time.Time,
	keyword string) (models.TransactionSlice, error) {
	// Eager loading
	queries := []q.QueryMod{}
	queries = append(queries, q.Where("deleted = ?", 0))
	queries = append(queries, q.And("date BETWEEN ? AND ?", startDate, endDate))
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
func (*transaction) Save(ID uint,
	invoiceID uint,
	unitID uint8,
	userID uint16,
	identityID uint,
	transactionType string,
	desc string,
	note string,
	quantity string,
	unitPrice string,
	totalPrice string,
	status string,
	date time.Time,
	isExpensive int8,
	tags []string,
) error {
	//fmt.Println("Saving tx....")
	// Transaction Object
	var tx models.Transaction
	tx.ID = ID
	tx.PersonID = userID
	tx.Type = transactionType
	tx.Description = desc
	tx.Note = note
	tx.Quantity = quantity
	tx.Price = unitPrice
	tx.TotalPrice = totalPrice
	tx.Status = status
	tx.IsExpensive = isExpensive
	tx.Date = date
	if invoiceID > 0 {
		tx.InvoiceID.SetValid(invoiceID)
	}
	if unitID > 0 {
		tx.UnitID.SetValid(unitID)
	}

	var err error
	if ID > 0 {
		err = tx.Update(DB)
	} else {
		err = tx.Insert(DB)
	}
	if err != nil {
		fmt.Println("Error saving", invoiceID, err)
		return err
	}

	//fmt.Println("tags in save", tags)
	count := len(tags)
	if count > 0 {
		// See https://github.com/golang/go/wiki/InterfaceSlice
		tagSet := make([]interface{}, count)
		for i, t := range tags {
			tagSet[i] = t
		}
		//fmt.Println("saving tags", tagSet)
		tagEntities, err := models.Tags(DB, q.Select("id"), q.WhereIn("tag IN ?", tagSet...)).All()
		if err == nil {
			err = tx.SetTags(DB, false, tagEntities...)
		}
	}

	/*if err != nil {
		fmt.Println("Error saving", tags, err)
	}*/

	return err
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
