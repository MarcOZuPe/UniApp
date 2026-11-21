import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { obtenerAprobados } from '../../../adapters/http/tutorApi.js';

function estaVencido(expediente) {
  if (!expediente.vence_en) return false;
  const vencimiento = new Date(expediente.vence_en);
  return !Number.isNaN(vencimiento.getTime()) && vencimiento.getTime() <= Date.now();
}

function fechaLegible(fecha) {
  return new Date(fecha).toLocaleDateString();
}

export default function TutoresAprobados() {
  const [expedientes, setExpedientes] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let activo = true;
    obtenerAprobados()
      .then((resultado) => {
        if (activo) setExpedientes(resultado.expedientes ?? []);
      })
      .catch((fallo) => {
        if (activo) setError(fallo.message);
      })
      .finally(() => {
        if (activo) setCargando(false);
      });
    return () => {
      activo = false;
    };
  }, []);

  return (
    <main className="pagina-tutor">
      <h1>Expedientes aprobados</h1>
      <p>Consulta el historial de expedientes aprobados por coordinación, incluidos los que ya vencieron.</p>
      {error && <p className="error" role="alert">{error}</p>}
      {cargando && <p role="status">Cargando expedientes aprobados…</p>}
      {!cargando && !error && expedientes.length === 0 && (
        <p>No hay expedientes aprobados.</p>
      )}
      {expedientes.map((expediente) => {
        const vencido = estaVencido(expediente);
        return (
          <article className="expediente" key={expediente.tutor_id}>
            <h2>Expediente {expediente.tutor_id}</h2>
            <p>
              <strong>Estado:</strong> {vencido ? 'Aprobado, vencido' : 'Aprobado, vigente'}
            </p>
            <p><strong>Presentación:</strong> {expediente.presentacion}</p>
            <p><strong>Cursos:</strong> {expediente.cursos}</p>
            <p><strong>Experiencia/formación:</strong> {expediente.experiencia}</p>
            {expediente.vence_en && (
              <p><strong>Aprobado hasta:</strong> {fechaLegible(expediente.vence_en)}</p>
            )}
            <h3>Documentos privados</h3>
            <ul>
              {expediente.documentos.map((documento) => (
                <li key={documento.id}>
                  <a href={`/api/tutor/documentos/${encodeURIComponent(documento.id)}`}>{documento.nombre}</a>
                </li>
              ))}
            </ul>
          </article>
        );
      })}
      <p><Link to="/coordinacion/tutores">Volver a pendientes</Link></p>
      <p><Link to="/coordinacion">Volver a coordinación</Link></p>
    </main>
  );
}
