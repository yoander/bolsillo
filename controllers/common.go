package controllers

import (
	"database/sql"
	"fmt"

	"github.com/kataras/iris/context"
	"github.com/kataras/iris/core/router"
)

// DB Connection
var DB *sql.DB

// ReverseRouter iris
var ReverseRouter *router.RoutePathReverser

var f = fmt.Sprintf

// Error500 print error message
func error500(ctx context.Context, msg string, header string) {
	ctx.StatusCode(500)
	ctx.Values().Set("header", header)
	ctx.Values().Set("message", msg)
}

type crud interface {
	Create(ctx context.Context)
	Read(ctx context.Context)
	Update(ctx context.Context)
	Delete(ctx context.Context)
}

type actions interface {
	crud
	List(ctx context.Context)
	Clone(ctx context.Context)
	Save(ctx context.Context)
	DumpAsJSON(ctx context.Context)
}
