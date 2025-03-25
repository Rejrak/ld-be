package responses

import (
	. "goa.design/goa/v3/dsl"
)

var PaginatedResponse = Type("PaginatedResponse", func() {
	Description("Risposta generica con supporto per la paginazione")
	Attribute("total", Int, "Numero totale di elementi", func() {
		Example(100)
	})
	Attribute("limit", Int, "Numero di elementi per pagina", func() {
		Example(10)
	})
	Attribute("offset", Int, "Offset attuale", func() {
		Example(0)
	})
	Attribute("totalPages", Int, "Numero totale di pagine", func() {
		Example(10)
	})
	Attribute("currentPage", Int, "Pagina corrente", func() {
		Example(1)
	})
	//Attribute("data", ArrayOf(Any), "Collezione di dati paginati", func() {
	//	Description("Collezione di dati paginati, di qualsiasi tipo")
	//})
})
