package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo"
)

type M map[string]interface{}

func main() {
	r := echo.New()

	r.GET("/", func(ctx echo.Context) error {
		data := "Hello from /index"
		return ctx.String(http.StatusOK, data)
	})

	r.GET("/html", func(ctx echo.Context) error {
		data := "Hello from /html"
		return ctx.HTML(http.StatusOK, data)
	})

	r.GET("/index", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusTemporaryRedirect, "/")
	})

	r.GET("/json", func(ctx echo.Context) error {
		data := M{"Message": "Hello", "Counter": 2}
		return ctx.JSON(http.StatusOK, data)
	})

	// curl -X GET http://localhost:9000/page1?name=hafizarr
	r.GET("/page1", func(ctx echo.Context) error {
		name := ctx.QueryParam("name")
		data := fmt.Sprintf("Hello %s", name)

		return ctx.String(http.StatusOK, data)
	})

	// curl -X GET http://localhost:9000/page2/hafizarr
	r.GET("/page2/:name", func(ctx echo.Context) error {
		name := ctx.Param("name")
		data := fmt.Sprintf("Hello %s", name)

		return ctx.String(http.StatusOK, data)
	})

	// curl -X GET http://localhost:9000/page3/tim/need/some/sleep
	r.GET("/page3/:name/*", func(ctx echo.Context) error {
		name := ctx.Param("name")
		message := ctx.Param("*")

		data := fmt.Sprintf("Hello %s, I have message for you: %s", name, message)

		return ctx.String(http.StatusOK, data)
	})

	// curl -X POST -F name=damian -F message=angry http://localhost:9000/page4
	r.POST("/page4", func(ctx echo.Context) error {
		name := ctx.FormValue("name")
		message := ctx.FormValue("message")

		data := fmt.Sprintf(
			"Hello %s, I have message for you: %s",
			name,
			strings.Replace(message, "/", "", 1),
		)

		return ctx.String(http.StatusOK, data)
	})

	// echo.WrapHandler Untuk Routing Handler Bertipe func(http.ResponseWriter,*http.Request) atau http.HandlerFunc
	r.GET("/echoWrapHandler/index", echo.WrapHandler(http.HandlerFunc(ActionIndex)))
	r.GET("/echoWrapHandler/home", echo.WrapHandler(ActionHome))
	r.GET("/echoWrapHandler/about", ActionAbout)

	// Routing Static Assets
	// curl -X GET http://localhost:9000/static/layout.js
	r.Static("/static", "assets")

	// Parsing Request Payload
	/* Form Data
	curl --location --request POST 'http://localhost:9000/user' \
	--header 'Content-Type: application/x-www-form-urlencoded' \
	--data-urlencode 'name=hafiz' \
	--data-urlencode 'email=hafiz@gmail.com'
	*/
	/* JSON Payload
			curl --location --request POST 'http://localhost:9000/user' \
	--header 'Content-Type: application/json' \
	--data-raw '{
		"name": "hafiz",
		"email": "hafiz@gmail.com"
	}'
	*/
	/* XML Payload
		curl --location --request POST 'http://localhost:9000/user' \
	--header 'Content-Type: application/xml' \
	--data-raw '<?xml version="1.0"?>\
	<Data>\
		<Name>hafiz</Name>\
		<Email>hafiz@gmail.com</Email>\
	</Data>'
	*/
	/* Query String
	curl --location --request GET 'http://localhost:9000/user?name=hafiz&email=hafiz@gmail.com'
	*/
	r.Any("/user", func(c echo.Context) (err error) {
		u := new(User)
		if err = c.Bind(u); err != nil {
			return
		}

		return c.JSON(http.StatusOK, u)
	})

	// Request payload validation
	/*
		curl --location --request POST 'http://localhost:9000/validation/users' \
		--header 'Content-Type: application/json' \
		--data-raw '{
			"name": "hafiz",
			"email": "hafiz@gmail.com",
			"age": 121
		}'
	*/
	r.Validator = &CustomValidator{validator: validator.New()}
	// Error Handler
	/*
		r.HTTPErrorHandler = func(err error, c echo.Context) {
			report, ok := err.(*echo.HTTPError)
			if !ok {
				report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}

			c.Logger().Error(report)
			c.JSON(report.Code, report)
		}
	*/

	// Human-Readable Error
	r.HTTPErrorHandler = func(err error, c echo.Context) {
		report, ok := err.(*echo.HTTPError)
		if !ok {
			report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if castedObject, ok := err.(validator.ValidationErrors); ok {
			for _, err := range castedObject {
				switch err.Tag() {
				case "required":
					report.Message = fmt.Sprintf("%s is required",
						err.Field())
				case "email":
					report.Message = fmt.Sprintf("%s is not valid email",
						err.Field())
				case "gte":
					report.Message = fmt.Sprintf("%s value must be greater than %s",
						err.Field(), err.Param())
				case "lte":
					report.Message = fmt.Sprintf("%s value must be lower than %s",
						err.Field(), err.Param())
				}

				break
			}
		}

		c.Logger().Error(report)
		c.JSON(report.Code, report)
	}

	// Custom Error Page
	/*
		r.HTTPErrorHandler = func(err error, c echo.Context) {
			report, ok := err.(*echo.HTTPError)
			if !ok {
				report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}

			errPage := fmt.Sprintf("%d.html", report.Code)
			if err := c.File(errPage); err != nil {
				c.HTML(report.Code, "Errrrooooorrrrr")
			}
		}
	*/

	r.POST("/validation/users", func(c echo.Context) error {
		u := new(UserValidation)
		if err := c.Bind(u); err != nil {
			return err
		}
		if err := c.Validate(u); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, true)
	})

	r.Start(":9000")
}

var ActionIndex = func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("from action index"))
}

var ActionHome = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("from action home"))
	},
)

var ActionAbout = echo.WrapHandler(
	http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("from action about"))
		},
	),
)

type User struct {
	Name  string `json:"name" form:"name" query:"name"`
	Email string `json:"email" form:"email" query:"email"`
}

type UserValidation struct {
	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age"   validate:"gte=0,lte=80"`
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
