package responses

import (
	. "goa.design/goa/v3/dsl"
)

// SuccessResponse è utilizzato per le risposte di successo
var SuccessResponse = ResultType("application/vnd.success+json", func() {
	Description("Indica che l'operazione ha avuto successo.")
	Attributes(func() {
		Attribute("message", String, "Messaggio di successo", func() {
			Example("Operation completed successfully.")
		})
	})
	View("default", func() {
		Attribute("message")
	})
})

// ErrorResponse è utilizzato per le risposte di errore
var ErrorResponse = ResultType("application/vnd.error+json", func() {
	Description("Descrizione dell'errore generico")
	Attributes(func() {
		Attribute("id", String, "ID univoco dell'errore", func() {
			Example("Aonp24i2")
		})
		Attribute("message", String, "Descrizione dell'errore", func() {
			Example("Invalid input provided.")
		})
		Attribute("temporary", Boolean, "Indica se l'errore è temporaneo", func() {
			Example(false)
		})
		Attribute("timeout", Boolean, "Indica se l'errore è dovuto a un timeout", func() {
			Example(false)
		})
		Attribute("fault", Boolean, "Indica se l'errore è dovuto a un problema del server", func() {
			Example(false)
		})
	})
	View("default", func() {
		Attribute("id")
		Attribute("message")
		Attribute("temporary")
		Attribute("timeout")
		Attribute("fault")
	})
})

var TokenResult = Type("TokenResult", func() {
	Attribute("accessToken", String, "Access Token")
	Attribute("idToken", String, "ID Token")
	Attribute("expiresIn", Int, "Expires In")
	Attribute("refreshExpiresIn", Int, "Refresh Expires In")
	Attribute("refreshToken", String, "Refresh Token")
	Attribute("tokenType", String, "Token Type")
	Attribute("notBeforePolicy", Int, "Not Before Policy")
	Attribute("sessionState", String, "Session State")
	Attribute("scope", String, "refreshToken")
	Required("accessToken", "refreshToken")
})
