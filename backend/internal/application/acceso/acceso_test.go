package acceso

import (
	"context"
	"errors"
	"testing"
	"time"
	memoria "uniapp/internal/adapters/repositories/memory/acceso"
	"uniapp/internal/infrastructure/lockout"
)

const claveOK = "Clave-Segura-2026!"

func TestFabricaPorRol(testingT *testing.T) {
	fabrica := NuevaFabrica(hasherDePrueba{}, "uni.edu.pe,uni.pe")
	publico, err := fabrica.CrearPublico(SolicitudRegistro{Email: "Ana@Example.com", NombreUso: "Ana", Clave: "una-clave-larga", Rol: RolEscolar})
	if err != nil {
		testingT.Fatal(err)
	}
	if publico.EmailConfirmado {
		testingT.Fatal("el registro público debe nacer sin confirmar")
	}
	interno, err := fabrica.CrearInterno(SolicitudRegistro{Email: "coord@example.com", NombreUso: "Coord", Clave: claveOK, Rol: RolCoordinador})
	if err != nil || !interno.EmailConfirmado {
		testingT.Fatalf("las cuentas internas nacen confirmadas: %v", err)
	}
	for _, correo := range []string{"testingT@uni.edu.pe", "testingT@uni.pe"} {
		if _, err := fabrica.CrearPublico(SolicitudRegistro{Email: correo, NombreUso: "Tito", Clave: claveOK, Rol: RolTutor}); err != nil {
			testingT.Fatalf("tutor con %s debe registrarse: %v", correo, err)
		}
	}
	var ev ErrValidacion
	for _, correo := range []string{"testingT@gmail.com", "testingT@uni.edu.pe.evil.com"} {
		_, err := fabrica.CrearPublico(SolicitudRegistro{Email: correo, NombreUso: "Tito", Clave: claveOK, Rol: RolTutor})
		if !errors.As(err, &ev) || ev.Campos["email"] == "" {
			testingT.Fatalf("tutor con %s no debe registrarse: %v", correo, err)
		}
	}
	if _, err := fabrica.CrearPublico(SolicitudRegistro{Email: "c@example.com", NombreUso: "Coord", Clave: claveOK, Rol: RolCoordinador}); !errors.Is(err, ErrRolNoPermitido) {
		testingT.Fatalf("coordinador no se registra desde la web: %v", err)
	}
}

func TestPoliticasStrategy(testingT *testing.T) {
	if len((PoliticaBasica{Min: 10}).Validar("larga-pero-simple", "a@b.c")) != 0 {
		testingT.Fatal("básica debe aceptar")
	}
	if len((PoliticaEstricta{Min: 12}).Validar("larga-pero-simple", "a@b.c")) == 0 {
		testingT.Fatal("estricta debe exigir clases de caracteres")
	}
}

func TestRegistroIngresoYSesion(testingT *testing.T) {
	servicio, bus, _, ev := nuevo(testingT)
	ctx := context.Background()
	var token string
	bus.Registrar(ConfirmacionSolicitada, nuevoObservadorDePrueba("captura-confirmacion", func(e Event) {
		token, _ = e.Payload["token"].(string)
	}))
	if _, err := servicio.Registro.Registrar(ctx, SolicitudRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar}); err != nil {
		testingT.Fatal(err)
	}
	if _, err := servicio.Registro.Registrar(ctx, SolicitudRegistro{Email: "E@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar}); !errors.Is(err, ErrEmailEnUso) {
		testingT.Fatalf("correo duplicado (sin distinguir mayúsculas): %v", err)
	}
	if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("e@example.com", "clave-larga-ok")); !errors.Is(err, ErrCredenciales) {
		testingT.Fatalf("un registro sin confirmar debe ser invisible (mensaje genérico): %v", err)
	}
	if err := servicio.Confirmacion.ConfirmarCorreo(ctx, token); err != nil {
		testingT.Fatal(err)
	}
	tok, u, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("e@example.com", "clave-larga-ok"))
	if err != nil || u.Rol != RolEscolar {
		testingT.Fatal(err)
	}
	if got, err := servicio.Sesion.UsuarioDeSesion(ctx, tok); err != nil || got.ID != u.ID {
		testingT.Fatal("sesión debe resolver al usuario")
	}
	_ = servicio.Sesion.Salir(ctx, tok)
	if _, err := servicio.Sesion.UsuarioDeSesion(ctx, tok); err == nil {
		testingT.Fatal("sesión cerrada no debe resolver")
	}
	if ev[UsuarioRegistrado] != 1 || ev[IngresoExitoso] != 1 {
		testingT.Fatalf("observers: %v", ev)
	}
}

func TestConfirmacionDeCorreo(testingT *testing.T) {
	servicio, bus, ahora, ev := nuevo(testingT)
	ctx := context.Background()
	var tokens []string
	bus.Registrar(ConfirmacionSolicitada, nuevoObservadorDePrueba("captura-confirmacion", func(e Event) {
		enlace, _ := e.Payload["token"].(string)
		tokens = append(tokens, enlace)
	}))
	datos := SolicitudRegistro{Email: "c@example.com", NombreUso: "Ceci", Clave: "clave-larga-ok", Rol: RolEscolar}
	registrado, err := servicio.Registro.Registrar(ctx, datos)
	if err != nil {
		testingT.Fatal(err)
	}
	if registrado.EmailConfirmado {
		testingT.Fatal("el registro público nace sin confirmar")
	}
	if len(tokens) != 1 || tokens[0] == "" {
		testingT.Fatalf("debía emitirse un enlace de confirmación: %v", tokens)
	}
	if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("c@example.com", "mala")); !errors.Is(err, ErrCredenciales) {
		testingT.Fatalf("con clave mala el mensaje debe ser genérico: %v", err)
	}
	if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("c@example.com", "clave-larga-ok")); !errors.Is(err, ErrCredenciales) {
		testingT.Fatalf("sin confirmar debe responder igual que una cuenta inexistente: %v", err)
	}
	if err := servicio.Confirmacion.ReenviarConfirmacion(ctx, "c@example.com"); err != nil || len(tokens) != 2 {
		testingT.Fatalf("reenviar debía emitir otro enlace: %v (%d)", err, len(tokens))
	}
	if err := servicio.Confirmacion.ReenviarConfirmacion(ctx, "nadie@example.com"); err != nil || len(tokens) != 2 {
		testingT.Fatalf("correo inexistente: sin error y sin enlace: %v (%d)", err, len(tokens))
	}
	if err := servicio.Confirmacion.ConfirmarCorreo(ctx, tokens[0]); err != nil {
		testingT.Fatalf("el enlace original sigue vigente tras reenviar: %v", err)
	}
	if err := servicio.Confirmacion.ConfirmarCorreo(ctx, tokens[0]); !errors.Is(err, ErrTokenInvalido) {
		testingT.Fatal("el enlace es de un solo uso")
	}
	if err := servicio.Confirmacion.ReenviarConfirmacion(ctx, "c@example.com"); err != nil || len(tokens) != 2 {
		testingT.Fatalf("una cuenta confirmada no recibe otro enlace: %v (%d)", err, len(tokens))
	}
	if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("c@example.com", "clave-larga-ok")); err != nil {
		testingT.Fatalf("tras confirmar debe ingresar: %v", err)
	}
	if ev[ConfirmacionSolicitada] != 2 || ev[CorreoConfirmado] != 1 {
		testingT.Fatalf("observers: %v", ev)
	}

	if _, err := servicio.Registro.Registrar(ctx, SolicitudRegistro{Email: "d@example.com", NombreUso: "Dani", Clave: "clave-larga-ok", Rol: RolEscolar}); err != nil {
		testingT.Fatal(err)
	}
	enlaceDeD := tokens[len(tokens)-1]
	*ahora = ahora.Add(31 * time.Minute)
	if err := servicio.Confirmacion.ConfirmarCorreo(ctx, enlaceDeD); !errors.Is(err, ErrTokenInvalido) {
		testingT.Fatal("el enlace de confirmación vence a los 30 minutos")
	}
	if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("d@example.com", "clave-larga-ok")); !errors.Is(err, ErrCredenciales) {
		testingT.Fatalf("un registro vencido sigue siendo invisible: %v", err)
	}
	liberados, err := servicio.Confirmacion.LimpiarRegistrosVencidos(ctx)
	if err != nil || liberados != 1 {
		testingT.Fatalf("debía liberarse un correo vencido: %d (%v)", liberados, err)
	}
	if _, err := servicio.Registro.Registrar(ctx, SolicitudRegistro{Email: "d@example.com", NombreUso: "Dani", Clave: "clave-larga-ok", Rol: RolEscolar}); err != nil {
		testingT.Fatalf("tras vencer el plazo, el correo vuelve a estar libre: %v", err)
	}
}

func TestRecuperacionNoEnviaAQuienNoConfirmo(testingT *testing.T) {
	servicio, bus, _, _ := nuevo(testingT)
	ctx := context.Background()
	var tokens []string
	bus.Registrar(RecuperacionSolicitada, nuevoObservadorDePrueba("captura-recuperacion", func(e Event) {
		token, _ := e.Payload["token"].(string)
		tokens = append(tokens, token)
	}))
	if _, err := servicio.Registro.Registrar(ctx, SolicitudRegistro{Email: "sin.confirmar@example.com", NombreUso: "Sin", Clave: "clave-larga-ok", Rol: RolEscolar}); err != nil {
		testingT.Fatal(err)
	}
	if err := servicio.Recuperacion.SolicitarRecuperacion(ctx, "sin.confirmar@example.com"); err != nil {
		testingT.Fatal(err)
	}
	if len(tokens) != 0 {
		testingT.Fatalf("un registro sin confirmar no debe recibir enlace de recuperación: %v", tokens)
	}
}

func TestSesionVence(testingT *testing.T) {
	servicio, bus, ahora, _ := nuevo(testingT)
	ctx := context.Background()
	registrarConfirmado(testingT, contextoRegistro{servicio: servicio, bus: bus, datos: SolicitudRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar}})
	tok, _, _ := servicio.Sesion.Ingresar(ctx, credencialesIngreso("e@example.com", "clave-larga-ok"))
	*ahora = ahora.Add(13 * time.Hour)
	if _, err := servicio.Sesion.UsuarioDeSesion(ctx, tok); err == nil {
		testingT.Fatal("sesión vencida")
	}
}

func TestBloqueoPorIntentos(testingT *testing.T) {
	servicio, bus, ahora, _ := nuevo(testingT)
	ctx := context.Background()
	registrarConfirmado(testingT, contextoRegistro{servicio: servicio, bus: bus, datos: SolicitudRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar}})
	for i := 0; i < 5; i++ {
		if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("e@example.com", "mala")); !errors.Is(err, ErrCredenciales) {
			testingT.Fatalf("intento %d: %v", i, err)
		}
	}
	if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("e@example.com", "clave-larga-ok")); !errors.Is(err, ErrDemasiadosIntentos) {
		testingT.Fatalf("debe bloquear incluso con clave correcta: %v", err)
	}
	*ahora = ahora.Add(16 * time.Minute)
	if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("e@example.com", "clave-larga-ok")); err != nil {
		testingT.Fatalf("tras la ventana debe permitir: %v", err)
	}
}

func TestRecuperacion(testingT *testing.T) {
	servicio, bus, ahora, _ := nuevo(testingT)
	ctx := context.Background()
	var token string
	bus.Registrar(RecuperacionSolicitada, nuevoObservadorDePrueba("captura-token", func(e Event) { token, _ = e.Payload["token"].(string) }))
	registrarConfirmado(testingT, contextoRegistro{servicio: servicio, bus: bus, datos: SolicitudRegistro{Email: "e@example.com", NombreUso: "Eli", Clave: "clave-larga-ok", Rol: RolEscolar}})
	sesion, _, _ := servicio.Sesion.Ingresar(ctx, credencialesIngreso("e@example.com", "clave-larga-ok"))

	if err := servicio.Recuperacion.SolicitarRecuperacion(ctx, "nadie@example.com"); err != nil || token != "" {
		testingT.Fatal("correo inexistente: sin error y sin token")
	}
	_ = servicio.Recuperacion.SolicitarRecuperacion(ctx, "e@example.com")
	if token == "" {
		testingT.Fatal("debía emitirse token")
	}
	var ev ErrValidacion
	if err := servicio.Recuperacion.RestablecerClave(ctx, SolicitudRestablecimiento{Token: token, NuevaClave: "corta"}); !errors.As(err, &ev) {
		testingT.Fatalf("clave débil: %v", err)
	}
	if err := servicio.Recuperacion.RestablecerClave(ctx, SolicitudRestablecimiento{Token: token, NuevaClave: "nueva-clave-larga"}); err != nil {
		testingT.Fatalf("la clave débil no debía gastar el token: %v", err)
	}
	if err := servicio.Recuperacion.RestablecerClave(ctx, SolicitudRestablecimiento{Token: token, NuevaClave: "otra-clave-larga"}); !errors.Is(err, ErrTokenInvalido) {
		testingT.Fatal("el token es de un solo uso")
	}
	if _, err := servicio.Sesion.UsuarioDeSesion(ctx, sesion); err == nil {
		testingT.Fatal("restablecer cierra las sesiones previas")
	}
	if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso("e@example.com", "nueva-clave-larga")); err != nil {
		testingT.Fatal(err)
	}
	_ = servicio.Recuperacion.SolicitarRecuperacion(ctx, "e@example.com")
	*ahora = ahora.Add(2 * time.Hour)
	if err := servicio.Recuperacion.RestablecerClave(ctx, SolicitudRestablecimiento{Token: token, NuevaClave: "tercera-clave-larga"}); !errors.Is(err, ErrTokenInvalido) {
		testingT.Fatal("token vencido")
	}
}

func TestObtenerPerfilFabricaSimple(testingT *testing.T) {
	for _, r := range []Rol{RolEscolar, RolApoderado, RolTutor, RolCoordinador} {
		p, err := ObtenerPerfil(r, "uni.edu.pe,uni.pe")
		if err != nil || p.Rol() != r {
			testingT.Fatalf("rol %s: %v", r, err)
		}
	}
	if p, _ := ObtenerPerfil(RolCoordinador, "x"); p.EsPublico() {
		testingT.Fatal("coordinación no es de registro público")
	}
	if _, err := ObtenerPerfil(Rol("intruso"), "x"); err == nil {
		testingT.Fatal("rol desconocido debe fallar")
	}
}

func TestValidadorClaveCambiaDeEstrategia(testingT *testing.T) {
	v := NuevoValidadorClave(PoliticaBasica{Min: 10})
	if len(v.Validar("larga-pero-simple", "a@b.c")) != 0 {
		testingT.Fatal("básica debe aceptar")
	}
	v.FijarPolitica(PoliticaEstricta{Min: 12})
	if len(v.Validar("larga-pero-simple", "a@b.c")) == 0 {
		testingT.Fatal("tras FijarPolitica debe aplicar la estricta")
	}
}

func nuevo(testingT *testing.T) (*Servicio, *busDePrueba, *time.Time, map[string]int) {
	testingT.Helper()
	hasher := hasherDePrueba{}
	bus := nuevoBusDePrueba()
	cuenta := map[string]int{}
	for _, n := range []string{UsuarioRegistrado, IngresoFallido, IngresoExitoso, RecuperacionSolicitada, ClaveRestablecida, ConfirmacionSolicitada, CorreoConfirmado} {
		n := n
		bus.Registrar(n, nuevoObservadorDePrueba("cuenta-"+n, func(Event) { cuenta[n]++ }))
	}
	ahora := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	servicio := NuevoServicio(Dependencias{
		Repositorio: memoria.NuevoRepositorioMemoria(),
		Fabrica:     NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         bus,
		Lockout:     lockout.NuevaMemoriaConReloj(func() time.Time { return ahora }),
	})
	servicio.FijarReloj(func() time.Time { return ahora })
	return servicio, bus, &ahora, cuenta
}

func credencialesIngreso(email, clave string) CredencialesIngreso {
	return CredencialesIngreso{Email: email, Clave: clave}
}

type contextoRegistro struct {
	servicio *Servicio
	bus      *busDePrueba
	datos    SolicitudRegistro
}

func registrarConfirmado(testingT *testing.T, contexto contextoRegistro) *Usuario {
	testingT.Helper()
	var token string
	contexto.bus.Registrar(ConfirmacionSolicitada, nuevoObservadorDePrueba("captura-confirmacion-"+contexto.datos.Email, func(evento Event) {
		token, _ = evento.Payload["token"].(string)
	}))
	usuario, err := contexto.servicio.Registro.Registrar(context.Background(), contexto.datos)
	if err != nil {
		testingT.Fatal(err)
	}
	if err := contexto.servicio.Confirmacion.ConfirmarCorreo(context.Background(), token); err != nil {
		testingT.Fatal(err)
	}
	return usuario
}
