package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
	"github.com/yoander/bolsillo/controllers"
	"github.com/yoander/bolsillo/models"
)

func main() {
	//p := fmt.Println
	f := fmt.Sprintf
	app := iris.New() // defaults to these
	rv := router.NewRoutePathReverser(app)
	controllers.ReverseRouter = rv
	// Open handle to database like normal
	db, err := sql.Open("mysql", "root@tcp(localhost:2483)/bolsillo?parseTime=true&loc=Local")
	controllers.DB = db
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

	//
	// =================== Invoices ======================
	//
	// List
	app.Get("/invoices", controllers.Invoices.List).Name = "ListInvoices"

	// Edit
	app.Get("/invoice/edit/{id:string}", controllers.Invoices.Read).Name = "EditInvoice"

	// Clone
	app.Get("/invoice/clone/{id:string}", controllers.Invoices.Clone).Name = "CloneInvoice"

	// Save
	app.Post("/invoice/save/{id:string}", controllers.Invoices.Save).Name = "SaveInvoice"

	// Save
	app.Get("/invoice/delete/{id:string}", controllers.Invoices.SofDelete).Name = "DeleteInvoice"

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
			ctx.ViewData("action", "Edit")
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
		} else {
			ctx.ViewData("action", "New")
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

			if isExpensive, err := strconv.ParseUint(ctx.PostValue("IsExpensive"), 10, 8); err != nil {
				error500(ctx, err.Error(), f("Error saving transactions %d", ID))
			} else {
				tx.IsExpensive = int8(isExpensive)
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
	app.Get("tags.json", controllers.Tags.DumpAsJSON).Name = "TagsJSON"

	//
	// =================== Units ======================
	app.Get("units.json", controllers.Units.DumpAsJSON).Name = "UnitsJSON"

	// http://localhost:8080
	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
