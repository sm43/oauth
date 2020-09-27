package design

import (
	. "goa.design/goa/v3/dsl"
	cors "goa.design/plugins/v3/cors/dsl"

	// Enables the zaplogger plugin
	_ "goa.design/plugins/v3/zaplogger"
)

var _ = API("oauth", func() {
	Title("OAuth Demo")
	Description("HTTP services for OAuth API Demo")
	Version("0.1")
	Meta("swagger:example", "false")
	Server("oauth", func() {
		Host("dev", func() {
			URI("http://localhost:8000")
		})

		Services("api")
	})

	cors.Origin("*", func() {
		cors.Headers("Content-Type", "Authorization")
		cors.Methods("GET", "POST")
	})
})
