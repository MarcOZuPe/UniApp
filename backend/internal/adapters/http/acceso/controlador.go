package acceso

import (
	"net/http"

	modelo "uniapp/internal/application/acceso"
)

type Controlador struct {
	*controladorBase
}

func Nuevo(servicio *modelo.Servicio, cookieSegura bool) *Controlador {
	baseDelControlador := &controladorBase{
		servicio:     servicio,
		cookieSegura: cookieSegura,
		errores:      NuevoMapeadorDeErrores(),
	}
	configurarCatalogoDeErrores(baseDelControlador.errores)
	return &Controlador{controladorBase: baseDelControlador}
}

func (ensamblador *Controlador) Rutas() http.Handler {
	enrutador := http.NewServeMux()
	ensamblador.montarControladoresDeCuenta(enrutador)
	ensamblador.montarControladoresDeVinculacion(enrutador)
	ensamblador.montarControladorSalud(enrutador)
	return enrutador
}

func configurarCatalogoDeErrores(mapeador *MapeadorDeErrores) {
	registrarErroresDeCuenta(mapeador)
	registrarErroresDeVinculacion(mapeador)
}

func (ensamblador *Controlador) montarControladoresDeCuenta(mux *http.ServeMux) {
	(&RegistroControlador{controladorBase: ensamblador.controladorBase}).Montar(mux)
	(&SesionControlador{controladorBase: ensamblador.controladorBase}).Montar(mux)
	(&RecuperacionControlador{controladorBase: ensamblador.controladorBase}).Montar(mux)
	(&ConfirmacionControlador{controladorBase: ensamblador.controladorBase}).Montar(mux)
}

func (ensamblador *Controlador) montarControladoresDeVinculacion(mux *http.ServeMux) {
	(&CoordinacionControlador{controladorBase: ensamblador.controladorBase}).Montar(mux)
	(&VinculacionControlador{controladorBase: ensamblador.controladorBase}).Montar(mux)
	(&AutorizacionControlador{controladorBase: ensamblador.controladorBase}).Montar(mux)
}

func (ensamblador *Controlador) montarControladorSalud(mux *http.ServeMux) {
	(&SaludControlador{}).Montar(mux)
}
