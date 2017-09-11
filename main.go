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
		//r := app.GetRoute(ctx.GetCurrentRoute().Name())
		//r.ResolvePath( ...)
		ctx.ViewData("URI", ctx.Request().URL.RequestURI())
		ctx.Next()
	})

	app.Get("/", func(ctx context.Context) {
		// Eager loading
		tran, err := models.Transactions(db, q.OrderBy("date DESC, invoice_id DESC, id DESC"), q.Load("Tags")).All()
		if err != nil {
			fmt.Println("Error Loading Transactions", err)
		} else {
			ctx.ViewData("Title", "Dashboard")
			//ctx.ViewData("Name", "iris") // {{.Name}} will render: iris
			ctx.ViewData("Transactions", tran)
			ctx.View("dashboard.go.html")
		}
	}).Name = "Home"

	app.Get("/transactions/transaction", func(ctx context.Context) {
		ctx.ViewData("Title", "Edit transaction")
		tags, err := models.Tags(db, q.OrderBy("tag ASC")).All()
		if err != nil {
			ctx.ViewData("Error", "Error Loading Tags")
			fmt.Println("Error Loading Tags", err)
		} else {
			ctx.ViewData("Tags", tags)
		}

		units, err := models.Units(db, q.OrderBy("name ASC")).All()
		if err != nil {
			ctx.ViewData("Error", "Error Loading Units")
			fmt.Println("Error Loading Units", err)
		} else {
			ctx.ViewData("Units", units)
		}
		ctx.View("transaction-form.go.html")

	}).Name = "EditTransaction" // Also New

	app.Get("/transactions", func(ctx context.Context) {
		//ctx.ViewData("route", ctx.GetCurrentRoute().Path())
	}).Name = "ListTransactions"

	app.Post("/transactions/transaction", func(ctx context.Context) {
		var tx models.Transaction
		tx.ID = 104
		tx.PersonID = 1
		tx.Type = ctx.PostValue("Type")
		tx.Description.SetValid(ctx.PostValue("Description"))
		if date, err := time.Parse("02 January, 2006", ctx.PostValue("Date")); err == nil {
			tx.Date = date.In(boil.GetLocation())
		} else {
			p(err)
		}

		tx.Note = ctx.PostValue("Note")
		tx.TotalPrice = ctx.PostValue("TotalPrice")
		tx.Quantity = ctx.PostValue("Quantity")

		if unitID, err := strconv.ParseUint(ctx.PostValue("Unit"), 10, 8); err == nil {
			tx.UnitID.SetValid(uint8(unitID))
			/*if unit, err := models.FindUnit(db, uint8(unitID)); err == nil {
				tx.SetUnit(db, false, unit)
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
