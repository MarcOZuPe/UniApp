import { diasSemana } from '../../application/oferta/ofertaService.js';

function nombreDia(valor) {
  return diasSemana.find((dia) => dia.valor === valor)?.nombre ?? 'Día desconocido';
}

export default function CalendarioSemanal({ slots, alEliminar, eliminandoId }) {
  return (
    <div className="calendario-semanal" aria-label="Disponibilidad semanal">
      {diasSemana.map((dia) => {
        const delDia = slots
          .filter((slot) => slot.dia_semana === dia.valor)
          .sort((a, b) => a.inicio.localeCompare(b.inicio));
        return (
          <section className="dia-calendario" key={dia.valor} aria-label={nombreDia(dia.valor)}>
            <h3>{dia.nombre}</h3>
            {delDia.length === 0 ? (
              <p className="dia-vacio">Sin disponibilidad</p>
            ) : (
              <ul className="lista-slots">
                {delDia.map((slot) => (
                  <li className="slot-calendario" key={slot.id}>
                    <span className="slot-hora">
                      {slot.inicio}–{slot.fin}
                    </span>
                    <span className="slot-duracion">60 min</span>
                    <button
                      className="boton-texto"
                      type="button"
                      aria-label={`Eliminar horario ${slot.inicio} del ${dia.nombre}`}
                      disabled={eliminandoId === slot.id}
                      onClick={() => alEliminar(slot)}
                    >
                      {eliminandoId === slot.id ? 'Eliminando…' : 'Quitar'}
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </section>
        );
      })}
    </div>
  );
}
