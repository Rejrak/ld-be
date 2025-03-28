package design

import . "goa.design/goa/v3/dsl"

var OAuth2 = OAuth2Security("oauth2", func() {
	Description("OAuth2 flow")

	PasswordFlow(
		"http://localhost:8080/realms/LastingDynamics/protocol/openid-connect/token",
		"http://localhost:8080/realms/LastingDynamics/protocol/openid-connect/token",
	)

	Scope("openid", "Access basic profile lasting_scope")
})
