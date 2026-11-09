package sesion

import (
	"encoding/json"
	"io"
	"net/http"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/sesion"
)

const maxBytesResultado = 16 * 1024

type solicitudResultado struct {
	Resumen string `json:"resumen"`
}

// registrarResultado permite al tutor autenticado registrar el resultado.
// El autor y la fecha los fija el servidor, nunca el cliente.
func (controlador *Controlador) registrarResultado(response http.ResponseWriter, request *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(request.Context())
	if usuario == nil {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusUnauthorized, codigo: "no_autenticado", mensaje: "Debes iniciar sesion.",
		})
		return
	}

	entrada, detalle := decodificarSolicitudResultado(response, request)
	if detalle != nil {
		responderError(response, *detalle)
		return
	}

	resultado, err := controlador.repositorio.RegistrarResultado(
		request.Context(),
		request.PathValue("id"),
		usuario.ID,
		entrada.Resumen,
		controlador.ahora(),
	)
	if err != nil {
		responderErrorRepositorio(response, err)
		return
	}

	vista.EscribirJSON(response, http.StatusCreated, presentarResultado(&resultado))
}

// decodificarSolicitudResultado acepta un unico objeto JSON acotado y sin
// campos ajenos (autor, fecha, etc.).
func decodificarSolicitudResultado(
	response http.ResponseWriter,
	request *http.Request,
) (solicitudResultado, *detalleErrorHTTP) {
	var entrada solicitudResultado
	decodificador := json.NewDecoder(http.MaxBytesReader(response, request.Body, maxBytesResultado))
	decodificador.DisallowUnknownFields()

	if err := decodificador.Decode(&entrada); err != nil {
		return entrada, &detalleErrorHTTP{
			estado: http.StatusBadRequest, codigo: "solicitud_invalida",
			mensaje: "Envia un objeto JSON con el resumen del resultado.",
		}
	}
	if err := decodificador.Decode(new(any)); err != io.EOF {
		return entrada, &detalleErrorHTTP{
			estado: http.StatusBadRequest, codigo: "solicitud_invalida",
			mensaje: "Envia un unico objeto JSON.",
		}
	}
	return entrada, nil
}

func presentarResultado(resultado *modelo.Resultado) any {
	if resultado == nil {
		return nil
	}
	return map[string]any{
		"usuario_id":    resultado.UsuarioID(),
		"resumen":       resultado.Resumen(),
		"registrado_en": resultado.RegistradoEn(),
	}
}
