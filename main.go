package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	boil.DebugMode = false
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

	// List a transaction
	app.Get("/transactions/transaction/{id:string}", func(ctx context.Context) {
		ctx.ViewData("Title", "Edit transaction")

		// Load transaction
		if ID, err := strconv.ParseUint(ctx.Params().Get("id"), 10, 64); err == nil {
			id := uint(ID)
			if id > 0 {
				if tx, err := models.FindTransaction(db, id); err == nil {
					ctx.ViewData("tx", tx)
				}
			}
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

	// List of transactions
	app.Get("/transactions", func(ctx context.Context) {
		//ctx.ViewData("route", ctx.GetCurrentRoute().Path())
	}).Name = "ListTransactions"

	// Save a transaction
	app.Post("/transactions/transaction", func(ctx context.Context) {
		var tx models.Transaction
		if ID, err := strconv.ParseUint(ctx.PostValue("ID"), 10, 64); err == nil {
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
			tx.Price = "0.00"

			if unitID, err := strconv.ParseUint(ctx.PostValue("Unit"), 10, 8); err == nil {
				tx.UnitID.SetValid(uint8(unitID))
			} else {
				p(err)
			}

			updTags := true
			if tx.ID == 0 {
				if err := tx.Insert(db); err != nil {
					updTags = false
					p(err)
				}

			} else {
				if err := tx.Update(db); err != nil {
					updTags = false
					p(err)
				}
			}

			if updTags {
				var IDs []interface{}
				if err := json.NewDecoder(strings.NewReader(ctx.PostValue("Tags"))).Decode(&IDs); len(IDs) > 0 {
					if tags, err := models.Tags(db, q.WhereIn("id IN ?", IDs...)).All(); err == nil {
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

		ctx.Redirect(rv.Path("EditTransaction"))
	}).Name = "SaveTransaction"

	// http://localhost:8080
	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
