import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';

import {
  crearIncidente,
  listarIncidentes,
  listarTutoriasIncidentes,
  revisarIncidente,
  urlEvidencia,
} from '../../../adapters/http/incidenteApi.js';
import {
  MAX_EVIDENCIAS,
  TAMANO_MAXIMO_EVIDENCIA,
  TIPOS_INCIDENTE,
  etiquetaTipo,
  formatearTamano,
  validarIncidente,
  validarRevision,
} from '../../../domain/incidente/validadores.js';
import { useSesion } from '../../hooks/useSesion.js';
import '../../styles/incidentes.css';

const FORMULARIO_INICIAL = {
  reserva_id: '',
  tipo: TIPOS_INCIDENTE[0].valor,
  descripcion: '',
};

function describirEstado(estado) {
  return estado === 'revisado' ? 'Revisado por coordinación' : 'Pendiente de revisión';
}

function describirResultado(resultado) {
  if (resultado === 'procede') return 'El incidente procede';
  if (resultado === 'no_procede') return 'El incidente no procede';
  return '';
}

export default function Incidentes() {
  const { usuario } = useSesion();
  const esCoordinador = usuario?.rol === 'coordinador';
  const esApoderado = usuario?.rol === 'apoderado';
  const inputEvidencias = useRef(null);

  const [tutorias, setTutorias] = useState([]);
  const [incidentes, setIncidentes] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [recarga, setRecarga] = useState(0);
  const [mensaje, setMensaje] = useState('');
  const [errorGeneral, setErrorGeneral] = useState('');

  const [formulario, setFormulario] = useState(FORMULARIO_INICIAL);
  const [archivos, setArchivos] = useState([]);
  const [errores, setErrores] = useState({});
  const [enviando, setEnviando] = useState(false);

  const [revisiones, setRevisiones] = useState({});
  const [erroresRevision, setErroresRevision] = useState({});
  const [revisando, setRevisando] = useState('');

  const tutoriaSeleccionada = tutorias.find(
    (tutoria) => tutoria.reserva_id === formulario.reserva_id
  );

  useEffect(() => {
    let vigente = true;
    setCargando(true);
    setErrorGeneral('');
    const consultas = [listarIncidentes()];
    if (!esCoordinador) consultas.push(listarTutoriasIncidentes());
    Promise.all(consultas)
      .then(([respuestaIncidentes, respuestaTutorias]) => {
        if (!vigente) return;
        setIncidentes(respuestaIncidentes?.incidentes ?? []);
        setTutorias(respuestaTutorias?.tutorias ?? []);
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
  }, [esCoordinador, recarga]);

  function cambiarCampo(campo, valor) {
    setFormulario((actual) => ({ ...actual, [campo]: valor }));
    setErrores((actual) => ({ ...actual, [campo]: '' }));
  }

  function cambiarArchivos(evento) {
    const seleccionados = Array.from(evento.target.files ?? []);
    setArchivos(seleccionados);
    setErrores((actual) => ({ ...actual, evidencias: '' }));
  }

  function quitarArchivo(indice) {
    setArchivos((actual) => actual.filter((_, posicion) => posicion !== indice));
  }

  async function enviarReporte(evento) {
    evento.preventDefault();
    setMensaje('');
    setErrorGeneral('');
    const encontrados = validarIncidente({ ...formulario, archivos });
    if (Object.keys(encontrados).length > 0) {
      setErrores(encontrados);
      return;
    }
    const datos = new FormData();
    datos.append('reserva_id', formulario.reserva_id);
    if (esApoderado && tutoriaSeleccionada) {
      datos.append('escolar_id', tutoriaSeleccionada.escolar_id);
    }
    datos.append('tipo', formulario.tipo);
    datos.append('descripcion', formulario.descripcion.trim());
    archivos.forEach((archivo) => datos.append('evidencias', archivo));

    setEnviando(true);
    setErrores({});
    try {
      await crearIncidente(datos);
      setFormulario(FORMULARIO_INICIAL);
      setArchivos([]);
      if (inputEvidencias.current) inputEvidencias.current.value = '';
      setMensaje('Reporte registrado. La evidencia queda disponible solo para ti, tu apoderado y coordinación.');
      setRecarga((valor) => valor + 1);
    } catch (error) {
      setErrores(error.campos || {});
      setErrorGeneral(error.message);
    } finally {
      setEnviando(false);
    }
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
          El tutor denunciado no tiene acceso a esos archivos.
        </p>
      ) : (
        <p>
          Reporta una posible falta del tutor ocurrida en una tutoría que ya atendiste.
          La evidencia es confidencial y solo la pueden abrir el escolar, su apoderado y
          coordinación. El tutor denunciado no tiene acceso a esos archivos.
        </p>
      )}

      {mensaje && <p className="incidente-exito" role="status">{mensaje}</p>}
      {errorGeneral && <p className="error" role="alert">{errorGeneral}</p>}

      {!esCoordinador && (
        <section aria-labelledby="titulo-reporte">
          <h2 id="titulo-reporte">Nuevo reporte</h2>
          {tutorias.length === 0 ? (
            <p>
              Aún no hay tutorías finalizadas disponibles para reportar. Cuando termine
              una sesión, podrás registrar un incidente sobre ella.
            </p>
          ) : (
            <form onSubmit={enviarReporte} noValidate>
              <label className="campo">
                <span>Tutoría atendida</span>
                <select
                  value={formulario.reserva_id}
                  onChange={(evento) => cambiarCampo('reserva_id', evento.target.value)}
                  aria-invalid={Boolean(errores.reserva_id)}
                >
                  <option value="">Elige una tutoría</option>
                  {tutorias.map((tutoria) => (
                    <option key={tutoria.reserva_id} value={tutoria.reserva_id}>
                      {tutoria.inicio_lima} · {tutoria.tutor_nombre} · {tutoria.curso}
                      {esApoderado ? ` · ${tutoria.escolar_nombre}` : ''}
                    </option>
                  ))}
                </select>
              </label>
              {errores.reserva_id && <p className="error">{errores.reserva_id}</p>}

              {tutoriaSeleccionada && (
                <p className="incidente-tutoria">
                  Tutor: <strong>{tutoriaSeleccionada.tutor_nombre}</strong>. Curso:{' '}
                  {tutoriaSeleccionada.curso}. Asistencia declarada:{' '}
                  {tutoriaSeleccionada.asistio ? 'sí' : 'no registrada'}.
                </p>
              )}

              <label className="campo">
                <span>Tipo de incidente</span>
                <select
                  value={formulario.tipo}
                  onChange={(evento) => cambiarCampo('tipo', evento.target.value)}
                  aria-invalid={Boolean(errores.tipo)}
                >
                  {TIPOS_INCIDENTE.map((opcion) => (
                    <option key={opcion.valor} value={opcion.valor}>{opcion.etiqueta}</option>
                  ))}
                </select>
              </label>
              {errores.tipo && <p className="error">{errores.tipo}</p>}

              <label className="campo" htmlFor="incidente-descripcion">
                <span>Descripción de lo ocurrido</span>
                <textarea
                  id="incidente-descripcion"
                  rows={6}
                  maxLength={2000}
                  value={formulario.descripcion}
                  onChange={(evento) => cambiarCampo('descripcion', evento.target.value)}
                  aria-invalid={Boolean(errores.descripcion)}
                />
              </label>
              <small>{formulario.descripcion.trim().length}/2000 caracteres</small>
              {errores.descripcion && <p className="error">{errores.descripcion}</p>}

              <label className="campo" htmlFor="incidente-evidencias">
                <span>Evidencias obligatorias (PDF, JPG o PNG)</span>
                <input
                  id="incidente-evidencias"
                  ref={inputEvidencias}
                  type="file"
                  multiple
                  accept=".pdf,.jpg,.jpeg,.png,application/pdf,image/jpeg,image/png"
                  onChange={cambiarArchivos}
                  aria-invalid={Boolean(errores.evidencias)}
                />
              </label>
              <small>
                De 1 a {MAX_EVIDENCIAS} archivos, hasta 10 MB cada uno. Solo se acepta
                contenido PDF, JPG o PNG.
              </small>
              {errores.evidencias && <p className="error" role="alert">{errores.evidencias}</p>}

              {archivos.length > 0 && (
                <ul className="incidente-archivos">
                  {archivos.map((archivo, indice) => (
                    <li key={`${archivo.name}-${indice}`}>
                      <span>{archivo.name} ({formatearTamano(archivo.size)})</span>
                      <button type="button" onClick={() => quitarArchivo(indice)}>
                        Quitar
                      </button>
                    </li>
                  ))}
                </ul>
              )}
              {archivos.some((archivo) => archivo.size > TAMANO_MAXIMO_EVIDENCIA) && (
                <p className="error">Cada evidencia debe pesar como máximo 10 MB.</p>
              )}

              <button type="submit" disabled={enviando}>
                {enviando ? 'Registrando…' : 'Registrar incidente'}
              </button>
            </form>
          )}
        </section>
      )}

      <section aria-labelledby="titulo-reportes">
        <h2 id="titulo-reportes">{esCoordinador ? 'Reportes recibidos' : 'Mis reportes'}</h2>
        {incidentes.length === 0 ? (
          <p>No hay incidentes registrados.</p>
        ) : (
          <ul className="incidente-lista">
            {incidentes.map((incidente) => (
              <li key={incidente.id} className="incidente-tarjeta">
                <IncidenteResumen incidente={incidente} mostrarEscolar={esCoordinador || esApoderado} />
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
              </li>
            ))}
          </ul>
        )}
      </section>

      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}

function IncidenteResumen({ incidente, mostrarEscolar }) {
  return (
    <div>
      <h3>{etiquetaTipo(incidente.tipo)}</h3>
      <p className="incidente-meta">
        {mostrarEscolar && <>Escolar: {incidente.escolar_nombre}. </>}
        Tutor: {incidente.tutor_nombre}. Curso: {incidente.curso}. Tutoría: {incidente.inicio_lima}.
      </p>
      <p>{incidente.descripcion}</p>
      <p><strong>Estado:</strong> {describirEstado(incidente.estado)}</p>
    </div>
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
