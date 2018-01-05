package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/kataras/iris/view"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/boil"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/core/router"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/sessions"
	"github.com/yoander/bolsillo/controllers"
	"github.com/yoander/bolsillo/tpl"
)

func error500(ctx context.Context, msg string, header string) {
	ctx.StatusCode(500)
	ctx.Values().Set("header", header)
	ctx.Values().Set("message", msg)
}

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

	sess := sessions.New(sessions.Config{Cookie: "TransactionFilter"})

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

	tpl.ReverseRouter = rv
	tpl.Ngx = iris.HTML("./views", ".gohtml"). /*.Layout("layout.gohtml")*/ Reload(true) // reload templates on each request (development mode)
	tpl.AddFuncs()

	app.RegisterView(tpl.Ngx)

	// Register static content
	app.StaticWeb("/assets", "./assets")

	app.UseGlobal(func(ctx context.Context) {
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
	app.Get("/invoice/delete/{id:string}", controllers.Invoices.SoftDelete).Name = "DeleteInvoice"

	//
	// =================== Transactions ======================
	//
	// List of transactions
	app.Get("/transactions", func(ctx context.Context) {
		s := sess.Start(ctx)

		startDate := ctx.FormValue("startDate")
		if startDate != "" {
			s.Set("startDate", startDate)
		} else {
			startDate = s.GetString("startDate")
		}

		endDate := ctx.FormValue("endDate")
		if endDate != "" {
			s.Set("endDate", endDate)
		} else {
			endDate = s.GetString("endDate")
		}

		keyword := ctx.FormValue("keyword")
		if keyword != "" {
			s.Set("keyword", keyword)
		} else {
			keyword = s.GetString("keyword")
		}

		ctx.ViewData("startDate", startDate)
		ctx.ViewData("endDate", endDate)
		ctx.ViewData("keyword", keyword)
		ctx.ViewData("Title", "Transactions")
		if transactions, err := controllers.Transactions.GetFilteredTransactions(startDate, endDate, keyword); err == nil {
			ctx.ViewData("transactions", transactions)
		} else {
			error500(ctx, err.Error(), "Error listing transactions!!!")
		}

		if ctx.IsAjax() {
			ctx.ViewLayout(view.NoLayout)
			ctx.View("transactions-table.gohtml")
		} else {
			ctx.View("transactions.gohtml")
		}

	}).Name = "ListTransactions"

	// Edit one transaction
	app.Get("/transaction/edit/{id:string}", controllers.Transactions.Read).Name = "EditTransaction" // Also New

	// Clone transaction
	app.Get("/transaction/clone/{id:string}", controllers.Transactions.Clone).Name = "CloneTransaction" // Also New

	// Save a transaction
	app.Post("/transaction/save/{id:string}", controllers.Transactions.Save).Name = "SaveTransaction"

	// Delete one transaction
	app.Get("/transaction/delete/{id:string}", controllers.Transactions.SoftDelete).Name = "DeleteTransaction"

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
