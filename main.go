package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	q "github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/core/router"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/yoander/bolsillo/models"
)

func main() {
	// Open handle to database like normal
	db, err := sql.Open("mysql", "root@tcp(localhost:2483)/bolsillo?parseTime=true")
	if err != nil {
		fmt.Println("Error opening connection", err)
		return
	}

	app := iris.New() // defaults to these

	// Optionally, add two built'n handlers
	// that can recover from any http-relative panics
	// and log the requests to the terminal.
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

	tmpl.AddFunc("active", func(RouteName string, CurrentURI string) string {
		if rv.Path(RouteName) == CurrentURI {
			return " active"
		}
		return ""
	})

	app.RegisterView(tmpl)

	// Register static content
	app.StaticWeb("/assets", "./assets")

	app.Use(func(ctx context.Context) {
		ctx.ViewData("URI", ctx.GetCurrentRoute().Path())
		ctx.Next()
	})

	app.Get("/", func(ctx context.Context) {
		// Eager loading
		tran, err := models.Transactions(db, q.OrderBy("date DESC, invoice_id DESC, id DESC"), q.Load("Tags")).All()
		if err != nil {
			fmt.Println("Error Loading Transactions", err)
		} else {
			ctx.Gzip(true)
			ctx.ViewData("Title", "Dashboard")
			//ctx.ViewData("Name", "iris") // {{.Name}} will render: iris
			ctx.ViewData("Transactions", tran)
			ctx.View("dashboard.go.html")
		}
	}).Name = "Home"

	app.Get("/transactions/transaction", func(ctx context.Context) {
		// Eager loading
		tags, err := models.Tags(db, q.OrderBy("tag ASC")).All()
		if err != nil {
			fmt.Println("Error Loading Tags", err)
		} else {
			ctx.Gzip(true)
			ctx.ViewData("Title", "Dashboard")
			ctx.ViewData("Tags", tags)
			//	ctx.ViewData("route", ctx.GetCurrentRoute().Path())
			ctx.View("transaction-form.go.html")
		}

	}).Name = "EditTransaction" // Also New

	app.Get("/transactions", func(ctx context.Context) {
		//ctx.ViewData("route", ctx.GetCurrentRoute().Path())
	}).Name = "ListTransactions"

	app.Post("/transactions/transaction", func(ctx context.Context) {
		// Eager loading
		tags, err := models.Tags(db, q.OrderBy("tag ASC")).All()
		if err != nil {
			fmt.Println("Error Loading Tags", err)
		} else {
			ctx.Gzip(true)
			ctx.ViewData("Tags", tags)
			ctx.View("transaction-form.go.html")
		}

	}).Name = "SaveTransaction"

	// http://localhost:8080
	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
