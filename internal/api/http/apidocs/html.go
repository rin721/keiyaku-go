package apidocs

import (
	"bytes"
	"html/template"
)

var swaggerTemplate = template.Must(template.New("swagger-ui").Parse(swaggerHTML))

type swaggerTemplateData struct {
	Title    string
	SpecPath string
}

func renderHTML(options Options) ([]byte, error) {
	var buffer bytes.Buffer
	err := swaggerTemplate.Execute(&buffer, swaggerTemplateData{
		Title:    options.Title,
		SpecPath: options.SpecPath,
	})
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

const swaggerHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html, body, #swagger-ui {
      margin: 0;
      min-height: 100%;
      background: #ffffff;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "{{ .SpecPath }}",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`
