package views

import (
	"reflect"
	"strconv"
	"time"

	"github.com/kataras/iris/core/router"
	"github.com/kataras/iris/view"
	"github.com/yoander/bolsillo/models"
)

// Ngx template
var Ngx *view.HTMLEngine

// ReverseRouter iris
var ReverseRouter *router.RoutePathReverser

// AddFuncs Add function to template engine
func AddFuncs() {
	// default template funcs are:
	//
	// - {{ urlpath "mynamedroute" "pathParameter_ifneeded" }}
	// - {{ render "header.html" }}
	// - {{ render_r "header.html" }} // partial relative path to current page
	// - {{ yield }}
	// - {{ current }}

	Ngx.AddFunc("dec", func(num int, step int) int {
		return num - step
	})

	Ngx.AddFunc("inc", func(num int, step int) int {
		return num + step
	})

	Ngx.AddFunc("fdec", func(num float64, step float64) float64 {
		return num - step
	})

	Ngx.AddFunc("finc", func(num float64, step float64) float64 {
		return num + step
	})

	Ngx.AddFunc("strtofloat", func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})

	Ngx.AddFunc("sum_prices", func(transactions models.TransactionSlice) float64 {
		sum := 0.0
		for _, t := range transactions {
			if v, err := strconv.ParseFloat(t.TotalPrice, 64); err == nil {
				sum += v
			}
		}
		return sum
	})

	Ngx.AddFunc("date", func(date time.Time, DateFormat string) string {
		return date.Format(DateFormat)
	})

	Ngx.AddFunc("now", func() time.Time {
		return time.Now()
	})

	Ngx.AddFunc("asset", func(filename string) string {
		return ReverseRouter.Path("Home") + "assets/" + filename
	})

	Ngx.AddFunc("active", func(URI string, CurrentURI string) string {
		if URI == CurrentURI {
			return " active"
		}
		return ""
	})

	Ngx.AddFunc("in", func(v interface{}, in interface{}) bool {
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

	Ngx.AddFunc("tag_exists", func(tag uint16, tags models.TagSlice) bool {
		for _, t := range tags {
			if t.ID == tag {
				return true
			}
		}
		return false
	})

	Ngx.AddFunc("join_tags", func(tags models.TagSlice, joinField string) string {
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
}
