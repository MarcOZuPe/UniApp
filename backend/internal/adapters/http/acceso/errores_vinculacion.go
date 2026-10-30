package acceso

import (
	"errors"
	"net/http"

	modelo "uniapp/internal/application/acceso"
)

// registrarErroresDeVinculacion registra las conversiones de errores de
// vínculo en respuestas HTTP.
//
// Patrón aplicado: Chain of Responsibility.
// MapeadorDeErrores evalúa las reglas registradas hasta encontrar una que
// reconozca el error.
func registrarErroresDeVinculacion(errorMapper *MapeadorDeErrores) {
	registrarErrorDeRolDeVinculacion(errorMapper)
	registrarErrorDeEscolarNoEncontrado(errorMapper)
	registrarErrorDeEscolarNoValido(errorMapper)
	registrarErrorDeEscolarYaVinculado(errorMapper)
	registrarErrorDeSolicitudPendiente(errorMapper)
	registrarErrorDeSinApoderado(errorMapper)
	registrarErrorDeVinculacionAjena(errorMapper)
}

func registrarErrorDeRolDeVinculacion(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return esAlguno(err, modelo.ErrNoEsApoderado, modelo.ErrNoEsEscolar)
		},
		respuestaConEstado(http.StatusForbidden, "rol_no_permitido"),
	)
}

func registrarErrorDeEscolarNoEncontrado(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrEscolarNoEncontrado)
		},
		respuestaConEstado(http.StatusNotFound, "escolar_no_encontrado"),
	)
}

func registrarErrorDeEscolarNoValido(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrEscolarNoValido)
		},
		respuestaConCampos(http.StatusUnprocessableEntity, "escolar_no_valido", "email_escolar"),
	)
}

func registrarErrorDeEscolarYaVinculado(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrYaVinculado)
		},
		respuestaConEstado(http.StatusConflict, "escolar_ya_vinculado"),
	)
}

func registrarErrorDeSolicitudPendiente(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrSolicitudPendiente)
		},
		respuestaConEstado(http.StatusConflict, "solicitud_pendiente"),
	)
}

func registrarErrorDeSinApoderado(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrSinApoderado)
		},
		respuestaConEstado(http.StatusConflict, "sin_apoderado"),
	)
}

func registrarErrorDeVinculacionAjena(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrVinculacionAjena)
		},
		respuestaConEstado(http.StatusForbidden, "no_autorizado"),
	)
}
