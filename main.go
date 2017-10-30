package main

import (
	"database/sql"
	"log"
	"reflect"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/boil"

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
	//f := fmt.Sprintf
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

	boil.DebugMode = true
	app.OnErrorCode(iris.StatusInternalServerError, func(ctx context.Context) {
		ctx.ViewData("Title", "Error!!!")
		ctx.ViewData("header", ctx.Values().GetString("header"))
		message := ctx.Values().GetString("message")
		if message == "" {
			message = "Internal Server error"
		}

		ctx.ViewData("message", message)

		ctx.View("error.gohtml")
	})

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
	app.Get("/transactions", controllers.Transactions.List).Name = "ListTransactions"

	// Edit one transaction
	app.Get("/transaction/edit/{id:string}", controllers.Transactions.Read).Name = "EditTransaction" // Also New

	// Clone transaction
	app.Get("/transaction/clone/{id:string}", controllers.Transactions.Clone).Name = "CloneTransaction" // Also New

	// Save a transaction
	app.Post("/transaction/save/{id:string}", controllers.Transactions.Save).Name = "SaveTransaction"

	// Delete one transaction
	app.Get("/transaction/delete/{id:string}", controllers.Transactions.SofDelete).Name = "DeleteTransaction"

	//
	// =================== Tags ======================
	//
	app.Get("tags.json", controllers.Tags.DumpAsJSON).Name = "TagsJSON"

	// List
	app.Get("/tags", controllers.Tags.List).Name = "ListTags"

	// Edit
	app.Get("/tag/edit/{id:string}", controllers.Tags.Read).Name = "EditTag"

	// Save
	app.Post("/tag/save/{id:string}", controllers.Tags.Save).Name = "SaveTag"

	// Save
	app.Get("/tag/delete/{id:string}", controllers.Tags.Delete).Name = "DeleteTag"

	//
	// =================== Units ======================
	app.Get("units.json", controllers.Units.DumpAsJSON).Name = "UnitsJSON"

	// http://localhost:8080
	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
