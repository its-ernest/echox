package abuse_test

import (
	"github.com/its-ernest/echox/abuse"
	"github.com/labstack/echo/v5"
)

func ExampleNew() {
	e := echo.New()

	// Instant-ban on sensitive files, warn on API spam
	e.Use(abuse.New(abuse.Config{
		Threshold: 100,
		Rules: []abuse.Rule{
			// immediate ban for ANY attempt on .env
			{Path: "/.env", Score: 100},

			// heavy score only for POST requests to login (brute force protection)
			{Path: "/api/v1/login", Method: "POST", Score: 20},

			// light score for general API exploration
			{Path: "/api/v1/*", Method: "GET", Score: 2},
		},
	}))
}
