package reserva

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// I09: los adjuntos viven en un almacén privado y pasan por cuarentena. La reserva se
// confirma antes de subirlos, así un fallo de análisis nunca la tumba.

type EstadoAdjunto string

const (
	AdjuntoEnCuarentena EstadoAdjunto = "en_cuarentena"
	AdjuntoValidado     EstadoAdjunto = "validado"
	AdjuntoRechazado    EstadoAdjunto = "rechazado"
)

var ErrAnalisisNoDisponible = errors.New("el análisis de adjuntos no está disponible")

// Analizador revisa el contenido y decide el estado del adjunto. Si el servicio no está
// disponible se devuelve ErrAnalisisNoDisponible y el adjunto queda en cuarentena.
type Analizador interface {
	Analizar(nombre, tipo string, contenido []byte) (EstadoAdjunto, string, error)
}

// AnalizadorNoDisponible representa el análisis apagado (por ejemplo, sin antivirus en el
// entorno). No es un error para la reserva: el adjunto queda en cuarentena.
type AnalizadorNoDisponible struct{}

func (AnalizadorNoDisponible) Analizar(string, string, []byte) (EstadoAdjunto, string, error) {
	return AdjuntoEnCuarentena, "pendiente de análisis: el servicio no está disponible", ErrAnalisisNoDisponible
}

// AnalizadorBasico es la verificación mínima disponible en el MVP: rechaza la firma de
// prueba EICAR y los tipos que no corresponden al contenido detectado.
type AnalizadorBasico struct{}

func (AnalizadorBasico) Analizar(_ string, tipo string, contenido []byte) (EstadoAdjunto, string, error) {
	if len(contenido) == 0 {
		return AdjuntoRechazado, "el archivo está vacío", nil
	}
	if bytes.Contains(bytes.ToUpper(contenido), []byte("EICAR-STANDARD-ANTIVIRUS-TEST-FILE")) {
		return AdjuntoRechazado, "el análisis detectó una firma de prueba de virus", nil
	}
	if _, permitido := TiposAdjuntoPermitidos[tipo]; !permitido {
		return AdjuntoRechazado, "el tipo de archivo no está permitido", nil
	}
	return AdjuntoValidado, "", nil
}

// Almacen guarda el contenido fuera de la base de datos. Lo implementa un adaptador de
// infraestructura; el modelo solo conoce el puerto.
type Almacen interface {
	Guardar(ctx context.Context, clave string, contenido []byte) error
	Leer(ctx context.Context, clave string) ([]byte, error)
	Eliminar(ctx context.Context, clave string) error
}

// TiposAdjuntoPermitidos asocia el tipo detectado por contenido con la extensión que debe
// llevar el nombre del archivo.
var TiposAdjuntoPermitidos = map[string]string{
	"application/pdf": ".pdf",
	"image/png":       ".png",
	"image/jpeg":      ".jpg",
	"text/plain":      ".txt",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
}

// ValidarEntradaAdjunto revisa nombre y tamaño antes de leer el contenido.
func ValidarEntradaAdjunto(nombre string, tamano int64) error {
	if strings.TrimSpace(nombre) == "" {
		return ErrAdjuntoInvalido
	}
	nombres := strings.TrimSpace(nombre)
	if utf8.RuneCountInString(nombres) > NombreAdjuntoMaximo {
		return ErrAdjuntoInvalido
	}
	if strings.ContainsAny(nombres, "/\\") || strings.HasPrefix(nombres, ".") {
		return ErrAdjuntoInvalido
	}
	if tamano <= 0 || tamano > AdjuntoMaximoBytes {
		return ErrAdjuntoInvalido
	}
	return nil
}

// TipoDetectado devuelve el tipo real del contenido; el tipo declarado por el cliente no
// se usa para decidir permisos.
func TipoDetectado(contenido []byte) string {
	if len(contenido) == 0 {
		return "application/octet-stream"
	}
	tipo := http.DetectContentType(contenido)
	// DetectContentType agrega parámetros (por ejemplo "; charset=utf-8").
	if indice := strings.IndexByte(tipo, ';'); indice >= 0 {
		tipo = strings.TrimSpace(tipo[:indice])
	}
	return tipo
}

// ExtensionPermitida indica la extensión esperada para un tipo detectado.
func ExtensionPermitida(tipo string) (string, bool) {
	extension, ok := TiposAdjuntoPermitidos[tipo]
	return extension, ok
}

// HuellaAdjunto calcula SHA-256 para trazabilidad y para detectar duplicados.
func HuellaAdjunto(contenido []byte) string {
	suma := sha256.Sum256(contenido)
	return hex.EncodeToString(suma[:])
}

// NuevoAdjunto arma el registro de metadatos de un archivo ya almacenado.
func NuevoAdjunto(nombre, tipo string, tamano int64, huella, clave string, estado EstadoAdjunto, motivo string, ahora time.Time) (*Adjunto, error) {
	if strings.TrimSpace(nombre) == "" || strings.TrimSpace(clave) == "" {
		return nil, ErrAdjuntoInvalido
	}
	if _, permitido := TiposAdjuntoPermitidos[tipo]; !permitido {
		return nil, ErrAdjuntoInvalido
	}
	if estado == "" {
		estado = AdjuntoEnCuarentena
	}
	return &Adjunto{
		Nombre:              strings.TrimSpace(nombre),
		Tipo:                tipo,
		TamanoBytes:         tamano,
		Estado:              estado,
		Motivo:              motivo,
		Huella:              huella,
		ClaveAlmacenamiento: clave,
		CreadoEn:            ahora.UTC(),
	}, nil
}

// Descargable indica si el adjunto puede entregarse por el enlace privado.
func (a Adjunto) Descargable() bool { return a.Estado == AdjuntoValidado }
