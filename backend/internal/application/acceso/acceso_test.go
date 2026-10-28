package acceso

import (
	"context"
	"errors"
	"testing"
	"time"
	memoria "uniapp/internal/adapters/repositories/memory/acceso"
)

const claveOK = "Clave-Segura-2026!"

func nuevo(t *testing.T) (*Servicio, *busDePrueba, *time.Time, map[string]int) {
	t.Helper()
	h := hasherDePrueba{}
	bus := nuevoBusDePrueba()
	cuenta := map[string]int{}
	for _, n := range []string{UsuarioRegistrado, IngresoFallido, IngresoExitoso, RecuperacionSolicitada, ClaveRestablecida, ConfirmacionSolicitada, CorreoConfirmado} {
		n := n
		bus.Registrar(n, nuevoObservadorDePrueba("cuenta-"+n, func(Event) { cuenta[n]++ }))
	}
	s := NuevoServicio(Dependencias{
		Repositorio: memoria.NuevoRepositorioMemoria(),
		Fabrica:     NuevaFabrica(h, "uni.edu.pe"),
		Hasher:      h,
		Bus:         bus,
	})
	ahora := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	s.FijarReloj(func() time.Time { return ahora })
	return s, bus, &ahora, cuenta
}

func registrarConfirmado(t *testing.T, s *Servicio, bus *busDePrueba, datos DatosRegistro) *Usuario {
	t.Helper()
	var token string
	bus.Registrar(ConfirmacionSolicitada, nuevoObservadorDePrueba("captura-confirmacion-"+datos.Email, func(e Event) {
		token, _ = e.Payload["token"].(string)
	}))
	usuario, err := s.Registrar(context.Background(), datos)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ConfirmarCorreo(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	return usuario
}

func TestFabricaPorRol(t *testing.T) {
	f := NuevaFabrica(hasherDePrueba{}, "uni.edu.pe,uni.pe")
	publico, err := f.CrearPublico(DatosRegistro{Email: "Ana@Example.com", NombreUso: "Ana", Clave: "una-clave-larga", Rol: RolEscolar})
	if err != nil {
		t.Fatal(err)
	}
	if publico.EmailConfirmado {
		t.Fatal("el registro público debe nacer sin confirmar")
	}
	interno, err := f.CrearInterno(DatosRegistro{Email: "coord@example.com", NombreUso: "Coord", Clave: claveOK, Rol: RolCoordinador})
	if err != nil || !interno.EmailConfirmado {
		t.Fatalf("las cuentas internas nacen confirmadas: %v", err)
	}
	for _, correo := range []string{"t@uni.edu.pe", "t@uni.pe"} {
		if _, err := f.CrearPublico(DatosRegistro{Email: correo, NombreUso: "Tito", Clave: claveOK, Rol: RolTutor}); err != nil {
			t.Fatalf("tutor con %s debe registrarse: %v", correo, err)
		}
	}
	var ev ErrValidacion
	for _, correo := range []string{"t@gmail.com", "t@uni.edu.pe.evil.com"} {
		_, err := f.CrearPublico(DatosRegistro{Email: correo, NombreUso: "Tito", Clave: claveOK, Rol: RolTutor})
		if !errors.As(err, &ev) || ev.Campos["email"] == "" {
			t.Fatalf("tutor con %s no debe registrarse: %v", correo, err)
		}
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
	s, bus, _, ev := nuevo(t)
	ctx := context.Background()
	var token string
	bus.Registrar(ConfirmacionSolicitada, nuevoObservadorDePrueba("captura-confirmacion", func(e Event) {
		token, _ = e.Payload["token"].(string)
	}))
	if _, err := s.Registrar(ctx, DatosRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Registrar(ctx, DatosRegistro{Email: "E@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar}); !errors.Is(err, ErrEmailEnUso) {
		t.Fatalf("correo duplicado (sin distinguir mayúsculas): %v", err)
	}
	if _, _, err := s.Ingresar(ctx, "e@example.com", "clave-larga-ok"); !errors.Is(err, ErrCredenciales) {
		t.Fatalf("un registro sin confirmar debe ser invisible (mensaje genérico): %v", err)
	}
	if err := s.ConfirmarCorreo(ctx, token); err != nil {
		t.Fatal(err)
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
	if ev[UsuarioRegistrado] != 1 || ev[IngresoExitoso] != 1 {
		t.Fatalf("observers: %v", ev)
	}
}

func TestConfirmacionDeCorreo(t *testing.T) {
	s, bus, ahora, ev := nuevo(t)
	ctx := context.Background()
	var tokens []string
	bus.Registrar(ConfirmacionSolicitada, nuevoObservadorDePrueba("captura-confirmacion", func(e Event) {
		enlace, _ := e.Payload["token"].(string)
		tokens = append(tokens, enlace)
	}))
	datos := DatosRegistro{Email: "c@example.com", NombreUso: "Ceci", Clave: "clave-larga-ok", Rol: RolEscolar}
	registrado, err := s.Registrar(ctx, datos)
	if err != nil {
		t.Fatal(err)
	}
	if registrado.EmailConfirmado {
		t.Fatal("el registro público nace sin confirmar")
	}
	if len(tokens) != 1 || tokens[0] == "" {
		t.Fatalf("debía emitirse un enlace de confirmación: %v", tokens)
	}
	if _, _, err := s.Ingresar(ctx, "c@example.com", "mala"); !errors.Is(err, ErrCredenciales) {
		t.Fatalf("con clave mala el mensaje debe ser genérico: %v", err)
	}
	if _, _, err := s.Ingresar(ctx, "c@example.com", "clave-larga-ok"); !errors.Is(err, ErrCredenciales) {
		t.Fatalf("sin confirmar debe responder igual que una cuenta inexistente: %v", err)
	}
	if err := s.ReenviarConfirmacion(ctx, "c@example.com"); err != nil || len(tokens) != 2 {
		t.Fatalf("reenviar debía emitir otro enlace: %v (%d)", err, len(tokens))
	}
	if err := s.ReenviarConfirmacion(ctx, "nadie@example.com"); err != nil || len(tokens) != 2 {
		t.Fatalf("correo inexistente: sin error y sin enlace: %v (%d)", err, len(tokens))
	}
	if err := s.ConfirmarCorreo(ctx, tokens[0]); err != nil {
		t.Fatalf("el enlace original sigue vigente tras reenviar: %v", err)
	}
	if err := s.ConfirmarCorreo(ctx, tokens[0]); !errors.Is(err, ErrTokenInvalido) {
		t.Fatal("el enlace es de un solo uso")
	}
	if err := s.ReenviarConfirmacion(ctx, "c@example.com"); err != nil || len(tokens) != 2 {
		t.Fatalf("una cuenta confirmada no recibe otro enlace: %v (%d)", err, len(tokens))
	}
	if _, _, err := s.Ingresar(ctx, "c@example.com", "clave-larga-ok"); err != nil {
		t.Fatalf("tras confirmar debe ingresar: %v", err)
	}
	if ev[ConfirmacionSolicitada] != 2 || ev[CorreoConfirmado] != 1 {
		t.Fatalf("observers: %v", ev)
	}

	if _, err := s.Registrar(ctx, DatosRegistro{Email: "d@example.com", NombreUso: "Dani", Clave: "clave-larga-ok", Rol: RolEscolar}); err != nil {
		t.Fatal(err)
	}
	enlaceDeD := tokens[len(tokens)-1]
	*ahora = ahora.Add(31 * time.Minute)
	if err := s.ConfirmarCorreo(ctx, enlaceDeD); !errors.Is(err, ErrTokenInvalido) {
		t.Fatal("el enlace de confirmación vence a los 30 minutos")
	}
	if _, _, err := s.Ingresar(ctx, "d@example.com", "clave-larga-ok"); !errors.Is(err, ErrCredenciales) {
		t.Fatalf("un registro vencido sigue siendo invisible: %v", err)
	}
	liberados, err := s.LimpiarRegistrosVencidos(ctx)
	if err != nil || liberados != 1 {
		t.Fatalf("debía liberarse un correo vencido: %d (%v)", liberados, err)
	}
	if _, err := s.Registrar(ctx, DatosRegistro{Email: "d@example.com", NombreUso: "Dani", Clave: "clave-larga-ok", Rol: RolEscolar}); err != nil {
		t.Fatalf("tras vencer el plazo, el correo vuelve a estar libre: %v", err)
	}
}

func TestRecuperacionNoEnviaAQuienNoConfirmo(t *testing.T) {
	s, bus, _, _ := nuevo(t)
	ctx := context.Background()
	var tokens []string
	bus.Registrar(RecuperacionSolicitada, nuevoObservadorDePrueba("captura-recuperacion", func(e Event) {
		token, _ := e.Payload["token"].(string)
		tokens = append(tokens, token)
	}))
	if _, err := s.Registrar(ctx, DatosRegistro{Email: "sin.confirmar@example.com", NombreUso: "Sin", Clave: "clave-larga-ok", Rol: RolEscolar}); err != nil {
		t.Fatal(err)
	}
	if err := s.SolicitarRecuperacion(ctx, "sin.confirmar@example.com"); err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 0 {
		t.Fatalf("un registro sin confirmar no debe recibir enlace de recuperación: %v", tokens)
	}
}

func TestSesionVence(t *testing.T) {
	s, bus, ahora, _ := nuevo(t)
	ctx := context.Background()
	registrarConfirmado(t, s, bus, DatosRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar})
	tok, _, _ := s.Ingresar(ctx, "e@example.com", "clave-larga-ok")
	*ahora = ahora.Add(13 * time.Hour)
	if _, err := s.UsuarioDeSesion(ctx, tok); err == nil {
		t.Fatal("sesión vencida")
	}
}

func TestBloqueoPorIntentos(t *testing.T) {
	s, bus, ahora, _ := nuevo(t)
	ctx := context.Background()
	registrarConfirmado(t, s, bus, DatosRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar})
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

func TestRecuperacion(t *testing.T) {
	s, bus, ahora, _ := nuevo(t)
	ctx := context.Background()
	var token string
	bus.Registrar(RecuperacionSolicitada, nuevoObservadorDePrueba("captura-token", func(e Event) { token, _ = e.Payload["token"].(string) }))
	registrarConfirmado(t, s, bus, DatosRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar})
	sesion, _, _ := s.Ingresar(ctx, "e@example.com", "clave-larga-ok")

	if err := s.SolicitarRecuperacion(ctx, "nadie@example.com"); err != nil || token != "" {
		t.Fatal("correo inexistente: sin error y sin token")
	}
	_ = s.SolicitarRecuperacion(ctx, "e@example.com")
	if token == "" {
		t.Fatal("debía emitirse token")
	}
	var ev ErrValidacion
	if err := s.RestablecerClave(ctx, token, "corta"); !errors.As(err, &ev) {
		t.Fatalf("clave débil: %v", err)
	}
	if err := s.RestablecerClave(ctx, token, "nueva-clave-larga"); err != nil {
		t.Fatalf("la clave débil no debía gastar el token: %v", err)
	}
	if err := s.RestablecerClave(ctx, token, "otra-clave-larga"); !errors.Is(err, ErrTokenInvalido) {
		t.Fatal("el token es de un solo uso")
	}
	if _, err := s.UsuarioDeSesion(ctx, sesion); err == nil {
		t.Fatal("restablecer cierra las sesiones previas")
	}
	if _, _, err := s.Ingresar(ctx, "e@example.com", "nueva-clave-larga"); err != nil {
		t.Fatal(err)
	}
	_ = s.SolicitarRecuperacion(ctx, "e@example.com")
	*ahora = ahora.Add(2 * time.Hour)
	if err := s.RestablecerClave(ctx, token, "tercera-clave-larga"); !errors.Is(err, ErrTokenInvalido) {
		t.Fatal("token vencido")
	}
}

func TestObtenerPerfilFabricaSimple(t *testing.T) {
	for _, r := range []Rol{RolEscolar, RolApoderado, RolTutor, RolCoordinador} {
		p, err := ObtenerPerfil(r, "uni.edu.pe,uni.pe")
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
