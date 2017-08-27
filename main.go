package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
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

	// - standard html  | iris.HTML(...)
	// - django         | iris.Django(...)
	// - pug(jade)      | iris.Pug(...)
	// - handlebars     | iris.Handlebars(...)
	// - amber          | iris.Amber(...)

	tmpl := iris.Django("./views", ".html")
	tmpl.Reload(true) // reload templates on each request (development mode)
	// default template funcs are:
	//
	// - {{ urlpath "mynamedroute" "pathParameter_ifneeded" }}
	// - {{ render "header.html" }}
	// - {{ render_r "header.html" }} // partial relative path to current page
	// - {{ yield }}
	// - {{ current }}
	//tmpl.AddFunc("greet", func(s string) string {
	//	return "Greetings " + s + "!"
	//})
	app.RegisterView(tmpl)
	// Open handle to database like normal
	// db, err := sql.Open("mysql", "root@tcp(localhost:2483)/bolsillo?parseTime=true")

	app.Get("/", func(ctx context.Context) {
		// Eager loading
		tran, err := models.Transactions(db, Load("Tags")).All()
		if err != nil {
			fmt.Println("Error Loading Transactions", err)
		} else {
			/*for i, tx := range txt {
				s := []string{}
				for _, t := range tx.R.Tags {
					s = append(s, t.Tag)
				}
				fmt.Println(i, tx.Description.String, tx.CreatedAt, strings.Join(s, ","))
			}*/
			ctx.Gzip(true)
			ctx.ViewData("Title", "Hi Page")
			ctx.ViewData("Name", "iris") // {{.Name}} will render: iris
			ctx.ViewData("Transactions", tran)
			// ctx.ViewData("", myCcustomStruct{})
			ctx.View("layout.html")
		}
	})

	// http://localhost:8080
	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
