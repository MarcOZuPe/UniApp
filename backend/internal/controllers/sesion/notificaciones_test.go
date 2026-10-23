package sesion

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	ctrl "uniapp/internal/controllers/acceso"
	"uniapp/internal/infra"
	acceso "uniapp/internal/models/acceso"
	modelo "uniapp/internal/models/sesion"
)

type bandejaPrueba struct {
	consultado string
	falla      bool
}

func (b *bandejaPrueba) NotificacionesDe(_ context.Context, id string) ([]modelo.Entrega, error) {
	b.consultado = id
	if b.falla {
		return nil, errors.New("detalle privado")
	}
	return []modelo.Entrega{{ID: 1, Asunto: "Confirmación", DestinatarioID: id}}, nil
}
func TestNotificacionesAutorizacionYErrores(t *testing.T) {
	h := infra.PBKDF2Hasher{Iteraciones: 1000}
	s := acceso.NuevoServicio(acceso.Dependencias{Repositorio: acceso.NuevoRepositorioMemoria(), Fabrica: acceso.NuevaFabrica(h, "uni.edu.pe"), Hasher: h, Bus: infra.NuevoBus()})
	if err := acceso.Sembrar(context.Background(), s, acceso.OpcionesSiembra{DominioTutor: "uni.edu.pe", Clave: claveI05}); err != nil {
		t.Fatal(err)
	}
	a := ctrl.Nuevo(s, false)
	b := &bandejaPrueba{}
	mux := http.NewServeMux()
	mux.Handle("/api/notificaciones", RutasNotificaciones(b, a.Autenticar()))
	mux.Handle("/", a.Rutas())
	for _, c := range []struct {
		email  string
		estado int
	}{{"", 401}, {"tutor.prueba@uni.edu.pe", 200}, {"escolar.prueba@example.test", 200}, {"apoderado.prueba@example.test", 403}, {"coord.prueba@example.test", 403}} {
		t.Run(c.email, func(t *testing.T) {
			b.consultado = ""
			var cookie *http.Cookie
			if c.email != "" {
				cookie = iniciarSesionI05(t, mux, c.email)
			}
			req := httptest.NewRequest("GET", "/api/notificaciones?usuario_id=ajeno", nil)
			if cookie != nil {
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code != c.estado {
				t.Fatalf("estado %d", w.Code)
			}
			if c.estado == 200 {
				u, err := s.UsuarioDeSesion(context.Background(), cookie.Value)
				if err != nil || b.consultado != u.ID {
					t.Fatal("identidad elegida por cliente")
				}
			} else if b.consultado != "" {
				t.Fatal("consulta no autorizada")
			}
		})
	}
	cookie := iniciarSesionI05(t, mux, "tutor.prueba@uni.edu.pe")
	b.falla = true
	req := httptest.NewRequest("GET", "/api/notificaciones", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 500 || strings.Contains(w.Body.String(), "privado") {
		t.Fatal("filtró error interno")
	}
	req = httptest.NewRequest("POST", "/api/notificaciones", nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 405 {
		t.Fatal("permitió mutación")
	}
}
