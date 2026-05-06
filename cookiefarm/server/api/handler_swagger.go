package api

import (
	_ "embed" // Embed the generated Swagger JSON document.
	"fmt"

	"github.com/gofiber/fiber/v3"
)

//go:embed docs/swagger.json
var swaggerJSON string

// HandleSwaggerDoc returns the generated OpenAPI v2 JSON document.
//
// @Summary Swagger document
// @Description Returns the generated Swagger/OpenAPI JSON document.
// @Tags docs
// @Success 200 {string} string "swagger json"
// @Router /swagger/doc.json [get]
func HandleSwaggerDoc(c fiber.Ctx) error {
	c.Set("Content-Type", fiber.MIMEApplicationJSONCharsetUTF8)
	return c.SendString(swaggerJSON)
}

// HandleSwaggerUI returns a lightweight Swagger UI HTML page.
//
// @Summary Swagger UI
// @Description Returns a simple Swagger UI page for browsing the API.
// @Tags docs
// @Success 200 {string} string "html"
// @Router /swagger [get]
func HandleSwaggerUI(c fiber.Ctx) error {
	specURL := "swagger/doc.json"

	html := fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>CookieFarm API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.ui = SwaggerUIBundle({
        url: %q,
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
      });
    </script>
  </body>
</html>`, specURL)

	c.Set("Content-Type", fiber.MIMETextHTMLCharsetUTF8)
	return c.SendString(html)
}
