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
			{Path: "/.env", Score: 100},
			{Path: "/api/v1/*", Score: 5},
		},
	}))
}
