package design

import (
	"be/design/errors"

	. "goa.design/goa/v3/dsl"
)

var TrainingPlan = Type("TrainingPlan", func() {
	Attribute("id", String, "TrainingPlan ID", func() {
		Format(FormatUUID)
		Example("11111111-2222-3333-4444-555555555555")
	})
	Attribute("name", String, "Name of the training plan", func() {
		Example("Upper Body Strength")
	})
	Attribute("description", String, "Description of the plan", func() {
		Example("A 4-week plan focused on upper body hypertrophy.")
	})
	Attribute("startDate", String, "Start date in ISO 8601", func() {
		Format(FormatDateTime)
		Example("2025-03-25T00:00:00Z")
	})
	Attribute("endDate", String, "End date in ISO 8601", func() {
		Format(FormatDateTime)
		Example("2025-04-25T00:00:00Z")
	})
	Attribute("userId", String, "ID of the user who owns the plan", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Required("id", "name", "startDate", "endDate", "userId")
})

var CreateTrainingPlanPayload = Type("CreateTrainingPlanPayload", func() {
	Attribute("name", String, "Name of the plan", func() {
		MinLength(1)
		Example("Upper Body Strength")
	})
	Attribute("description", String, "Description", func() {
		Example("A plan for strength.")
	})
	Attribute("startDate", String, func() {
		Format(FormatDateTime)
		Example("2025-03-25T00:00:00Z")
	})
	Attribute("endDate", String, func() {
		Format(FormatDateTime)
		Example("2025-04-25T00:00:00Z")
	})
	Attribute("userId", String, func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Required("name", "startDate", "endDate", "userId")
})

var TrainingPlanService = Service("training_plan", func() {
	Security(OAuth2, func() {
		Scope("openid")
	})
	Description("Service for managing training plans")

	HTTP(func() {
		Path("/training-plans")
	})

	Error("notFound", errors.NotFound)
	Error("internalServerError", errors.InternalServerError)
	Error("badRequest", errors.BadRequest)

	Method("create", func() {
		Payload(func() {
			AccessToken("token", String, "OAuth2 access token used to perform authorization")
			Extend(CreateTrainingPlanPayload)
		})
		Result(TrainingPlan)
		HTTP(func() {
			POST("")
			Response(StatusCreated)
		})
	})

	Method("get", func() {
		Payload(func() {
			AccessToken("token", String, "OAuth2 access token used to perform authorization")

			Field(1, "id", String, "Training plan ID", func() {
				Format(FormatUUID)
			})
			Required("id", "token")
		})
		Result(TrainingPlan)
		HTTP(func() {
			GET("/{id}")
			Response(StatusOK)
		})
	})

	Method("list", func() {
		Payload(func() {
			AccessToken("token", String, "OAuth2 access token used to perform authorization")
			Attribute("userId", String, "Filter by user ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})

			Attribute("startAfter", String, "Filter plans starting after this date (ISO 8601)", func() {
				Format(FormatDateTime)
				Example("2024-01-01T00:00:00Z")
			})

			Attribute("limit", Int, "Max number of results", func() {
				Minimum(1)
				Maximum(100)
				Default(20)
				Example(10)
			})

			Attribute("offset", Int, "Results to skip", func() {
				Minimum(0)
				Default(0)
				Example(0)
			})
		})
		Result(ArrayOf(TrainingPlan))
		HTTP(func() {
			GET("")
			Param("userId")
			Param("startAfter")
			Param("limit")
			Param("offset")
			Response(StatusOK)
		})
	})

	Method("update", func() {
		Payload(func() {
			AccessToken("token", String, "OAuth2 access token used to perform authorization")
			Attribute("id", String, func() {
				Format(FormatUUID)
			})
			Extend(CreateTrainingPlanPayload)
			Required("id")
		})
		Result(TrainingPlan)
		HTTP(func() {
			PUT("/{id}")
			Response(StatusOK)
		})
	})

	Method("delete", func() {
		Payload(func() {
			AccessToken("token", String, "OAuth2 access token used to perform authorization")
			Field(1, "id", String, func() {
				Format(FormatUUID)
			})
			Required("id")
		})
		HTTP(func() {
			DELETE("/{id}")
			Response(StatusNoContent)
		})
	})
})
