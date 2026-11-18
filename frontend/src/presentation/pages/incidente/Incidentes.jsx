import { useEffect, useMemo, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';

import {
  listarIncidentes,
  revisarIncidente,
  urlEvidencia,
} from '../../../adapters/http/incidenteApi.js';
import {
  etiquetaTipo,
  formatearTamano,
  validarRevision,
} from '../../../domain/incidente/validadores.js';
import { useSesion } from '../../hooks/useSesion.js';
import '../../styles/incidentes.css';

function describirEstado(estado) {
  return estado === 'revisado' ? 'Revisado por coordinación' : 'Pendiente de revisión';
}

function describirResultado(resultado) {
  if (resultado === 'procede') return 'El incidente procede';
  if (resultado === 'no_procede') return 'El incidente no procede';
  return '';
}

function describirReportante(incidente) {
  const rol = incidente.reportado_por_rol
    || (incidente.reportado_por === incidente.escolar_id ? 'escolar' : 'apoderado');
  return rol === 'escolar' ? 'Escolar' : 'Apoderado';
}

function agruparPorTutor(incidentes) {
  const grupos = new Map();
  for (const incidente of incidentes) {
    const clave = incidente.tutor_id || incidente.tutor_nombre;
    if (!grupos.has(clave)) {
      grupos.set(clave, {
        clave,
        tutorNombre: incidente.tutor_nombre,
        incidentes: [],
        pendientes: 0,
      });
    }
    const grupo = grupos.get(clave);
    grupo.incidentes.push(incidente);
    if (incidente.estado === 'pendiente') grupo.pendientes += 1;
  }
  return Array.from(grupos.values()).sort(
    (a, b) => b.pendientes - a.pendientes || a.tutorNombre.localeCompare(b.tutorNombre, 'es')
  );
}

function contarIncidentes(cantidad) {
  return `${cantidad} ${cantidad === 1 ? 'reporte' : 'reportes'}`;
}

function contarPendientes(cantidad) {
  return `${cantidad} ${cantidad === 1 ? 'pendiente' : 'pendientes'}`;
}

export default function Incidentes() {
  const { usuario } = useSesion();
  const location = useLocation();
  const esCoordinador = usuario?.rol === 'coordinador';
  const esApoderado = usuario?.rol === 'apoderado';

  const [incidentes, setIncidentes] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [recarga, setRecarga] = useState(0);
  const [mensaje, setMensaje] = useState(location.state?.mensaje ?? '');
  const [errorGeneral, setErrorGeneral] = useState('');

  const [revisiones, setRevisiones] = useState({});
  const [erroresRevision, setErroresRevision] = useState({});
  const [revisando, setRevisando] = useState('');
  const [abiertos, setAbiertos] = useState(() => new Set());
  const [tutoresAbiertos, setTutoresAbiertos] = useState(() => new Set());
  const gruposPorTutor = useMemo(() => agruparPorTutor(incidentes), [incidentes]);

  useEffect(() => {
    let vigente = true;
    setCargando(true);
    setErrorGeneral('');
    listarIncidentes()
      .then((respuesta) => {
        if (!vigente) return;
        setIncidentes(respuesta?.incidentes ?? []);
      })
      .catch((error) => {
        if (vigente) setErrorGeneral(error.message);
      })
      .finally(() => {
        if (vigente) setCargando(false);
      });
    return () => {
      vigente = false;
    };
  }, [recarga]);

  function alternarIncidente(incidenteID) {
    setAbiertos((actual) => {
      const siguiente = new Set(actual);
      if (siguiente.has(incidenteID)) siguiente.delete(incidenteID);
      else siguiente.add(incidenteID);
      return siguiente;
    });
  }

  function alternarTutor(clave) {
    setTutoresAbiertos((actual) => {
      const siguiente = new Set(actual);
      if (siguiente.has(clave)) siguiente.delete(clave);
      else siguiente.add(clave);
      return siguiente;
    });
  }

  function cambiarRevision(incidenteID, campo, valor) {
    setRevisiones((actual) => ({
      ...actual,
      [incidenteID]: { ...(actual[incidenteID] || { resultado: 'procede', comentario: '' }), [campo]: valor },
    }));
    setErroresRevision((actual) => ({ ...actual, [incidenteID]: {} }));
  }

  async function enviarRevision(incidenteID) {
    setMensaje('');
    setErrorGeneral('');
    const revision = revisiones[incidenteID] || { resultado: 'procede', comentario: '' };
    const encontrados = validarRevision(revision);
    if (Object.keys(encontrados).length > 0) {
      setErroresRevision((actual) => ({ ...actual, [incidenteID]: encontrados }));
      return;
    }
    setRevisando(incidenteID);
    try {
      await revisarIncidente(incidenteID, revision);
      setMensaje('Revisión registrada.');
      setRecarga((valor) => valor + 1);
    } catch (error) {
      setErroresRevision((actual) => ({ ...actual, [incidenteID]: error.campos || {} }));
      setErrorGeneral(error.message);
    } finally {
      setRevisando('');
    }
  }

  function renderIncidentes(lista) {
    return lista.map((incidente) => {
      const abierto = abiertos.has(incidente.id);
      const detalleID = `incidente-detalle-${incidente.id}`;
      return (
        <li key={incidente.id} className="incidente-tarjeta">
          <button
            type="button"
            className="incidente-desplegable"
            aria-expanded={abierto}
            aria-controls={detalleID}
            onClick={() => alternarIncidente(incidente.id)}
          >
            <span className="incidente-desplegable-texto">
              <strong>{etiquetaTipo(incidente.tipo)}</strong>
              <span className="incidente-meta">
                {esCoordinador || esApoderado ? `Escolar: ${incidente.escolar_nombre}. ` : ''}
                Reportado por: {describirReportante(incidente)}. Tutor: {incidente.tutor_nombre}.
                Curso: {incidente.curso}. Tutoría: {incidente.inicio_lima}.
              </span>
            </span>
            <span className="incidente-desplegable-estado">{describirEstado(incidente.estado)}</span>
            <span className="incidente-desplegable-icono" aria-hidden="true">
              {abierto ? '−' : '+'}
            </span>
          </button>

          {abierto && (
            <div id={detalleID} className="incidente-detalle">
              <p>{incidente.descripcion}</p>
              <Evidencias incidente={incidente} />
              {esCoordinador && incidente.estado === 'pendiente' && (
                <Revision
                  incidente={incidente}
                  revision={revisiones[incidente.id] || { resultado: 'procede', comentario: '' }}
                  errores={erroresRevision[incidente.id] || {}}
                  revisando={revisando === incidente.id}
                  alCambiar={(campo, valor) => cambiarRevision(incidente.id, campo, valor)}
                  alEnviar={() => enviarRevision(incidente.id)}
                />
              )}
              {incidente.estado === 'revisado' && (
                <div className="incidente-revision-final">
                  <strong>{describirResultado(incidente.resultado)}</strong>
                  {incidente.comentario_revision && <p>{incidente.comentario_revision}</p>}
                </div>
              )}
            </div>
          )}
        </li>
      );
    });
  }

  if (cargando) {
    return (
      <main className="pagina-incidentes">
        <h1>Incidentes</h1>
        <p role="status">Cargando reportes…</p>
      </main>
    );
  }

  return (
    <main className="pagina-incidentes">
      <h1>Incidentes</h1>
      {esCoordinador ? (
        <p>
          Revisa los incidentes reportados por escolares y apoderados. La evidencia es
          confidencial y solo la pueden abrir el escolar, su apoderado y coordinación.
        </p>
      ) : (
        <p>
          Consulta los incidentes que registraste y sus evidencias. La evidencia es
          confidencial y solo la pueden abrir el escolar, su apoderado y coordinación.
          El tutor denunciado no tiene acceso a esos archivos.
        </p>
      )}

      {!esCoordinador && (
        <p>
          <Link to="/incidentes/nuevo">Registrar un incidente</Link>
        </p>
      )}

      {mensaje && <p className="incidente-exito" role="status">{mensaje}</p>}
      {errorGeneral && <p className="error" role="alert">{errorGeneral}</p>}

      <section aria-labelledby="titulo-reportes">
        <h2 id="titulo-reportes">{esCoordinador ? 'Reportes recibidos' : 'Mis reportes'}</h2>
        {incidentes.length === 0 ? (
          <p>No hay incidentes registrados.</p>
        ) : esCoordinador ? (
          <div className="incidente-grupos">
            {gruposPorTutor.map((grupo, indice) => {
              const abierto = tutoresAbiertos.has(grupo.clave);
              const detalleID = `incidente-grupo-detalle-${indice}`;
              return (
                <section
                  key={grupo.clave}
                  className="incidente-grupo"
                  aria-labelledby={`incidente-grupo-${indice}`}
                >
                  <h3 id={`incidente-grupo-${indice}`} className="incidente-grupo-titulo">
                    <button
                      type="button"
                      className="incidente-desplegable incidente-grupo-desplegable"
                      aria-expanded={abierto}
                      aria-controls={detalleID}
                      onClick={() => alternarTutor(grupo.clave)}
                    >
                      <span className="incidente-desplegable-texto">
                        <strong>{grupo.tutorNombre}</strong>
                        <span className="incidente-meta">
                          {contarIncidentes(grupo.incidentes.length)} · {contarPendientes(grupo.pendientes)}
                        </span>
                      </span>
                      <span className="incidente-desplegable-icono" aria-hidden="true">
                        {abierto ? '−' : '+'}
                      </span>
                    </button>
                  </h3>
                  {abierto && (
                    <ul id={detalleID} className="incidente-lista">
                      {renderIncidentes(grupo.incidentes)}
                    </ul>
                  )}
                </section>
              );
            })}
          </div>
        ) : (
          <ul className="incidente-lista">{renderIncidentes(incidentes)}</ul>
        )}
      </section>

      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}

function Evidencias({ incidente }) {
  return (
    <div>
      <h4>Evidencias</h4>
      <ul>
        {incidente.evidencias.map((evidencia) => (
          <li key={evidencia.id}>
            <a
              href={evidencia.descargar_en || urlEvidencia(incidente.id, evidencia.id)}
              download={evidencia.nombre}
            >
              {evidencia.nombre}
            </a>{' '}
            <small>({formatearTamano(evidencia.tamano_bytes)})</small>
          </li>
        ))}
      </ul>
    </div>
  );
}

function Revision({ incidente, revision, errores, revisando, alCambiar, alEnviar }) {
  return (
    <fieldset className="incidente-revision">
      <legend>Revisión de coordinación</legend>
      <label className="campo">
        <span>Resultado</span>
        <select
          value={revision.resultado}
          onChange={(evento) => alCambiar('resultado', evento.target.value)}
        >
          <option value="procede">Procede</option>
          <option value="no_procede">No procede</option>
        </select>
      </label>
      {errores.resultado && <p className="error">{errores.resultado}</p>}
      <label className="campo">
        <span>Comentario de revisión</span>
        <textarea
          rows={3}
          maxLength={1000}
          value={revision.comentario}
          onChange={(evento) => alCambiar('comentario', evento.target.value)}
          aria-invalid={Boolean(errores.comentario)}
        />
      </label>
      {errores.comentario && <p className="error">{errores.comentario}</p>}
      <button type="button" onClick={alEnviar} disabled={revisando}>
        {revisando ? 'Registrando…' : `Registrar revisión de ${incidente.escolar_nombre}`}
      </button>
    </fieldset>
  );
}
