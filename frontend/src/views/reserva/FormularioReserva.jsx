import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import Campo from '../../components/Campo.jsx';
import {
  crearReserva,
  listarSlots,
  subirAdjunto,
  urlAdjunto,
} from '../../controllers/reserva/apiReserva.js';
import {
  adjuntosDeArchivos,
  contarCaracteres,
  formatearTamano,
  validarReserva,
} from '../../models/reserva/validadores.js';

const FORMULARIO_INICIAL = { slot_id: '', tema: '', objetivo: '', grado: '' };

function crearClaveIdempotencia() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  return `reserva-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function describirAdjunto(adjunto) {
  if (adjunto.error) return `No se pudo subir: ${adjunto.error}`;
  if (adjunto.estado === 'validado') return 'Listo para descargar';
  if (adjunto.estado === 'rechazado') return `Rechazado: ${adjunto.motivo || 'no permitido'}`;
  return `En cuarentena: ${adjunto.motivo || 'pendiente de análisis'}`;
}

export default function FormularioReserva() {
  const [slots, setSlots] = useState([]);
  const [formulario, setFormulario] = useState(FORMULARIO_INICIAL);
  const [archivos, setArchivos] = useState([]);
  const [errores, setErrores] = useState({});
  const [mensaje, setMensaje] = useState('');
  const [cargandoSlots, setCargandoSlots] = useState(true);
  const [enviando, setEnviando] = useState(false);
  const [reserva, setReserva] = useState(null);
  const [adjuntos, setAdjuntos] = useState([]);
  const noDisponibles = slots.filter((slot) => !slot.disponible);

  useEffect(() => {
    let activo = true;
    setCargandoSlots(true);
    listarSlots()
      .then((datos) => {
        if (activo) setSlots(datos?.slots ?? []);
      })
      .catch((error) => {
        if (activo) setMensaje(error.message);
      })
      .finally(() => {
        if (activo) setCargandoSlots(false);
      });
    return () => {
      activo = false;
    };
  }, [reserva]);

  const cambiarCampo = (campo) => (evento) =>
    setFormulario((actual) => ({ ...actual, [campo]: evento.target.value }));

  const elegirArchivos = (evento) => setArchivos(Array.from(evento.target.files ?? []));

  async function enviar(evento) {
    evento.preventDefault();
    setMensaje('');
    const datos = { ...formulario };
    const encontrados = validarReserva({ ...datos, adjuntos: adjuntosDeArchivos(archivos) });
    setErrores(encontrados);
    if (Object.keys(encontrados).length > 0) return;

    setEnviando(true);
    try {
      // La reserva se crea primero: si un adjunto falla, la reserva queda confirmada.
      const creada = await crearReserva(datos, crearClaveIdempotencia());
      setReserva(creada);
      setFormulario(FORMULARIO_INICIAL);
      const resultados = [];
      for (const archivo of archivos) {
        try {
          resultados.push(await subirAdjunto(creada.id, archivo));
        } catch (error) {
          resultados.push({ nombre: archivo.name, error: error.message });
        }
      }
      setAdjuntos(resultados);
      setArchivos([]);
    } catch (error) {
      setErrores(error.campos ?? {});
      setMensaje(error.message);
    } finally {
      setEnviando(false);
    }
  }

  function reiniciar() {
    setReserva(null);
    setAdjuntos([]);
    setErrores({});
    setMensaje('');
  }

  if (reserva) {
    return (
      <main>
        <h1>Reserva confirmada</h1>
        <p role="status">
          Tu sesión de <strong>{reserva.curso}</strong> con {reserva.tutor_nombre} quedó
          agendada para el <strong>{reserva.inicio_lima}</strong> (hora de Lima).
        </p>
        <ul>
          <li>Tema: {reserva.tema}</li>
          <li>Grado: {reserva.grado}</li>
          <li>Duración: 60 minutos</li>
          <li>Número de reserva: {reserva.id}</li>
        </ul>

        <h2>Adjuntos</h2>
        {adjuntos.length === 0 ? (
          <p>Sin adjuntos.</p>
        ) : (
          <ul>
            {adjuntos.map((adjunto) => (
              <li key={adjunto.id ?? adjunto.nombre}>
                {adjunto.nombre} {adjunto.tamano_bytes ? `· ${formatearTamano(adjunto.tamano_bytes)}` : ''}
                {' — '}
                {describirAdjunto(adjunto)}
                {adjunto.estado === 'validado' && (
                  <> · <a href={urlAdjunto(reserva.id, adjunto.id)} download>Descargar</a></>
                )}
              </li>
            ))}
          </ul>
        )}
        <p>
          <small>
            Los adjuntos pasan por cuarentena antes de habilitarse. Si el análisis no está
            disponible, la reserva sigue confirmada y el archivo queda en revisión.
          </small>
        </p>

        <button type="button" onClick={reiniciar}>Hacer otra reserva</button>
        <p><Link to="/">Volver al inicio</Link></p>
      </main>
    );
  }

  return (
    <main>
      <h1>Reservar una tutoría</h1>
      <p>
        Elige un horario publicado. La reserva se confirma al enviarla y debe empezar al menos
        48 horas después de este momento.
      </p>
      <form onSubmit={enviar} noValidate>
        <div className="campo">
          <label htmlFor="slot">Horario</label>
          <select
            id="slot"
            value={formulario.slot_id}
            onChange={cambiarCampo('slot_id')}
            aria-invalid={!!errores.slot_id}
            aria-describedby={errores.slot_id ? 'slot-error' : 'slot-ayuda'}
            disabled={cargandoSlots}
          >
            <option value="">Elige un horario</option>
            {slots.map((slot) => (
              <option key={slot.id} value={slot.id} disabled={!slot.disponible}>
                {slot.curso} · {slot.inicio_lima}
                {slot.disponible ? '' : ` · no disponible: ${slot.motivo}`}
              </option>
            ))}
          </select>
          {cargandoSlots && <small id="slot-ayuda">Cargando horarios…</small>}
          {!cargandoSlots && !errores.slot_id && (
            <small id="slot-ayuda">Los horarios que no cumplen las 48 horas aparecen deshabilitados.</small>
          )}
          {errores.slot_id && <small id="slot-error" className="error" role="alert">{errores.slot_id}</small>}
          {!cargandoSlots && noDisponibles.length > 0 && (
            <details className="no-disponibles">
              <summary>{noDisponibles.length} horario(s) no disponible(s)</summary>
              <ul>
                {noDisponibles.map((slot) => (
                  <li key={slot.id}>{slot.curso} · {slot.inicio_lima}: {slot.motivo}</li>
                ))}
              </ul>
            </details>
          )}
        </div>

        <div className="campo">
          <label htmlFor="tema">Tema</label>
          <input
            id="tema"
            value={formulario.tema}
            onChange={cambiarCampo('tema')}
            aria-invalid={!!errores.tema}
            aria-describedby="tema-ayuda"
            maxLength={200}
          />
          <small id="tema-ayuda">Entre 10 y 200 caracteres. Llevas {contarCaracteres(formulario.tema)}.</small>
          {errores.tema && <small className="error" role="alert">{errores.tema}</small>}
        </div>

        <div className="campo">
          <label htmlFor="objetivo">Objetivo</label>
          <textarea
            id="objetivo"
            rows={4}
            value={formulario.objetivo}
            onChange={cambiarCampo('objetivo')}
            aria-invalid={!!errores.objetivo}
            aria-describedby="objetivo-ayuda"
            maxLength={1000}
          />
          <small id="objetivo-ayuda">Entre 20 y 1000 caracteres. Llevas {contarCaracteres(formulario.objetivo)}.</small>
          {errores.objetivo && <small className="error" role="alert">{errores.objetivo}</small>}
        </div>

        <Campo
          etiqueta="Grado"
          value={formulario.grado}
          onChange={cambiarCampo('grado')}
          error={errores.grado}
          ayuda="Por ejemplo, 5to de secundaria."
          maxLength={80}
        />

        <div className="campo">
          <label htmlFor="adjuntos">Adjuntos (opcional)</label>
          <input id="adjuntos" type="file" multiple onChange={elegirArchivos} />
          <small>
            Hasta 5 archivos de 10 MB (PDF, PNG, JPG, TXT o DOCX). Se suben después de confirmar
            la reserva y pasan por cuarentena.
          </small>
          {archivos.length > 0 && (
            <ul>
              {archivos.map((archivo) => (
                <li key={archivo.name}>{archivo.name} · {formatearTamano(archivo.size)}</li>
              ))}
            </ul>
          )}
          {errores.adjuntos && <small className="error" role="alert">{errores.adjuntos}</small>}
        </div>

        {mensaje && <p className="error" role="alert">{mensaje}</p>}
        <button disabled={enviando}>{enviando ? 'Reservando…' : 'Confirmar reserva'}</button>
      </form>
      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}
