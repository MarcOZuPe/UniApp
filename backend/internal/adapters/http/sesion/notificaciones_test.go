package sesion

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	ctrl "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	acceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/sesion"
	"uniapp/internal/infrastructure"
)

func TestNotificacionesAutorizacionYErrores(testingT *testing.T) {
	servicio := crearServicioAcceso(testingT)
	controladorAcceso := ctrl.Nuevo(servicio, false)
	bandeja := &bandejaPrueba{}

	multiplexor := http.NewServeMux()
	multiplexor.Handle("/api/notificaciones", RutasNotificaciones(bandeja, controladorAcceso.Autenticar()))
	multiplexor.Handle("/", controladorAcceso.Rutas())

	casos := []struct {
		email  string
		estado int
	}{
		{"", http.StatusUnauthorized},
		{"tutor.prueba@uni.edu.pe", http.StatusOK},
		{"escolar.prueba@example.test", http.StatusOK},
		{"apoderado.prueba@example.test", http.StatusForbidden},
		{"coord.prueba@example.test", http.StatusForbidden},
	}

	for _, caso := range casos {
		testingT.Run(caso.email, func(testingT *testing.T) {
			bandeja.consultado = ""
			var cookie *http.Cookie
			if caso.email != "" {
				cookie = iniciarSesionI05(testingT, multiplexor, caso.email)
			}

			respuesta := ejecutarPeticion(opcionesPeticionSesion{
				manejador: multiplexor,
				metodo:    http.MethodGet,
				ruta:      "/api/notificaciones?usuario_id=ajeno",
				cookie:    cookie,
			})
			if respuesta.Code != caso.estado {
				testingT.Fatalf("estado %d", respuesta.Code)
			}

			if caso.estado == http.StatusOK {
				usuario, err := servicio.Sesion.UsuarioDeSesion(context.Background(), cookie.Value)
				if err != nil || bandeja.consultado != usuario.ID {
					testingT.Fatal("identidad elegida por cliente")
				}
			} else if bandeja.consultado != "" {
				testingT.Fatal("consulta no autorizada")
			}
		})
	}

	cookie := iniciarSesionI05(testingT, multiplexor, "tutor.prueba@uni.edu.pe")
	bandeja.falla = true
	respuesta := ejecutarPeticion(opcionesPeticionSesion{
		manejador: multiplexor,
		metodo:    http.MethodGet,
		ruta:      "/api/notificaciones",
		cookie:    cookie,
	})
	if respuesta.Code != http.StatusInternalServerError || strings.Contains(respuesta.Body.String(), "privado") {
		testingT.Fatal("filtró error interno")
	}

	respuesta = ejecutarPeticion(opcionesPeticionSesion{
		manejador: multiplexor,
		metodo:    http.MethodPost,
		ruta:      "/api/notificaciones",
		cookie:    cookie,
	})
	if respuesta.Code != http.StatusMethodNotAllowed {
		testingT.Fatal("permitió mutación")
	}
}

type bandejaPrueba struct {
	consultado string
	falla      bool
}

func (bandeja *bandejaPrueba) NotificacionesDe(_ context.Context, id string) ([]modelo.Entrega, error) {
	bandeja.consultado = id
	if bandeja.falla {
		return nil, errors.New("detalle privado")
	}
	return []modelo.Entrega{{ID: 1, Asunto: "Confirmación", DestinatarioID: id}}, nil
}

func crearServicioAcceso(testingT *testing.T) *acceso.Servicio {
	testingT.Helper()

	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	servicio := acceso.NuevoServicio(acceso.Dependencias{
		Repositorio: memoriaAcceso.NuevoRepositorioMemoria(),
		Fabrica:     acceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})
	if err := acceso.Sembrar(context.Background(), servicio, acceso.DatosSiembra{
		DominioTutor: "uni.edu.pe",
		Clave:        claveI05,
	}); err != nil {
		testingT.Fatal(err)
	}
	return servicio
}
