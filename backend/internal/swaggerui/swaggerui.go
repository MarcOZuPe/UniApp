package swaggerui

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	assets "github.com/flowchartsman/swaggerui"
)

// Prefijo es la ruta pública donde se monta Swagger UI.
const Prefijo = "/swagger"

// initializerJS habilita el envío de cookies de sesión para que "Try it out"
// funcione contra los endpoints protegidos desde el mismo origen.
const initializerJS = `window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "./swagger_spec",
    dom_id: '#swagger-ui',
    deepLinking: true,
    withCredentials: true,
    requestInterceptor: function(request) {
      request.credentials = 'include';
      return request;
    },
    presets: [
      SwaggerUIBundle.presets.apis
    ]
  });
};
`

// Rutas sirve Swagger UI y el contrato OpenAPI indicado desde el propio backend.
// No usa CDN: los recursos estáticos vienen embebidos en el binario.
func Rutas(rutaEspecificacion string) (http.Handler, error) {
	if strings.TrimSpace(rutaEspecificacion) == "" {
		return nil, fmt.Errorf("falta la ruta del contrato OpenAPI")
	}

	especificacion, err := os.ReadFile(rutaEspecificacion)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el contrato OpenAPI %q: %w", rutaEspecificacion, err)
	}

	manejador := http.StripPrefix(Prefijo, conTipoDeContenido(assets.Handler(especificacion)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == Prefijo+"/swagger-initializer.js" {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			_, _ = w.Write([]byte(initializerJS))
			return
		}
		if r.URL.Path == Prefijo {
			destino := Prefijo + "/"
			if r.URL.RawQuery != "" {
				destino += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, destino, http.StatusMovedPermanently)
			return
		}
		manejador.ServeHTTP(w, r)
	}), nil
}

func conTipoDeContenido(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/swagger_spec" {
			w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		}
		siguiente.ServeHTTP(w, r)
	})
}
