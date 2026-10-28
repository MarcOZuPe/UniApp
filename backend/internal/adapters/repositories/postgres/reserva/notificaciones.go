package reserva

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// I10: llamada exclusivamente antes del commit de una reserva NUEVA de I09.
// No resuelve direcciones ni envía correo dentro de la transacción.
func encolarNotificaciones(ctx context.Context, tx *sql.Tx, id string, r *Reserva, anticipacion time.Duration) error {
	if anticipacion <= 0 {
		anticipacion = 24 * time.Hour
	}
	lima := time.FixedZone("Lima", -5*60*60)
	cuerpo := fmt.Sprintf("Reserva %s. Tutoría de %s: %s (hora de Lima).", id, r.Curso, r.Inicio.In(lima).Format("02/01/2006 15:04"))
	recordatorio := r.Inicio.Add(-anticipacion)
	if recordatorio.Before(r.CreadaEn) {
		recordatorio = r.CreadaEn
	}
	for _, destinatario := range []string{r.EscolarID, r.TutorID} {
		for _, tipo := range []string{"confirmacion", "recordatorio"} {
			cuando, asunto := r.CreadaEn, "Reserva confirmada"
			if tipo == "recordatorio" {
				if !r.Inicio.After(r.CreadaEn) {
					continue
				}
				cuando, asunto = recordatorio, "Recordatorio de tutoría"
			}
			for _, canal := range []string{"interno", "correo"} {
				if _, err := tx.ExecContext(ctx, `INSERT INTO sesion_outbox
      (reserva_id,destinatario_id,tipo,canal,asunto,cuerpo,disponible_en)
      VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (reserva_id,destinatario_id,tipo,canal) DO NOTHING`,
					id, destinatario, tipo, canal, asunto, cuerpo, cuando); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
