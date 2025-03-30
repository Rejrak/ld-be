package errors

import (
	. "goa.design/goa/v3/dsl"
)

func CommonResponses() {
	Response("badRequest", StatusBadRequest)
	Response("unauthorized", StatusUnauthorized)
	Response("internalServerError", StatusInternalServerError)
	Response("notFound", StatusNotFound)
	Response("forbidden", StatusForbidden)
}

var Unauthorized = Type("Unauthorized", func() {
	Description("User not authorized to access the resource")
	Attribute("message", String, "Descrizione dell'errore", func() {
		Default("Utente già registrato a")
	})
	Required("message")
})

var InternalServerError = Type("InternalServerError", func() {
	Description("Errore nel server")
	Attribute("message", String, "Descrizione dell'errore", func() {
		Default("Errore di comunicazione con il server")
	})
	Required("message")
})

var NotFound = Type("NotFound", func() {
	Description("Dato non trovato all'interno del sistema ")
	Attribute("message", String, "Descrizione dell'errore", func() {
		Default("Dato non trovato")
	})
	Required("message")
})

var BadRequest = Type("BadRequest", func() {
	Description("Body di risposta per la richiesta non valida (400)")
	Attribute("name", String, "Nome dell'errore", func() {
		Example("invalid_range")
	})
	Attribute("id", String, "ID dell'errore", func() {
		Example("Aonp24i2")
	})
	Attribute("message", String, "Descrizione dettagliata dell'errore", func() {
		Example("ID must be greater or equal than 1 but got value -1")
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
	Required("name", "id", "message", "temporary", "timeout", "fault")
})

var Forbidden = Type("Forbidden", func() {
	Description("Cannot access the resource")
	Attribute("message", String, "Detailed description of the error", func() {
		Default("Access to the resource is forbidden")
	})
	Required("message")
})
