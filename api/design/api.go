package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("api", func() {
	Description("The api service exposes endpoint to authenticate User against GitHub OAuth and get user details")

	Error("invalid-code", ErrorResult, "Invalid Authorization code")
	Error("invalid-token", ErrorResult, "Invalid User token")
	Error("internal-error", ErrorResult, "Internal Server Error")

	// Authenticate method defines the /oauth redirect API which authenticates user
	// against the GitHub OAuth
	// Once we get user Authorization code for user from GitHub we pass in this API
	// and it then validates that code with GitHub API and
	// if it is valid then gets user details
	Method("Authenticate", func() {
		Description("Authenticates users against GitHub OAuth")
		Payload(func() {
			Attribute("code", String, "OAuth Authorization code of User")
			Required("code")
		})
		Result(func() {
			Attribute("token", String, "JWT")
			Required("token")
		})

		HTTP(func() {
			POST("/oauth/redirect")
			Param("code")

			Response(StatusOK)
			Response("invalid-code", StatusBadRequest)
			Response("internal-error", StatusInternalServerError)
		})
	})

	// Details Method defines /details endpoint which requires user JWT to access
	// The User will get JWT after Signing and then pass JWT in this API
	// It validates the JWT and returns User details if valid
	Method("Details", func() {
		Description("Find user details")
		Security(JWTAuth)
		Payload(func() {
			Token("token", String, "JWT")
			Required("token")
		})
		Result(User)

		HTTP(func() {
			GET("/details")
			Header("token:Authorization")

			Response(StatusOK)
			Response("internal-error", StatusInternalServerError)
			Response("invalid-token", StatusUnauthorized)
		})
	})
})

// JWTAuth defines security for /details API which requires User JWT
// to access the API
var JWTAuth = JWTSecurity("jwt", func() {
	Description("Secures endpoint by requiring a valid JWT retrieved via the /oauth/redirect endpoint.")
})

// User Object
var User = Type("User", func() {
	Attribute("id", UInt, "ID is the unique id of User", func() {
		Example("id", 1)
	})
	Attribute("name", String, "Name of User", func() {
		Example("name", "Shivam")
	})
	Attribute("githubID", String, "GitHub ID", func() {
		Example("githubID", "sm43")
	})

	Required("id", "name", "githubID")
})
