import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  obtenerPendientes,
  revisarExpediente,
} from '../../controllers/tutor/apiTutor.js';

function fechaHoyLocal() {
  const ahora = new Date();
  ahora.setMinutes(ahora.getMinutes() - ahora.getTimezoneOffset());
  return ahora.toISOString().slice(0, 10);
}

export default function RevisarTutores() {
  const [expedientes, setExpedientes] = useState([]);
  const [fechas, setFechas] = useState({});
  const [motivos, setMotivos] = useState({});
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState('');
  const [procesando, setProcesando] = useState('');
  const [validacion, setValidacion] = useState({});

  const cargar = useCallback(async () => {
    setError('');
    try {
      const resultado = await obtenerPendientes();
      setExpedientes(resultado.expedientes);
    } catch (fallo) {
      setError(fallo.message);
    } finally {
      setCargando(false);
    }
  }, []);

  useEffect(() => {
    cargar();
  }, [cargar]);

  async function decidir(tutorId, decision) {
    setProcesando(tutorId);
    setError('');
    setValidacion({});
    try {
      const datos = decision === 'aprobado'
        ? { estado: decision, vence_en: new Date(`${fechas[tutorId]}T23:59:59.999`).toISOString() }
        : { estado: decision, motivo_rechazo: motivos[tutorId] || '' };
      await revisarExpediente(tutorId, datos);
      await cargar();
    } catch (fallo) {
      setError(fallo.message);
      setValidacion(fallo.campos || {});
    } finally {
      setProcesando('');
    }
  }

  return (
    <main className="pagina-tutor">
      <h1>Revisión de expedientes de tutores</h1>
      {error && <p className="error" role="alert">{error}</p>}
      {cargando && <p role="status">Cargando expedientes…</p>}
      {!cargando && expedientes.length === 0 && <p>No hay expedientes pendientes de revisión.</p>}
      {expedientes.map((expediente) => (
        <article className="expediente" key={expediente.tutor_id}>
          <h2>Expediente {expediente.tutor_id}</h2>
          <p><strong>Presentación:</strong> {expediente.presentacion}</p>
          <p><strong>Cursos:</strong> {expediente.cursos}</p>
          <p><strong>Experiencia/formación:</strong> {expediente.experiencia}</p>
          <h3>Documentos privados</h3>
          <ul>
            {expediente.documentos.map((documento) => (
              <li key={documento.id}>
                <a href={`/api/tutor/documentos/${encodeURIComponent(documento.id)}`}>{documento.nombre}</a>
              </li>
            ))}
          </ul>
          <label className="campo">
            Aprobado hasta
            <input
              type="date"
              value={fechas[expediente.tutor_id] || ''}
              onChange={(evento) => setFechas({ ...fechas, [expediente.tutor_id]: evento.target.value })}
              min={fechaHoyLocal()}
            />
            {validacion.vence_en && <small className="error">{validacion.vence_en}</small>}
          </label>
          <button
            type="button"
            disabled={procesando === expediente.tutor_id || !fechas[expediente.tutor_id]}
            onClick={() => decidir(expediente.tutor_id, 'aprobado')}
          >
            Aprobar
          </button>
          <label className="campo">
            Motivo del rechazo
            <textarea
              value={motivos[expediente.tutor_id] || ''}
              onChange={(evento) => setMotivos({ ...motivos, [expediente.tutor_id]: evento.target.value })}
              rows={2}
            />
            {validacion.motivo_rechazo && <small className="error">{validacion.motivo_rechazo}</small>}
          </label>
          <button
            type="button"
            disabled={procesando === expediente.tutor_id || !motivos[expediente.tutor_id]?.trim()}
            onClick={() => decidir(expediente.tutor_id, 'rechazado')}
          >
            Rechazar
          </button>
        </article>
      ))}
      <p><Link to="/coordinacion">Volver a coordinación</Link></p>
    </main>
  );
}
