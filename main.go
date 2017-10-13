package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/boil"
	q "github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/core/router"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/yoander/bolsillo/models"
)

func main() {
	//p := fmt.Println
	f := fmt.Sprintf
	app := iris.New() // defaults to these
	// Open handle to database like normal
	db, err := sql.Open("mysql", "root@tcp(localhost:2483)/bolsillo?parseTime=true&loc=Local")
	// Optionally set the writer as well. Defaults to os.Stdout
	//fh, err := os.Create("debug.txt")
	//boil.DebugWriter = fh

	if err != nil {
		log.Fatal(err)
	}

	/*simpleDebug := func() string {
		_, fileName, fileLine, ok := runtime.Caller(1)
		var s string
		if ok {
			s = fmt.Sprintf("%s:%d", fileName, fileLine)
		} else {
			s = ""
		}
		return s
	}*/

	boil.DebugMode = false
	app.OnErrorCode(iris.StatusInternalServerError, func(ctx context.Context) {
		ctx.ViewData("Title", "Error!!!")
		ctx.ViewData("header", ctx.Values().GetString("header"))
		ctx.ViewData("message", ctx.Values().GetString("message"))
		ctx.View("error.gohtml")
	})

	error500 := func(ctx context.Context, msg string, header string) {
		ctx.StatusCode(500)
		ctx.Values().Set("header", header)
		ctx.Values().Set("message", msg)
	}

	boil.SetLocation(time.Local)

	app.Use(recover.New())
	app.Use(logger.New())

	rv := router.NewRoutePathReverser(app)

	tmpl := iris.HTML("./views", ".gohtml").Layout("layout.gohtml")
	tmpl.Reload(true) // reload templates on each request (development mode)
	// default template funcs are:
	//
	// - {{ urlpath "mynamedroute" "pathParameter_ifneeded" }}
	// - {{ render "header.html" }}
	// - {{ render_r "header.html" }} // partial relative path to current page
	// - {{ yield }}
	// - {{ current }}

	tmpl.AddFunc("dec", func(num int, step int) int {
		return num - step
	})

	tmpl.AddFunc("inc", func(num int, step int) int {
		return num + step
	})

	tmpl.AddFunc("fdec", func(num float64, step float64) float64 {
		return num - step
	})

	tmpl.AddFunc("finc", func(num float64, step float64) float64 {
		return num + step
	})

	tmpl.AddFunc("strtofloat", func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})

	tmpl.AddFunc("sum_prices", func(transactions models.TransactionSlice) float64 {
		sum := 0.0
		for _, t := range transactions {
			if v, err := strconv.ParseFloat(t.TotalPrice, 64); err == nil {
				sum += v
			}
		}
		return sum
	})

	tmpl.AddFunc("date", func(date time.Time, DateFormat string) string {
		return date.Format(DateFormat)
	})

	tmpl.AddFunc("now", func() time.Time {
		return time.Now()
	})

	tmpl.AddFunc("asset", func(filename string) string {
		return rv.Path("Home") + "assets/" + filename
	})

	tmpl.AddFunc("active", func(URI string, CurrentURI string) string {
		if URI == CurrentURI {
			return " active"
		}
		return ""
	})

	tmpl.AddFunc("in", func(v interface{}, in interface{}) bool {
		val := reflect.Indirect(reflect.ValueOf(in))
		switch val.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < val.Len(); i++ {
				if v == val.Index(i).Interface() {
					return true
				}
			}
		}
		return false
	})

	tmpl.AddFunc("tag_exists", func(tag uint16, tags models.TagSlice) bool {
		for _, t := range tags {
			if t.ID == tag {
				return true
			}
		}
		return false
	})

	tmpl.AddFunc("join_tags", func(tags models.TagSlice, joinField string) string {
		len := len(tags) - 1
		s := ""
		for i, t := range tags {
			if joinField == "id" {
				s += strconv.Itoa(int(t.ID))
			} else if joinField == "tag" {
				s += t.Tag
			}
			if i < len {
				s += ","
			}
		}
		return s
	})

	app.RegisterView(tmpl)

	// Register static content
	app.StaticWeb("/assets", "./assets")

	app.Use(func(ctx context.Context) {
		ctx.Gzip(true)
		ctx.ViewData("URI", ctx.Request().URL.RequestURI())
		ctx.Next()
	})

	// Home / Dashboard
	app.Get("/", func(ctx context.Context) {
		// Eager loading
		ctx.ViewData("Title", "Dashboard")
		ctx.View("dashboard.gohtml")
	}).Name = "Home"

	// Invoices
	app.Get("/invoices", func(ctx context.Context) {
		// Eager loading
		inv, err := models.Invoices(db, q.Where("deleted = ?", 0), q.OrderBy("date DESC, id DESC")).All()
		if err != nil {
			error500(ctx, err.Error(), "Loading invoices")
		}
		ctx.ViewData("Title", "Invoices")
		ctx.ViewData("Invoices", inv)
		ctx.View("invoices.gohtml")
	}).Name = "ListInvoices"

	// Edit invoice
	app.Get("/invoice/edit/{id:string}", func(ctx context.Context) {
		ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64)
		if err != nil {
			error500(ctx, err.Error(), "Editing invoice")
		}
		ctx.ViewData("Title", "Edit Invoice")
		id := uint(ID)
		if id > 0 {
			inv, err := models.FindInvoice(db, id)
			if err != nil {
				error500(ctx, err.Error(), "Finding invoice: "+string(id))
			}
			ctx.ViewData("action", "Edit")
			ctx.ViewData("inv", inv)
		} else {
			ctx.ViewData("action", "New")
		}
		ctx.View("invoice-form.gohtml")
	}).Name = "EditInvoice"

	// Clone invoice
	app.Get("/invoice/clone/{id:string}", func(ctx context.Context) {
		id := ctx.Params().Get("id")
		log.Print(f("Clonning invoice %s", id))
		if ID, err := strconv.ParseUint(id, 10, 64); err != nil {
			error500(ctx, err.Error(), f("Error clonning invoice %s", ID))
		} else if ID > 0 {
			if inv, err := models.FindInvoice(db, uint(ID)); err != nil {
				error500(ctx, err.Error(), f("Error clonning invoice %s", ID))
			} else {
				now := time.Now().Local()
				inv.ID = 0
				inv.Code = strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Int())
				inv.Date = now
				inv.CreatedAt = now
				inv.UpdatedAt = now
				if err := inv.Insert(db); err != nil {
					error500(ctx, err.Error(), f("Error clonning invoice %s", ID))
				}
				log.Println(inv)
			}

		} else {
			error500(ctx, f("Invoice %s could not be clonned", ID), f("Error clonning invoice %s", ID))
		}
		ctx.Redirect(rv.Path("ListInvoices"))
	}).Name = "CloneInvoice"

	// Save invoice
	app.Post("/invoice/save/{id:string}", func(ctx context.Context) {
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
				if err := inv.Update(db); err != nil {
					error500(ctx, err.Error(), f("Error saving invoice %d", ID))
				}
			} else if err := inv.Insert(db); err != nil {
				error500(ctx, err.Error(), f("Error saving invoice %d", ID))
			}
		}
		ctx.Redirect(rv.Path("ListInvoices"))
	}).Name = "SaveInvoice"
	// End of Invoices

	//
	// =================== Transactions ======================
	//
	// List of transactions
	app.Get("/transactions", func(ctx context.Context) {
		// Eager loading
		tran, err := models.Transactions(db, q.Where("deleted = ?", 0), q.OrderBy("date DESC, id DESC"), q.Load("Tags")).All()
		if err != nil {
			error500(ctx, err.Error(), "Error Loading Transactions")
		}

		ctx.ViewData("Title", "Transactions")
		ctx.ViewData("Transactions", tran)
		ctx.View("transactions.gohtml")
	}).Name = "ListTransactions"

	// Edit one transaction
	app.Get("/transaction/edit/{id:string}", func(ctx context.Context) {
		id := ctx.Params().Get("id")
		if ID, err := strconv.ParseUint(id, 10, 64); err != nil {
			error500(ctx, err.Error(), f("Error editing transaction %s", ID))
		} else if ID > 0 {
			if tx, err := models.FindTransaction(db, uint(ID)); err != nil {
				error500(ctx, err.Error(), f("Error editing transaction %s", ID))
			} else {
				ctx.ViewData("tx", tx)
				if txTags, err := tx.Tags(db, q.Select("id, tag")).All(); err != nil {
					error500(ctx, err.Error(), f("Error editing transaction %s", ID))
				} else {
					ctx.ViewData("txTags", txTags)
				}
			}
		}

		if invoices, err := models.Invoices(db, q.Select("id", "code", "date", "note"), q.OrderBy("date DESC")).All(); err != nil {
			error500(ctx, err.Error(), f("Error editing transaction %s", id))
		} else {
			ctx.ViewData("invoices", invoices)
		}
		// Load tags
		if tags, err := models.Tags(db, q.Select("id", "tag"), q.OrderBy("tag ASC")).All(); err != nil {
			error500(ctx, err.Error(), f("Error editing transaction %s", id))
		} else {
			ctx.ViewData("Tags", tags)
		}
		if units, err := models.Units(db, q.Select("id", "name", "symbol"), q.OrderBy("name ASC")).All(); err != nil {
			error500(ctx, err.Error(), f("Error editing transaction %s", id))
		} else {
			ctx.ViewData("Units", units)
		}

		ctx.View("transaction-form.gohtml")
	}).Name = "EditTransaction" // Also New

	// Clone transaction
	app.Get("/transaction/clone/{id:string}", func(ctx context.Context) {
		if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err != nil {
			error500(ctx, err.Error(), "Error Loading Transactions")
		} else if id := uint(ID); id > 0 {
			if tx, err := models.FindTransaction(db, id); err != nil {
				error500(ctx, err.Error(), "Error Loading Transactions")
			} else if txTags, err := tx.Tags(db, q.Select("ID")).All(); err != nil {
				error500(ctx, err.Error(), "Error Loading Transactions")
			} else {
				now := time.Now().Local()
				tx.ID = 0
				tx.Date = now
				tx.CreatedAt = now
				tx.UpdatedAt = now
				if err := tx.Insert(db); err != nil {
					error500(ctx, err.Error(), "Error Loading Transactions")
				}
				if len(txTags) > 0 {
					tx.SetTags(db, false, txTags...)
				}
			}
		}
		ctx.Redirect(rv.Path("ListTransactions"))
	}).Name = "CloneTransaction" // Also New

	// Save a transaction
	app.Post("/transaction/save/{id:string}", func(ctx context.Context) {
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

			updTags := true

			if tx.ID > 0 {
				if err := tx.Update(db); err != nil {
					updTags = false
					error500(ctx, err.Error(), f("Error saving transactions %d", ID))
				}
			} else if err := tx.Insert(db); err != nil {
				error500(ctx, err.Error(), f("Error saving transactions %d", ID))
			}

			if updTags {
				var IDs []interface{}
				if err := json.NewDecoder(strings.NewReader(ctx.PostValue("Tags"))).Decode(&IDs); err != nil {
					error500(ctx, err.Error(), f("Error saving transactions %d", ID))
				}

				if len(IDs) > 0 {
					if tags, err := models.Tags(db, q.Select("id"), q.WhereIn("id IN ?", IDs...)).All(); err != nil {
						error500(ctx, err.Error(), f("Error saving transactions %d", ID))
					} else if err := tx.SetTags(db, false, tags...); err != nil {
						error500(ctx, err.Error(), f("Error saving transactions %d", ID))
					}
				}
			}
		}
		ctx.Redirect(rv.Path("ListTransactions"))
	}).Name = "SaveTransaction"

	// Delete one transaction
	app.Get("/transaction/delete/{id:string}", func(ctx context.Context) {
		if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err != nil {
			error500(ctx, err.Error(), f("Error deleting transactions %d", ID))
		} else if ID > 0 {
			if tx, err := models.FindTransaction(db, uint(ID)); err != nil {
				error500(ctx, err.Error(), f("Error deleting transactions %d", ID))
			} else {
				tx.Deleted = 1
				if err := tx.Update(db); err != nil {
					error500(ctx, err.Error(), f("Error deleting transactions %d", ID))
				}
			}
		}

		ctx.Redirect(rv.Path("ListTransactions"))
	}).Name = "DeleteTransaction"
	//
	// =================== End of Transactions ======================
	//

	//
	// =================== Tags ======================
	//
	app.Get("tags.json", func(ctx context.Context) {
		// Load tags
		if tags, err := models.Tags(db, q.Select("id", "tag"), q.OrderBy("tag ASC")).All(); err != nil {
			error500(ctx, err.Error(), f("Error loading tags as %s", "json"))
		} else {
			ctx.JSON(tags)
		}
	}).Name = "TagsJSON"

	// http://localhost:8080
	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
