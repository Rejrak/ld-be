package design

import . "goa.design/goa/v3/dsl"

var OAuth2 = OAuth2Security("oauth2", func() {
	Description("OAuth2 flow")

	AuthorizationCodeFlow(
		"https://keycloak.example.com/realms/myrealm/protocol/openid-connect/auth",
		"https://keycloak.example.com/realms/myrealm/protocol/openid-connect/token",
		"",
	)

	Scope("openid", "Access basic profile")
})
