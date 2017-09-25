package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	app := iris.New() // defaults to these
	p := fmt.Println
	// Open handle to database like normal
	db, err := sql.Open("mysql", "root@tcp(localhost:2483)/bolsillo?parseTime=true&loc=Local")
	boil.DebugMode = true
	// Optionally set the writer as well. Defaults to os.Stdout
	//fh, err := os.Create("debug.txt")
	//boil.DebugWriter = fh

	if err != nil {
		p("Error opening connection", err)
		return
	}

	boil.SetLocation(time.Local)

	app.Use(recover.New())
	app.Use(logger.New())

	rv := router.NewRoutePathReverser(app)

	tmpl := iris.HTML("./views", ".html").Layout("layout.go.html")
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

	tmpl.AddFunc("css", func(filename string) string {
		return rv.Path("Home") + "assets/css/" + filename
	})

	tmpl.AddFunc("cssv", func(filename string) string {
		return rv.Path("Home") + "assets/vendor/css/" + filename
	})

	tmpl.AddFunc("fontv", func(filename string) string {
		return rv.Path("Home") + "assets/vendor/fonts/" + filename
	})

	tmpl.AddFunc("js", func(filename string) string {
		return rv.Path("Home") + "assets/js/" + filename
	})

	tmpl.AddFunc("jsv", func(filename string) string {
		return rv.Path("Home") + "assets/vendor/js/" + filename
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
		tran, err := models.Transactions(db, q.OrderBy("date DESC, invoice_id DESC, id DESC"), q.Load("Tags")).All()
		if err != nil {
			p("Error Loading Transactions", err)
		} else {
			ctx.ViewData("Title", "Dashboard")
			//ctx.ViewData("Name", "iris") // {{.Name}} will render: iris
			ctx.ViewData("Transactions", tran)
			ctx.View("dashboard.go.html")
		}
	}).Name = "Home"

	// List of transactions
	app.Get("/invoices", func(ctx context.Context) {
		// Eager loading
		if inv, err := models.Invoices(db, q.Where("deleted = ?", 0), q.OrderBy("date DESC, id DESC")).All(); err == nil {
			ctx.ViewData("Title", "Invoices")
			//ctx.ViewData("Name", "iris") // {{.Name}} will render: iris
			ctx.ViewData("Invoices", inv)
			ctx.View("invoices.go.html")
		} else {
			p(err)
		}
		//ctx.ViewData("route", ctx.GetCurrentRoute().Path())
	}).Name = "ListInvoices"

	// List of transactions
	app.Get("/transactions", func(ctx context.Context) {
		// Eager loading
		if tran, err := models.Transactions(db, q.Where("deleted = ?", 0), q.OrderBy("date DESC, invoice_id DESC, id DESC"), q.Load("Tags")).All(); err == nil {
			ctx.ViewData("Title", "Transactions")
			//ctx.ViewData("Name", "iris") // {{.Name}} will render: iris
			ctx.ViewData("Transactions", tran)
			ctx.View("transactions.go.html")
		} else {
			p(err)
		}
		//ctx.ViewData("route", ctx.GetCurrentRoute().Path())
	}).Name = "ListTransactions"

	// List a transaction
	app.Get("/transactions/transaction/{id:string}", func(ctx context.Context) {
		ctx.ViewData("Title", "Edit transaction")

		// Load transaction
		if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err == nil {
			id := uint(ID)
			if id > 0 {
				if tx, err := models.FindTransaction(db, id); err == nil {
					ctx.ViewData("tx", tx)
					if txTags, err := tx.Tags(db, q.Select("ID")).All(); err == nil {
						ctx.ViewData("txTags", txTags)
					} else {
						p(err)
					}
				}
			}
		} else {
			p(err)
		}

		// Load tags
		if invoices, err := models.Invoices(db, q.Select("id", "code", "date", "note"), q.OrderBy("date DESC")).All(); err == nil {
			ctx.ViewData("invoices", invoices)
		} else {
			p(err)
		}

		// Load tags
		if tags, err := models.Tags(db, q.Select("id", "tag"), q.OrderBy("tag ASC")).All(); err == nil {
			ctx.ViewData("Tags", tags)
		} else {
			//ctx.ViewData("Error", "Error Loading Tags")
			p(err)
		}

		if units, err := models.Units(db, q.Select("id", "name", "symbol"), q.OrderBy("name ASC")).All(); err == nil {
			ctx.ViewData("Units", units)
		} else {
			p(err)
		}

		ctx.View("transaction-form.go.html")

	}).Name = "EditTransaction" // Also New

	// Clone a transaction
	app.Get("/transactions/transaction/clone/{id:string}", func(ctx context.Context) {
		// Load transaction
		if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err == nil {
			id := uint(ID)
			if id > 0 {
				if tx, err := models.FindTransaction(db, id); err == nil {
					txTags, err := tx.Tags(db, q.Select("ID")).All()
					if err != nil {
						p(err)
					}
					now := time.Now().Local()
					tx.ID = 0
					tx.Date = now
					tx.CreatedAt = now
					tx.UpdatedAt = now
					if err := tx.Insert(db); err == nil {
						if len(txTags) > 0 {
							tx.SetTags(db, false, txTags...)
						}
					} else {
						p(err)
					}
				}
			}
		} else {
			p(err)
		}
		ctx.Redirect(rv.Path("ListTransactions"))
	}).Name = "CloneTransaction" // Also New

	// Save a transaction
	app.Post("/transactions/transaction/{id:string}", func(ctx context.Context) {
		var tx models.Transaction
		if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err == nil {
			tx.ID = uint(ID)
			tx.PersonID = 1
			tx.Type = ctx.PostValue("Type")
			tx.Description = ctx.PostValue("Description")
			if date, err := time.Parse("02 January, 2006", ctx.PostValue("Date")); err == nil {
				// Add local time offset to date in order to avoid the DB driver insert as -1 day
				now := time.Now().Local()
				duration := time.Duration(now.Hour()) * time.Hour
				duration += time.Duration(now.Minute()) * time.Minute
				duration += time.Duration(now.Second()) * time.Second
				duration += time.Duration(now.Nanosecond()) * time.Nanosecond
				tx.Date = date.Add(duration)
			} else {
				p(err)
			}

			tx.Note = ctx.PostValue("Note")
			tx.TotalPrice = ctx.PostValue("TotalPrice")
			tx.Quantity = ctx.PostValue("Quantity")
			tx.Price = ctx.PostValue("UnitPrice")

			if invoiceID, err := strconv.ParseUint(ctx.PostValue("Invoice"), 10, 8); err == nil && invoiceID > 0 {
				tx.InvoiceID.SetValid(uint(invoiceID))
			} else {
				p(err)
			}

			if unitID, err := strconv.ParseUint(ctx.PostValue("Unit"), 10, 8); err == nil && unitID > 0 {
				tx.UnitID.SetValid(uint8(unitID))
			} else {
				p(err)
			}

			updTags := true

			if tx.ID > 0 {
				if err := tx.Update(db); err != nil {
					updTags = false
					p(err)
				}
			} else if err := tx.Insert(db); err != nil {
				updTags = false
				p(err)
			}

			if updTags {
				var IDs []interface{}
				if err := json.NewDecoder(strings.NewReader(ctx.PostValue("Tags"))).Decode(&IDs); len(IDs) > 0 {
					if tags, err := models.Tags(db, q.Select("id"), q.WhereIn("id IN ?", IDs...)).All(); err == nil {
						if err := tx.SetTags(db, false, tags...); err != nil {
							p(err)
						}
					} else {
						p(err)
					}
				} else if err != nil {
					p(err)
				}
			}
		} else {
			p(err)
		}

		ctx.Redirect(rv.Path("ListTransactions"))
	}).Name = "SaveTransaction"

	// Clone a transaction
	app.Get("/transactions/transaction/delete/{id:string}", func(ctx context.Context) {
		// Load transaction
		if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err == nil {
			id := uint(ID)
			if id > 0 {
				if tx, err := models.FindTransaction(db, id); err == nil {
					tx.Deleted = 1
					if err := tx.Update(db); err != nil {
						p(err)
					}
				}
			}
		} else {
			p(err)
		}
		ctx.Redirect(rv.Path("ListTransactions"))
	}).Name = "DeleteTransaction"

	// http://localhost:8080
	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
