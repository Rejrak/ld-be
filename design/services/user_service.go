package services

import (
	"be/design/errors"

	. "goa.design/goa/v3/dsl"
)

var User = Type("User", func() {
	Attribute("id", String, "Unique ID of the user", func() {
		Example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
		Format(FormatUUID)
	})
	Attribute("kcId", String, "Keycloak ID", func() {
		Example("550e8400-e29b-41d4-a716-446655440000")
		Format(FormatUUID)
	})
	Attribute("firstName", String, "First name of the user", func() {
		Example("John")
	})
	Attribute("lastName", String, "Last name of the user", func() {
		Example("Doe")
	})
	Attribute("nickname", String, "Nickname", func() {
		Example("JD")
	})
	Attribute("admin", Boolean, "Is the user an admin?", func() {
		Default(false)
		Example(false)
	})
	Required("id", "kcId", "firstName", "lastName")
})

var CreateUserPayload = Type("CreateUserPayload", func() {
	Attribute("kcId", String, "Keycloak ID", func() {
		Example("550e8400-e29b-41d4-a716-446655440000")
		Format(FormatUUID)
	})
	Attribute("firstName", String, "First name", func() {
		Example("John")
	})
	Attribute("lastName", String, "Last name", func() {
		Example("Doe")
	})
	Attribute("nickname", String, "Nickname", func() {
		Example("JD")
	})
	Attribute("admin", Boolean, "Is admin?", func() {
		Default(false)
		Example(false)
	})
	Required("kcId", "firstName", "lastName")
})

var UserService = Service("user", func() {
	Description("User service for managing users")

	HTTP(func() {
		Path("/user")
	})

	Error("unauthorized", errors.Unauthorized, "Autenticazione fallita")
	Error("internalServerError", errors.InternalServerError, "Errore nel server")
	Error("notFound", errors.NotFound, "Dato non trovato all'interno del sistema")
	Error("badRequest", errors.BadRequest, "Parametri non validi")

	Method("create", func() {
		Description("Create a new user")
		Payload(CreateUserPayload)
		Result(User)
		HTTP(func() {
			POST("/")
			Response(StatusCreated)
		})
	})

	Method("get", func() {
		Description("Get a user by ID")
		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
				Format(FormatUUID)
			})
			Required("id")
		})
		Result(User)
		HTTP(func() {
			GET("/{id}")
			Response(StatusOK)
			errors.CommonResponses()
		})
	})

	Method("list", func() {
		Description("List all users with pagination")
		Payload(func() {
			Attribute("limit", Int, "Number of users to return per page", func() {
				Example(10)
				Minimum(1)
				Maximum(100)
				Default(10)
			})
			Attribute("offset", Int, "Number of users to skip", func() {
				Example(0)
				Minimum(0)
				Default(0)
			})
		})
		Result(ArrayOf(User))
		HTTP(func() {
			GET("")
			Param("limit")
			Param("offset")
			Response(StatusOK)
		})
	})

	Method("update", func() {
		Description("Update a user")
		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
				Format(FormatUUID)
			})
			Attribute("kcId", String, "Keycloak ID", func() {
				Example("550e8400-e29b-41d4-a716-446655440000")
				Format(FormatUUID)
			})
			Attribute("firstName", String, "First name", func() {
				Example("John")
			})
			Attribute("lastName", String, "Last name", func() {
				Example("Doe")
			})
			Attribute("nickname", String, "Nickname", func() {
				Example("JD")
			})
			Attribute("admin", Boolean, "Is admin?", func() {
				Default(false)
				Example(false)
			})
			Required("id", "kcId", "firstName", "lastName")
		})
		Result(User)
		HTTP(func() {
			PUT("/{id}")
			Response(StatusOK)
			errors.CommonResponses()
		})
	})

	Method("delete", func() {
		Description("Delete a user")
		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
				Format(FormatUUID)
			})
			Required("id")
		})
		HTTP(func() {
			DELETE("/{id}")
			Response(StatusNoContent)
			errors.CommonResponses()
		})
	})
})