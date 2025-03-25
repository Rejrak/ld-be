package payloads

import (
	. "goa.design/goa/v3/dsl"
)

var PaginationPayload = Type("PaginationPayload", func() {
	Description("Payload generico per paginazione e ordinamento")
	Attribute("limit", Int, "Numero massimo di elementi da restituire", func() {
		Example(25)
		Minimum(1)
		Maximum(100)
		Default(25)
	})
	Attribute("offset", Int, "Offset per la paginazione", func() {
		Example(0)
		Minimum(0)
		Default(0)
	})
	Attribute("order_by", String, "Campo per l'ordinamento", func() {
		Example("created_at")
		Default("created_at")
	})
	Attribute("order_dir", String, "Direzione dell'ordinamento (ASC o DESC)", func() {
		Enum("ASC", "DESC")
		Default("ASC")
	})
})
