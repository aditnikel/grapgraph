package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("grapgraph", func() {
	Title("Grapgraph API")
	Description("High-performance fraud detection and financial network analysis platform built with Goa and FalkorDB.")
	Server("grapgraph", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
		})
	})
})

var _ = Service("openapi", func() {
	Description("The openapi service serves the OpenAPI specification and interactive documentation.")
	HTTP(func() {
		Path("/")
	})

	Method("index", func() {
		Description("Provides a simple landing page.")
		Payload(Empty)
		Result(String)
		HTTP(func() {
			GET("/")
			Response(StatusOK)
		})
	})

	Method("docs", func() {
		Description("Serves the interactive Swagger UI.")
		Payload(Empty)
		Result(String)
		HTTP(func() {
			GET("/docs")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
		})
	})

	Files("/openapi.json", "gen/http/openapi3.json")
})

var _ = Service("health", func() {
	Description("Health check service for monitoring service and database connectivity.")
	Method("get", func() {
		Description("Returns the health status of the API and its underlying graph database.")
		Payload(Empty)
		Result(HealthResponse)
		HTTP(func() {
			GET("/healthz")
			Response(StatusOK)
			Response(StatusServiceUnavailable)
		})
	})
})

var HealthResponse = Type("HealthResponse", func() {
	Description("Health status of the system components.")
	Attribute("ok", Boolean, "Whether the service is fully operational.", func() { Example(true) })
	Attribute("error", String, "Error message if the service is unhealthy.", func() { Example("Database connection failed") })
	Required("ok")
})
