package acceso

import (
	"context"
	"errors"
	"testing"
	"time"

	"uniapp/internal/adapters"
	"uniapp/internal/events"
)

const claveOK = "Clave-Segura-2026!"

func nuevo(t *testing.T) (*Servicio, *events.Bus, *time.Time, map[string]int) {
	t.Helper()
	h := adapters.PBKDF2Hasher{Iteraciones: 1000}
	bus := events.NuevoBus()
	cuenta := map[string]int{}
	for _, n := range []string{events.UsuarioRegistrado, events.IngresoFallido, events.IngresoExitoso, events.RecuperacionSolicitada, events.ClaveRestablecida} {
		n := n
		bus.Registrar(n, events.NuevoObservadorFunc("cuenta-"+n, func(events.Event) { cuenta[n]++ }))
	}
	s := NuevoServicio(Dependencias{
		Repositorio: NuevoRepositorioMemoria(),
		Fabrica:     NuevaFabrica(h, "uni.edu.pe"),
		Hasher:      h,
		Bus:         bus,
	})
	ahora := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	s.FijarReloj(func() time.Time { return ahora })
	return s, bus, &ahora, cuenta
}

func TestFabricaPorRol(t *testing.T) {
	f := NuevaFabrica(adapters.PBKDF2Hasher{Iteraciones: 1000}, "uni.edu.pe")
	if _, err := f.CrearPublico(DatosRegistro{Email: "Ana@Example.com", NombreUso: "Ana", Clave: "una-clave-larga", Rol: RolEscolar}); err != nil {
		t.Fatal(err)
	}
	var ev ErrValidacion
	_, err := f.CrearPublico(DatosRegistro{Email: "t@gmail.com", NombreUso: "Tito", Clave: claveOK, Rol: RolTutor})
	if !errors.As(err, &ev) || ev.Campos["email"] == "" {
		t.Fatalf("tutor sin correo institucional debe fallar: %v", err)
	}
	if _, err := f.CrearPublico(DatosRegistro{Email: "c@example.com", NombreUso: "Coord", Clave: claveOK, Rol: RolCoordinador}); !errors.Is(err, ErrRolNoPermitido) {
		t.Fatalf("coordinador no se registra desde la web: %v", err)
	}
}

func TestPoliticasStrategy(t *testing.T) {
	if len((PoliticaBasica{Min: 10}).Validar("larga-pero-simple", "a@b.c")) != 0 {
		t.Fatal("básica debe aceptar")
	}
	if len((PoliticaEstricta{Min: 12}).Validar("larga-pero-simple", "a@b.c")) == 0 {
		t.Fatal("estricta debe exigir clases de caracteres")
	}
}

func TestRegistroIngresoYSesion(t *testing.T) {
	s, _, _, ev := nuevo(t)
	ctx := context.Background()
	if _, err := s.Registrar(ctx, DatosRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Registrar(ctx, DatosRegistro{Email: "E@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar}); !errors.Is(err, ErrEmailEnUso) {
		t.Fatalf("correo duplicado (sin distinguir mayúsculas): %v", err)
	}
	tok, u, err := s.Ingresar(ctx, "e@example.com", "clave-larga-ok")
	if err != nil || u.Rol != RolEscolar {
		t.Fatal(err)
	}
	if got, err := s.UsuarioDeSesion(ctx, tok); err != nil || got.ID != u.ID {
		t.Fatal("sesión debe resolver al usuario")
	}
	_ = s.Salir(ctx, tok)
	if _, err := s.UsuarioDeSesion(ctx, tok); err == nil {
		t.Fatal("sesión cerrada no debe resolver")
	}
	if ev[events.UsuarioRegistrado] != 1 || ev[events.IngresoExitoso] != 1 {
		t.Fatalf("observers: %v", ev)
	}
}

func TestSesionVence(t *testing.T) {
	s, _, ahora, _ := nuevo(t)
	ctx := context.Background()
	_, _ = s.Registrar(ctx, DatosRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar})
	tok, _, _ := s.Ingresar(ctx, "e@example.com", "clave-larga-ok")
	*ahora = ahora.Add(13 * time.Hour)
	if _, err := s.UsuarioDeSesion(ctx, tok); err == nil {
		t.Fatal("sesión vencida")
	}
}

func TestBloqueoPorIntentos(t *testing.T) {
	s, _, ahora, _ := nuevo(t)
	ctx := context.Background()
	_, _ = s.Registrar(ctx, DatosRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar})
	for i := 0; i < 5; i++ {
		if _, _, err := s.Ingresar(ctx, "e@example.com", "mala"); !errors.Is(err, ErrCredenciales) {
			t.Fatalf("intento %d: %v", i, err)
		}
	}
	if _, _, err := s.Ingresar(ctx, "e@example.com", "clave-larga-ok"); !errors.Is(err, ErrDemasiadosIntentos) {
		t.Fatalf("debe bloquear incluso con clave correcta: %v", err)
	}
	*ahora = ahora.Add(16 * time.Minute)
	if _, _, err := s.Ingresar(ctx, "e@example.com", "clave-larga-ok"); err != nil {
		t.Fatalf("tras la ventana debe permitir: %v", err)
	}
}

func TestObtenerPerfilFabricaSimple(t *testing.T) {
	for _, r := range []Rol{RolEscolar, RolApoderado, RolTutor, RolCoordinador} {
		p, err := ObtenerPerfil(r, "uni.edu.pe")
		if err != nil || p.Rol() != r {
			t.Fatalf("rol %s: %v", r, err)
		}
	}
	if p, _ := ObtenerPerfil(RolCoordinador, "x"); p.EsPublico() {
		t.Fatal("coordinación no es de registro público")
	}
	if _, err := ObtenerPerfil(Rol("intruso"), "x"); err == nil {
		t.Fatal("rol desconocido debe fallar")
	}
}

func TestValidadorClaveCambiaDeEstrategia(t *testing.T) {
	v := NuevoValidadorClave(PoliticaBasica{Min: 10})
	if len(v.Validar("larga-pero-simple", "a@b.c")) != 0 {
		t.Fatal("básica debe aceptar")
	}
	v.FijarPolitica(PoliticaEstricta{Min: 12})
	if len(v.Validar("larga-pero-simple", "a@b.c")) == 0 {
		t.Fatal("tras FijarPolitica debe aplicar la estricta")
	}
}
