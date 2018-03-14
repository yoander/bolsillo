package controllers

import (
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
func (*transaction) List(startDate time.Time,
	endDate time.Time,
	keyword string) (models.TransactionSlice, float64, float64, float64, error) {
	sqlBuilder := []q.QueryMod{}
	sqlBuilder = append(sqlBuilder, q.Select("id, description, type, status, price, total_price, date, note, unit_id"))
	sqlBuilder = append(sqlBuilder, q.Where("deleted = ?", 0))
	sqlBuilder = append(sqlBuilder, q.And("date BETWEEN ? AND ?", startDate, endDate))
	if keyword != "" {
		sqlBuilder = append(sqlBuilder, q.And("CONCAT_WS(':', description, note) LIKE ?", "%"+keyword+"%"))
	}
	sqlBuilder = append(sqlBuilder, q.OrderBy("date DESC, id DESC"))
	sqlBuilder = append(sqlBuilder, q.Load("Tags", "Unit"))

	transactions, err := models.Transactions(DB, sqlBuilder...).All()

	if err != nil || len(transactions) == 0 {
		return transactions, 0, 0, 0, err
	}

	sqlBuilder = nil
	sqlBuilder = append(sqlBuilder,
		q.Select("IFNULL(SUM(total_price), 0) AS expenses"),
		//	q.From("transactions"),
		q.Where("type = ?", "EXP"),
		q.And("deleted = ?", 0),
		q.And("date BETWEEN ? AND ?", startDate, endDate))

	if keyword != "" {
		sqlBuilder = append(sqlBuilder, q.And("CONCAT_WS(':', description, note) LIKE ?", "%"+keyword+"%"))
	}

	row := models.Transactions(DB, sqlBuilder...).QueryRow()
	var expenses float64
	if err := row.Scan(&expenses); err != nil {
		return transactions, 0, 0, 0, err
	}

	sqlBuilder = nil
	sqlBuilder = append(sqlBuilder,
		q.Select("IFNULL(SUM(total_price), 0) AS incomes"),
		//q.From("transactions"),
		q.Where("type = ?", "INC"),
		q.And("deleted = ?", 0),
		q.And("date BETWEEN ? AND ?", startDate, endDate))

	if keyword != "" {
		sqlBuilder = append(sqlBuilder, q.And("CONCAT_WS(':', description, note) LIKE ?", "%"+keyword+"%"))
	}

	row = models.Transactions(DB, sqlBuilder...).QueryRow()
	var incomes float64

	if err := row.Scan(&incomes); err != nil {
		return transactions, expenses, 0, -expenses, err
	}

	return transactions, expenses, incomes, incomes - expenses, err
}

// List transactions
func (*transaction) GetByInvoiceID(invoiceID uint64) (models.TransactionSlice, error) {
	sqlBuilder := []q.QueryMod{}
	sqlBuilder = append(sqlBuilder, q.Select("id, description, type, status, price, total_price, date, note, unit_id"))
	sqlBuilder = append(sqlBuilder, q.Where("deleted = ?", 0))
	sqlBuilder = append(sqlBuilder, q.And("invoice_id IS NOT NULL AND invoice_id = ?", invoiceID))
	sqlBuilder = append(sqlBuilder, q.OrderBy("date DESC, id DESC"))
	sqlBuilder = append(sqlBuilder, q.Load("Tags", "Unit"))

	return models.Transactions(DB, sqlBuilder...).All()
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
		return err
	}

	count := len(tags)
	if (err != nil) && (count > 0) {
		// See https://github.com/golang/go/wiki/InterfaceSlice
		tagSet := make([]interface{}, count)
		for i, t := range tags {
			tagSet[i] = t
		}
		tagEntities, err := models.Tags(DB, q.Select("id"), q.WhereIn("tag IN ?", tagSet...)).All()
		if err == nil {
			err = tx.SetTags(DB, false, tagEntities...)
		}
	}

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
