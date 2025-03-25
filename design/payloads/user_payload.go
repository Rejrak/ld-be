package payloads

import (
	. "goa.design/goa/v3/dsl"
)

var UserListPayload = Type("UserListPayload", func() {
	Description("Payload per ottenere una lista di utenti con supporto per paginazione e filtri")
	Extend(PaginationPayload)
	Attribute("first_name", String, "Filtro per il nome dell'utente", func() {
		Example("John")
	})
	Attribute("last_name", String, "Filtro per il cognome dell'utente", func() {
		Example("Doe")
	})
	Attribute("email", String, "Filtro per l'email dell'utente", func() {
		Format(FormatEmail)
		Example("johndoe@example.com")
	})
	Attribute("active", Boolean, "Filtro per lo stato attivo dell'utente", func() {
		Example(true)
	})
})

var CreateUserPayload = Type("CreateUserPayload", func() {
	Description("Payload per la creazione di un nuovo utente")
	Attribute("first_name", String, "Nome dell'utente", func() {
		MinLength(1)
		MaxLength(50)
		Example("John")
	})
	Attribute("last_name", String, "Cognome dell'utente", func() {
		MinLength(1)
		MaxLength(50)
		Example("Doe")
	})
	Attribute("email", String, "Email", func() {
		Format(FormatEmail)
		Example("johndoe@example.com")
	})
	Attribute("password", String, "Password dell'utente", func() {
		MinLength(8)
		MaxLength(32)
		Example("userpassword")
	})
	Attribute("phone_number", String, "Numero di telefono", func() {
		Example("+1234567890")
	})
	Attribute("bio", String, "Breve biografia dell'utente", func() {
		Example("A short bio")
	})
	Attribute("nickname", String, "Nickname dell'utente", func() {
		Example("jdoe")
	})
	Attribute("date_of_birth", String, "Data di nascita", func() {
		Format(FormatDateTime)
		Example("1990-01-01T00:00:00Z")
	})
	Attribute("sex", String, "Sesso", func() {
		Enum("M", "F")
		Example("M")
	})
	Attribute("tax_code", String, "Codice fiscale", func() {
		Example("ABC123")
	})
	Attribute("active", Boolean, "Stato dell'utente", func() {
		Default(true)
	})
	Required("first_name", "last_name", "email", "password")
})

var UpdateUserPayload = Type("UpdateUserPayload", func() {
	Description("Payload per l'aggiornamento di un utente esistente")
	Attribute("id", String, "ID dell'utente", func() {
		Format(FormatUUID)
		Example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	})
	Attribute("first_name", String, "Nome dell'utente", func() {
		MinLength(1)
		MaxLength(50)
		Example("John")
	})
	Attribute("last_name", String, "Cognome dell'utente", func() {
		MinLength(1)
		MaxLength(50)
		Example("Doe")
	})
	Attribute("email", String, "Email", func() {
		Format(FormatEmail)
		Example("johndoe@example.com")
	})
	Attribute("phone_number", String, "Numero di telefono", func() {
		Example("+1234567890")
	})
	Attribute("bio", String, "Breve biografia dell'utente", func() {
		Example("A short bio")
	})
	Attribute("nickname", String, "Nickname dell'utente", func() {
		Example("jdoe")
	})
	Attribute("date_of_birth", String, "Data di nascita", func() {
		Format(FormatDateTime)
		Example("1990-01-01T00:00:00Z")
	})
	Attribute("sex", String, "Sesso", func() {
		Enum("M", "F")
		Example("M")
	})
	Attribute("tax_code", String, "Codice fiscale", func() {
		Example("ABC123")
	})
	Attribute("active", Boolean, "Stato dell'utente", func() {
		Default(true)
	})
	Required("id")
})
