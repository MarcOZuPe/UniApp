package sesion

import (
	"time"

	modelo "uniapp/internal/application/sesion"
)

// MinutosJSON es la vista de un registro de minutos (I17).
type MinutosJSON struct {
	SesionID     string    `json:"sesion_id"`
	ReservaID    string    `json:"reserva_id"`
	TutorID      string    `json:"tutor_id"`
	EscolarID    string    `json:"escolar_id"`
	FinSesion    time.Time `json:"fin_sesion"`
	Estado       string    `json:"estado"`
	Propuestos   int       `json:"minutos_propuestos"`
	Verificados  *int      `json:"minutos_verificados"`
	Computables  int       `json:"minutos_computables"`
	DisputadoPor string    `json:"disputado_por,omitempty"`
}

// EntradaAuditoriaJSON es la vista de una entrada de auditoria de minutos.
type EntradaAuditoriaJSON struct {
	Accion        string    `json:"accion"`
	ActorID       string    `json:"actor_id"`
	ValorAnterior *int      `json:"valor_anterior"`
	ValorNuevo    *int      `json:"valor_nuevo"`
	Motivo        string    `json:"motivo"`
	RegistradaEn  time.Time `json:"registrada_en"`
}

func PresentarMinutos(r *modelo.RegistroMinutos) MinutosJSON {
	return MinutosJSON{
		SesionID:     r.SesionID(),
		ReservaID:    r.ReservaID(),
		TutorID:      r.TutorID(),
		EscolarID:    r.EscolarID(),
		FinSesion:    r.FinSesion(),
		Estado:       string(r.Estado()),
		Propuestos:   r.MinutosPropuestos(),
		Verificados:  puntero(r.MinutosVerificados()),
		Computables:  r.MinutosComputables(),
		DisputadoPor: r.DisputadoPor(),
	}
}

func PresentarListaMinutos(registros []*modelo.RegistroMinutos) []MinutosJSON {
	resultado := make([]MinutosJSON, 0, len(registros))
	for _, r := range registros {
		resultado = append(resultado, PresentarMinutos(r))
	}
	return resultado
}

func PresentarAuditoria(entradas []modelo.EntradaAuditoria) []EntradaAuditoriaJSON {
	resultado := make([]EntradaAuditoriaJSON, 0, len(entradas))
	for _, e := range entradas {
		resultado = append(resultado, EntradaAuditoriaJSON{
			Accion:        string(e.Accion()),
			ActorID:       e.ActorID(),
			ValorAnterior: puntero(e.ValorAnterior()),
			ValorNuevo:    puntero(e.ValorNuevo()),
			Motivo:        e.Motivo(),
			RegistradaEn:  e.RegistradaEn(),
		})
	}
	return resultado
}

func puntero(valor int, presente bool) *int {
	if !presente {
		return nil
	}
	return &valor
}
