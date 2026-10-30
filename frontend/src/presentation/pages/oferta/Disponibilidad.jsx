import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import CalendarioSemanal from '../../components/CalendarioSemanal.jsx';
import { salir } from '../../../application/acceso/sesion.js';
import {
  diasSemana,
  modoDemoOferta,
  ofertaService,
  validarNuevoSlot,
} from '../../../application/oferta/ofertaService.js';
import { useSesion } from '../../hooks/useSesion.js';
import '../../styles/oferta.css';

const CUPO_MINIMO = 180;
const CUPO_MAXIMO = 360;
const CUPOS_DISPONIBLES = [180, 240, 300, 360];

function crearOpcionesHora() {
  return Array.from({ length: 18 }, (_, indice) => {
    const minutos = (6 + indice) * 60;
    return `${String(Math.floor(minutos / 60)).padStart(2, '0')}:00`;
  });
}

function horaFin(horaInicio) {
  const minutos = Number(horaInicio.slice(0, 2)) * 60 + Number(horaInicio.slice(3)) + 60;
  if (minutos === 24 * 60) return '24:00';
  return `${String(Math.floor(minutos / 60)).padStart(2, '0')}:${String(minutos % 60).padStart(2, '0')}`;
}

function mensajeError(error) {
  if (error?.campos && Object.keys(error.campos).length > 0) {
    return Object.values(error.campos).join(' ');
  }
  return error?.message || 'No se pudo completar la solicitud.';
}

export default function Disponibilidad() {
  const { usuario } = useSesion();
  const [disponibilidad, setDisponibilidad] = useState(null);
  const [cargando, setCargando] = useState(true);
  const [dia, setDia] = useState(1);
  const [inicio, setInicio] = useState('09:00');
  const [error, setError] = useState('');
  const [aviso, setAviso] = useState('');
  const [guardando, setGuardando] = useState(false);
  const [eliminandoId, setEliminandoId] = useState(null);
  const [actualizandoCupo, setActualizandoCupo] = useState(false);

  const slots = disponibilidad?.slots ?? [];
  const cupo = disponibilidad?.cupo_semanal_minutos ?? CUPO_MINIMO;
  const horasInicio = useMemo(crearOpcionesHora, []);
  const minutosAsignados = slots.length * 60;

  const cargar = useCallback(async () => {
    setCargando(true);
    setError('');
    try {
      setDisponibilidad(await ofertaService.obtenerDisponibilidad());
      return true;
    } catch (fallo) {
      setError(mensajeError(fallo));
      return false;
    } finally {
      setCargando(false);
    }
  }, []);

  useEffect(() => {
    cargar();
  }, [cargar]);

  async function agregarSlot(evento) {
    evento.preventDefault();
    setError('');
    setAviso('');
    const nuevo = { dia_semana: Number(dia), inicio, fin: horaFin(inicio) };
    const problema = validarNuevoSlot(nuevo, slots, cupo);
    if (problema) {
      setError(problema);
      return;
    }
    setGuardando(true);
    try {
      await ofertaService.agregarSlots({ slots: [nuevo] });
      if (await cargar()) setAviso('Horario agregado a tu disponibilidad.');
    } catch (fallo) {
      setError(mensajeError(fallo));
    } finally {
      setGuardando(false);
    }
  }

  async function eliminarSlot(slot) {
    setError('');
    setAviso('');
    setEliminandoId(slot.id);
    try {
      await ofertaService.eliminarSlot(slot.id);
      if (await cargar()) setAviso('Horario eliminado.');
    } catch (fallo) {
      setError(mensajeError(fallo));
    } finally {
      setEliminandoId(null);
    }
  }

  async function cambiarCupo(evento) {
    const nuevoCupo = Number(evento.target.value);
    if (nuevoCupo === cupo) return;
    setError('');
    setAviso('');
    setActualizandoCupo(true);
    try {
      await ofertaService.cambiarCupo(nuevoCupo);
      if (await cargar()) setAviso('Cupo semanal actualizado.');
    } catch (fallo) {
      setError(mensajeError(fallo));
    } finally {
      setActualizandoCupo(false);
    }
  }

  return (
    <main className="pagina-oferta">
      <header className="encabezado-oferta">
        <div>
          <Link className="enlace-volver" to="/">← Inicio</Link>
          <p className="sobrelinea">FLUJO C · OFERTA</p>
          <h1>Mi disponibilidad</h1>
          <p className="subtitulo-oferta">
            Organiza tus horarios de tutoría. Cada espacio tiene una duración de 60 minutos.
          </p>
        </div>
        <div className="perfil-tutor">
          <span className="avatar-tutor" aria-hidden="true">
            {(usuario?.nombre_uso || 'T').slice(0, 1).toUpperCase()}
          </span>
          <span>{usuario?.nombre_uso}</span>
          <button type="button" className="boton-secundario" onClick={salir}>Salir</button>
        </div>
      </header>

      {modoDemoOferta && (
        <div className="aviso-demo" role="status">
          <strong>Modo de demostración</strong>
          <span>Los cambios se guardan solo en esta sesión del navegador.</span>
        </div>
      )}

      {error && <div className="alerta-oferta alerta-error" role="alert">{error}</div>}
      {aviso && <div className="alerta-oferta alerta-exito" role="status">{aviso}</div>}

      <section className="resumen-oferta" aria-label="Resumen de disponibilidad">
        <div className="tarjeta-resumen">
          <span className="etiqueta-resumen">Cupo semanal</span>
          <label className="selector-cupo">
            <select
              aria-label="Cupo semanal en minutos"
              value={cupo}
              disabled={cargando || actualizandoCupo}
              onChange={cambiarCupo}
            >
              {CUPOS_DISPONIBLES.map((valor) => (
                <option key={valor} value={valor}>{valor / 60} horas · {valor} min</option>
              ))}
            </select>
            <span className="ayuda-resumen">Entre {CUPO_MINIMO / 60} y {CUPO_MAXIMO / 60} horas por semana</span>
          </label>
        </div>
        <div className="tarjeta-resumen">
          <span className="etiqueta-resumen">Horas publicadas</span>
          <strong className="valor-resumen">{minutosAsignados / 60} <small>/ {cupo / 60} h</small></strong>
          <span className="ayuda-resumen">{minutosAsignados} de {cupo} minutos asignados</span>
        </div>
        <div className="tarjeta-resumen tarjeta-canal">
          <span className="etiqueta-resumen">Canal privado</span>
          <strong className="canal-estado"><span aria-hidden="true">●</span> Solo para ti</strong>
          <span className="ayuda-resumen">Tu disponibilidad requiere una sesión de tutor.</span>
        </div>
      </section>

      <section className="panel-oferta">
        <div className="cabecera-panel">
          <div>
            <p className="sobrelinea">TU SEMANA</p>
            <h2>Calendario semanal</h2>
          </div>
          <span className="zona-horaria">UTC−05:00 · America/Lima</span>
        </div>
        {cargando ? (
          <p className="estado-carga" role="status">Cargando disponibilidad…</p>
        ) : (
          <CalendarioSemanal slots={slots} alEliminar={eliminarSlot} eliminandoId={eliminandoId} />
        )}
      </section>

      <section className="panel-oferta panel-agregar">
        <div>
          <p className="sobrelinea">NUEVO HORARIO</p>
          <h2>Agregar un espacio</h2>
          <p className="texto-ayuda">Los horarios usan intervalos [inicio, fin): dos espacios contiguos no se solapan.</p>
        </div>
        <form className="formulario-slot" onSubmit={agregarSlot}>
          <label>
            Día
            <select value={dia} onChange={(evento) => setDia(evento.target.value)}>
              {diasSemana.map((opcion) => (
                <option key={opcion.valor} value={opcion.valor}>{opcion.nombre}</option>
              ))}
            </select>
          </label>
          <label>
            Inicio
            <select value={inicio} onChange={(evento) => setInicio(evento.target.value)}>
              {horasInicio.map((hora) => <option key={hora} value={hora}>{hora}</option>)}
            </select>
          </label>
          <div className="fin-slot">
            <span>Fin</span>
            <strong>{horaFin(inicio)}</strong>
            <small>60 minutos</small>
          </div>
          <button type="submit" disabled={guardando || cargando}>
            {guardando ? 'Guardando…' : 'Agregar horario'}
          </button>
        </form>
        <p className="nota-edicion">
          Para cambiar un horario existente, elimínalo y agrega el nuevo espacio.
        </p>
      </section>

      <section className="panel-informacion" aria-label="Privacidad y excepciones">
        <div className="icono-privacidad" aria-hidden="true">◈</div>
        <div>
          <h2>Tu espacio privado</h2>
          <p>
            Esta disponibilidad se consulta con tu sesión de tutor y no es pública. El backend actual
            todavía no expone un identificador o metadatos de canal privado.
          </p>
          <p>
            Las excepciones por fecha tampoco están disponibles en la API actual; aquí se muestran
            únicamente los horarios semanales recurrentes.
          </p>
        </div>
      </section>
    </main>
  );
}
