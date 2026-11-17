import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  obtenerMiExpediente,
  presentarExpediente,
} from '../../../adapters/http/tutorApi.js';

const MAXIMO_ARCHIVO = 10 * 1024 * 1024;

function estaVencido(expediente) {
  if (expediente?.estado !== 'aprobado' || !expediente.vence_en) return false;
  const vencimiento = new Date(expediente.vence_en);
  return !Number.isNaN(vencimiento.getTime()) && vencimiento.getTime() <= Date.now();
}

function fechaLegible(fecha) {
  return new Date(fecha).toLocaleDateString();
}

export default function ExpedienteTutor() {
  const [expediente, setExpediente] = useState(null);
  const [sinExpediente, setSinExpediente] = useState(false);
  const [cargando, setCargando] = useState(true);
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState('');
  const [camposError, setCamposError] = useState({});

  async function cargar() {
    setCargando(true);
    setError('');
    try {
      setExpediente(await obtenerMiExpediente());
      setSinExpediente(false);
    } catch (fallo) {
      if (fallo.estado === 404) {
        setExpediente(null);
        setSinExpediente(true);
      } else {
        setError(fallo.message);
      }
    } finally {
      setCargando(false);
    }
  }

  useEffect(() => {
    cargar();
  }, []);

  async function enviar(evento) {
    evento.preventDefault();
    setError('');
    setCamposError({});
    const formulario = new FormData(evento.currentTarget);
    const documentos = formulario.getAll('documentos');
    if (documentos.length < 1 || documentos.length > 5) {
      setCamposError({ documentos: 'Adjunta entre 1 y 5 archivos PDF.' });
      return;
    }
    if (documentos.some((archivo) => archivo.size > MAXIMO_ARCHIVO)) {
      setCamposError({ documentos: 'Cada archivo debe pesar como máximo 10 MB.' });
      return;
    }

    setEnviando(true);
    try {
      setExpediente(await presentarExpediente(formulario));
      setSinExpediente(false);
    } catch (fallo) {
      setError(fallo.message);
      setCamposError(fallo.campos || {});
    } finally {
      setEnviando(false);
    }
  }

  if (cargando) return <main><p role="status">Cargando expediente…</p></main>;

  const vencido = estaVencido(expediente);
  const estadoVisual = vencido ? 'vencido' : expediente?.estado;
  const mostrarFormulario = !expediente || expediente.estado === 'rechazado' || vencido;
  const puedeReenviar = Boolean(expediente) && (expediente.estado === 'rechazado' || vencido);

  return (
    <main className="pagina-tutor">
      <h1>Mi expediente de tutor</h1>
      {error && <p className="error" role="alert">{error}</p>}
      {expediente && expediente.estado !== 'rechazado' && (
        <section aria-live="polite">
          <h2>Estado: {estadoVisual}</h2>
          {expediente.estado === 'pendiente' && <p>Coordinación revisará tu expediente antes de habilitar tu perfil.</p>}
          {expediente.estado === 'aprobado' && !vencido && (
            <p>Aprobado hasta {fechaLegible(expediente.vence_en)}.</p>
          )}
          {vencido && (
            <p>
              Tu expediente venció el {fechaLegible(expediente.vence_en)}. Vuelve a presentarlo para que
              coordinación lo revise nuevamente.
            </p>
          )}
          {expediente.motivo_rechazo && <p>Motivo: {expediente.motivo_rechazo}</p>}
          <h3>Documentos enviados</h3>
          <ul>
            {expediente.documentos.map((documento) => (
              <li key={documento.id}>
                <a href={`/api/tutor/documentos/${encodeURIComponent(documento.id)}`}>{documento.nombre}</a>
              </li>
            ))}
          </ul>
        </section>
      )}
      {expediente?.estado === 'rechazado' && (
        <p role="alert">Tu expediente fue rechazado. Motivo: {expediente.motivo_rechazo}</p>
      )}
      {mostrarFormulario && (
        <>
          <p>
            {sinExpediente
              ? 'Completa tu expediente y adjunta documentos PDF. Tu perfil no se publicará hasta que coordinación lo apruebe.'
              : vencido
                ? 'Tu aprobación venció. Actualiza los datos y documentos si es necesario y vuelve a presentarlos.'
                : 'Puedes volver a presentar el expediente con los cambios indicados por coordinación.'}
          </p>
          <form onSubmit={enviar} encType="multipart/form-data">
            <label className="campo">
              Nombre de uso y breve presentación
              <textarea
                name="presentacion"
                required
                maxLength={500}
                rows={4}
                defaultValue={expediente?.presentacion ?? ''}
              />
              {camposError.presentacion && <small className="error">{camposError.presentacion}</small>}
            </label>
            <label className="campo">
              Cursos o temas que puedes enseñar
              <textarea
                name="cursos"
                required
                maxLength={500}
                rows={3}
                defaultValue={expediente?.cursos ?? ''}
              />
              {camposError.cursos && <small className="error">{camposError.cursos}</small>}
            </label>
            <label className="campo">
              Experiencia o formación relevante
              <textarea
                name="experiencia"
                required
                maxLength={2000}
                rows={4}
                defaultValue={expediente?.experiencia ?? ''}
              />
              {camposError.experiencia && <small className="error">{camposError.experiencia}</small>}
            </label>
            <label className="campo">
              Documentos de sustento (PDF, hasta 10 MB cada uno; máximo 5)
              <input name="documentos" type="file" accept="application/pdf,.pdf" multiple required />
              {camposError.documentos && <small className="error">{camposError.documentos}</small>}
            </label>
            <button type="submit" disabled={enviando}>
              {enviando ? 'Enviando…' : puedeReenviar ? 'Volver a presentar' : 'Enviar a revisión'}
            </button>
          </form>
        </>
      )}
      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}
