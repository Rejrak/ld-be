// design/design.go
package design

import (
	"os"

	. "goa.design/goa/v3/dsl"
)

var _ = API("be_service", func() {
	Title("Lasting Dynamics Example Service")
	Description("This service is a simple example of my backend template implementation.")
	Server("ausa_service", func() {
		Host("localhost", func() {
			URI("http://localhost:9090")
		})
	})

	HTTP(func() {
		Path("/api/v1")
	})

	Security(OAuth2, func() {
		Scope("openid")
	})

	Meta("swagger:settings", `{
		"swagger-ui-init-oauth": {
		  "clientId": "`+os.Getenv("KC_CLIENT_ID")+`",
		  "clientSecret": "`+os.Getenv("KC_CLIENT_SECRET")+`",
		  "appName": "be_service",
		  "usePkceWithAuthorizationCodeGrant": true
		}
	  }`)

})

var userService = UserService
