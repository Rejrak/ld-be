package responses

import (
	. "goa.design/goa/v3/dsl"
)

var UserListResponse = ResultType("application/vnd.user.list+json", func() {
	Description("Risposta contenente una lista di utenti paginata con metadati")
	Extend(PaginatedResponse)                                                 // Estende la risposta generica paginata
	Attribute("data", CollectionOf(UserResponse), "Lista di utenti paginata") // Specifica che `data` è una collezione di `UserResponse`
	View("default", func() {
		Attribute("total")                                                        // Totale degli elementi
		Attribute("limit")                                                        // Limite per pagina
		Attribute("offset")                                                       // Offset per pagina
		Attribute("totalPages", Int, "Numero totale di pagine")                   // Numero totale di pagine
		Attribute("currentPage", Int, "Pagina corrente")                          // Pagina corrente
		Attribute("data", CollectionOf(UserResponse), "Lista di utenti paginata") // Override per specificare che `data` è una collezione di `UserResponse`
	})
})

var UserDetailResponse = ResultType("application/vnd.user.detail+json", func() {
	Description("Risposta contenente i dettagli di un utente")
	Attributes(func() {
		Attribute("id", String, "ID dell'utente", func() {
			Example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
		})
		Attribute("firstName", String, "Nome dell'utente", func() {
			Example("John")
		})
		Attribute("lastName", String, "Cognome dell'utente", func() {
			Example("Doe")
		})
		Attribute("email", String, "Email dell'utente", func() {
			Format(FormatEmail)
			Example("johndoe@example.com")
		})
		Attribute("phoneNumber", String, "Numero di telefono", func() {
			Example("+1234567890")
		})
		Attribute("bio", String, "Biografia dell'utente", func() {
			Example("A short bio")
		})
		Attribute("nickname", String, "Nickname dell'utente", func() {
			Example("jdoe")
		})
		Attribute("dateOfBirth", String, "Data di nascita", func() {
			Format(FormatDateTime)
			Example("1990-01-01T00:00:00Z")
		})
		Attribute("sex", String, "Sesso", func() {
			Enum("M", "F")
			Example("M")
		})
		Attribute("taxCode", String, "Codice fiscale", func() {
			Example("ABC123")
		})
		Attribute("active", Boolean, "Stato dell'utente", func() {
			Default(true)
		})

	})
	View("default", func() {
		Attribute("id")
		Attribute("firstName")
		Attribute("lastName")
		Attribute("email")
		Attribute("phoneNumber")
		Attribute("bio")
		Attribute("nickname")
		Attribute("dateOfBirth")
		Attribute("sex")
		Attribute("taxCode")
		Attribute("active")
		Attribute("memberships")
		Attribute("events")
	})
})

var UserResponse = ResultType("application/vnd.user+json", func() {
	Description("Risposta contenente i dettagli di un utente")
	Extend(UserDetailResponse)
	View("default", func() {
		Attribute("id")
		Attribute("email")
		Attribute("nickname")
		Attribute("active")
	})
})
