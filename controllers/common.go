package controllers

import (
	"database/sql"
	"fmt"

	"github.com/kataras/iris/context"
)

// DB Connection
var DB *sql.DB

var f = fmt.Sprintf

// Error500 print error message
func error500(ctx context.Context, msg string, header string) {
	ctx.StatusCode(500)
	ctx.Values().Set("header", header)
	ctx.Values().Set("message", msg)
}

// CRUD unexported
type crud interface {
	Create(ctx context.Context)
	Read(ctx context.Context)
	Update(ctx context.Context)
	Delete(ctx context.Context)
}

// Actions unexported
type Actions interface {
	crud
	List(ctx context.Context)
	Clone(ctx context.Context)
	DumpAsJSON(ctx context.Context)
}
