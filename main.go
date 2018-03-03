package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	"github.com/yoander/bolsillo/views"
)

func error500(ctx context.Context, msg string, header string) {
	ctx.StatusCode(500)
	ctx.Values().Set("header", header)
	ctx.Values().Set("message", msg)
}

func main() {
	//p := fmt.Println
	f := fmt.Sprintf
	app := iris.New() // defaults to these
	rv := router.NewRoutePathReverser(app)
	controllers.ReverseRouter = rv
	// Open handle to database like normal
	db, err := sql.Open("mysql", "root:root@tcp(localhost:2483)/bolsillo?parseTime=true&loc=Local")
	if err != nil {
		log.Fatal(err)
	}
	controllers.DB = db
	boil.SetLocation(time.Local)
	boil.DebugMode = true
	// Optionally set the writer as well. Defaults to os.Stdout
	//fh, err := os.Create("debug.txt")
	//boil.DebugWriter = fh

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

	filterSess := sessions.New(sessions.Config{Cookie: "filter"})
	parseFilterParams := func(ctx context.Context, sessionPrefix string) (time.Time, time.Time, string, error) {
		now := time.Now().Local()
		s := filterSess.Start(ctx)
		var startDate, endDate time.Time
		var err error
		strStartDate := ctx.FormValue("startDate")
		if strStartDate != "" {
			s.Set(sessionPrefix+"_start_date", strStartDate)
		} else {
			strStartDate = s.GetString(sessionPrefix + "_start_date")
		}

		strEndDate := ctx.FormValue("endDate")
		if strEndDate != "" {
			s.Set(sessionPrefix+"_end_date", strEndDate)
		} else {
			strEndDate = s.GetString(sessionPrefix + "_end_date")
		}

		keyword := ctx.FormValue("keyword")
		if ctx.IsAjax() {
			s.Set(sessionPrefix+"_keyword", keyword)
		} else {
			keyword = s.GetString(sessionPrefix + "_keyword")
		}

		ctx.ViewData("startDate", strStartDate)
		ctx.ViewData("endDate", strEndDate)
		ctx.ViewData("keyword", keyword)

		if strStartDate == "" {
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		} else if startDate, err = time.Parse("02.01.2006", strStartDate); err != nil {
			return now, now, "", err
		}

		if strEndDate == "" {
			endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		} else if endDate, err = time.Parse("02.01.2006", strEndDate); err != nil {
			return startDate, now, "", err
		}

		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 99999999999, time.UTC)

		return startDate, endDate, keyword, err
	}

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

	app.Use(recover.New())
	app.Use(logger.New())

	views.ReverseRouter = rv
	views.Ngx = iris.HTML("./views", ".gohtml"). /*.Layout("layout.gohtml")*/ Reload(true) // reload templates on each request (development mode)
	views.AddFuncs()

	app.RegisterView(views.Ngx)

	// Register static content
	app.StaticWeb("/assets", "./web")

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
	app.Get("/invoices", func(ctx context.Context) {
		if startDate, endDate, keyword, err := parseFilterParams(ctx, "invoices"); err == nil {
			if invoicePrices, err := controllers.
				Invoices.
				List(startDate, endDate, keyword); err == nil {
				ctx.ViewData("Invoices", invoicePrices)
			} else {
				error500(ctx, err.Error(), "Error listing invoices!!!")
			}
		} else {
			error500(ctx, err.Error(), "Error listing invoices!!!")
		}
		ctx.ViewData("Title", "Invoices")
		if ctx.IsAjax() {
			ctx.ViewLayout(view.NoLayout)
			ctx.View("invoices-table.gohtml")
		} else {
			ctx.View("invoices.gohtml")
		}
	}).Name = "ListInvoices"

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
	// List
	app.Get("/transactions", func(ctx context.Context) {
		if startDate, endDate, keyword, err := parseFilterParams(ctx, "transactions"); err == nil {
			if transactions, expenses, incomes, profit, err := controllers.
				Transactions.
				List(startDate, endDate, keyword); err == nil {
				ctx.ViewData("transactions", transactions)
				ctx.ViewData("expenses", expenses)
				ctx.ViewData("incomes", incomes)
				ctx.ViewData("profit", fmt.Sprintf("%.2f", profit))
			} else {
				error500(ctx, err.Error(), "Error listing transactions!!!")
			}
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
	app.Post("/transaction/save/{id:string}", func(ctx context.Context) {
		id := ctx.Params().Get("id")
		ID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", id))
		}

		isExpensive, err := strconv.ParseUint(ctx.PostValue("IsExpensive"), 10, 8)
		if err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", id))
		}

		date, err := time.Parse("02.01.2006", ctx.PostValue("Date"))
		if err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", id))
		}
		// Set correct date time
		date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)

		invoiceID, err := strconv.ParseUint(ctx.PostValue("Invoice"), 10, 8)
		if err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", id))
		}

		unitID, err := strconv.ParseUint(ctx.PostValue("Unit"), 10, 8)
		if err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", id))
		}
		//tags :=
		err = controllers.Transactions.Save(uint(ID),
			uint(invoiceID),
			uint8(unitID),
			1,
			0,
			ctx.PostValue("Type"),
			ctx.PostValue("Description"),
			ctx.PostValue("Note"),
			ctx.PostValue("Quantity"),
			ctx.PostValue("UnitPrice"),
			ctx.PostValue("TotalPrice"),
			ctx.PostValue("Status"),
			date,
			int8(isExpensive),
			strings.Split(ctx.PostValue("tags"), ","))

		//fmt.Println("tags", strings.Split(ctx.PostValue("tags"), ","))
		if err != nil {
			error500(ctx, err.Error(), f("Error saving transactions %d", id))
		}

		ctx.Redirect(rv.Path("ListTransactions"))

	}).Name = "SaveTransaction"

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

	// http://localhost:8081
	app.Run(iris.Addr(":8081"), iris.WithoutServerError(iris.ErrServerClosed))
}
