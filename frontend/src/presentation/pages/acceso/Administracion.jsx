import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import {
  crearGeneral,
  eliminarGeneral,
  listarGenerales,
  listarLogsAdministracion,
} from '../../../application/acceso/administracion.js';
import { listarFacultades } from '../../../application/acceso/registro.js';
import '../../styles/panel-coordinacion.css';

const FORMULARIO_INICIAL = {
  email: '',
  nombre_uso: '',
  clave: '',
  facultad_id: '',
};

const ETIQUETAS_CATEGORIA = {
  seguridad: 'Seguridad',
  administracion: 'Administración',
  sistema: 'Sistema',
};

const ETIQUETAS_EVENTO = {
  'clave.cambiada': 'Cambio de contraseña',
  'coordinador_general.creado': 'Coordinador general creado',
  'coordinador_general.eliminado': 'Coordinador general eliminado',
};

export default function Administracion() {
  const [generales, setGenerales] = useState([]);
  const [facultades, setFacultades] = useState([]);
  const [logs, setLogs] = useState([]);
  const [formulario, setFormulario] = useState(FORMULARIO_INICIAL);
  const [errores, setErrores] = useState({});
  const [cargando, setCargando] = useState(true);
  const [enviando, setEnviando] = useState(false);
  const [eliminando, setEliminando] = useState('');
  const [mensaje, setMensaje] = useState('');
  const [errorGeneral, setErrorGeneral] = useState('');

  function aplicarDatos({ generalesRecibidos, facultadesRecibidas, logsRecibidos }) {
    setGenerales(generalesRecibidos);
    setFacultades(facultadesRecibidas);
    setLogs(logsRecibidos);
  }

  async function consultarAdministracion() {
    const [respuestaGenerales, respuestaFacultades, respuestaLogs] = await Promise.all([
      listarGenerales(),
      listarFacultades(),
      listarLogsAdministracion(100),
    ]);
    return {
      generalesRecibidos: respuestaGenerales?.coordinadores_generales ?? [],
      facultadesRecibidas: respuestaFacultades?.facultades ?? [],
      logsRecibidos: respuestaLogs?.logs ?? [],
    };
  }

  async function cargar() {
    aplicarDatos(await consultarAdministracion());
  }

  useEffect(() => {
    let vigente = true;
    consultarAdministracion()
      .then((datos) => {
        if (vigente) aplicarDatos(datos);
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
  }, []);

  function cambiarCampo(campo, valor) {
    setFormulario((actual) => ({ ...actual, [campo]: valor }));
    setErrores((actual) => ({ ...actual, [campo]: '' }));
  }

  async function crear(evento) {
    evento.preventDefault();
    setMensaje('');
    setErrorGeneral('');
    setErrores({});
    setEnviando(true);
    try {
      await crearGeneral(formulario);
      setFormulario(FORMULARIO_INICIAL);
      setMensaje('Coordinador general creado.');
      await cargar();
    } catch (error) {
      setErrores(error.campos || {});
      setErrorGeneral(error.message);
    } finally {
      setEnviando(false);
    }
  }

  async function eliminar(general) {
    if (!window.confirm(`¿Eliminar el acceso de ${general.nombre_uso}?`)) return;
    setMensaje('');
    setErrorGeneral('');
    setEliminando(general.id);
    try {
      await eliminarGeneral(general.id);
      setMensaje('Coordinador general eliminado.');
      await cargar();
    } catch (error) {
      setErrorGeneral(error.message);
    } finally {
      setEliminando('');
    }
  }

  return (
    <main className="pagina-coordinacion">
      <h1>Administración del sistema</h1>
      <p>
        Gestión de coordinadores generales y consulta de la bitácora de seguridad.
      </p>
      <p><Link to="/perfil">Mi perfil</Link> · <Link to="/">Volver al inicio</Link></p>

      {mensaje && <p className="incidente-exito" role="status">{mensaje}</p>}
      {errorGeneral && <p className="error" role="alert">{errorGeneral}</p>}

      <FormularioGeneral
        formulario={formulario}
        facultades={facultades}
        errores={errores}
        enviando={enviando}
        alCambiar={cambiarCampo}
        alEnviar={crear}
      />

      <ListaGenerales
        generales={generales}
        cargando={cargando}
        eliminando={eliminando}
        alEliminar={eliminar}
      />

      <Bitacora logs={logs} />
    </main>
  );
}

function FormularioGeneral({ formulario, facultades, errores, enviando, alCambiar, alEnviar }) {
  return (
    <section aria-labelledby="titulo-nuevo-general">
      <h2 id="titulo-nuevo-general">Nuevo coordinador general</h2>
      <form onSubmit={alEnviar} noValidate>
        <label className="campo">
          <span>Correo</span>
          <input
            type="email"
            autoComplete="email"
            value={formulario.email}
            onChange={(evento) => alCambiar('email', evento.target.value)}
            aria-invalid={Boolean(errores.email)}
          />
        </label>
        {errores.email && <p className="error">{errores.email}</p>}

        <label className="campo">
          <span>Nombre de uso</span>
          <input
            value={formulario.nombre_uso}
            maxLength={60}
            onChange={(evento) => alCambiar('nombre_uso', evento.target.value)}
            aria-invalid={Boolean(errores.nombre_uso)}
          />
        </label>
        {errores.nombre_uso && <p className="error">{errores.nombre_uso}</p>}

        <label className="campo">
          <span>Facultad</span>
          <select
            value={formulario.facultad_id}
            onChange={(evento) => alCambiar('facultad_id', evento.target.value)}
            aria-invalid={Boolean(errores.facultad_id)}
          >
            <option value="">Elige una facultad</option>
            {facultades.map((facultad) => (
              <option key={facultad.id} value={facultad.id}>{facultad.nombre}</option>
            ))}
          </select>
        </label>
        {errores.facultad_id && <p className="error">{errores.facultad_id}</p>}

        <label className="campo">
          <span>Clave inicial</span>
          <input
            type="password"
            autoComplete="new-password"
            value={formulario.clave}
            onChange={(evento) => alCambiar('clave', evento.target.value)}
            aria-invalid={Boolean(errores.clave)}
          />
        </label>
        {errores.clave && <p className="error">{errores.clave}</p>}

        <button type="submit" disabled={enviando}>
          {enviando ? 'Creando…' : 'Crear coordinador general'}
        </button>
      </form>
    </section>
  );
}

function ListaGenerales({ generales, cargando, eliminando, alEliminar }) {
  return (
    <section aria-labelledby="titulo-generales">
      <h2 id="titulo-generales">Coordinadores generales activos</h2>
      {cargando ? (
        <p role="status">Cargando administración…</p>
      ) : generales.length === 0 ? (
        <p>No hay coordinadores generales activos.</p>
      ) : (
        <ul className="panel-coordinadores">
          {generales.map((general) => (
            <li key={general.id}>
              <span>
                <strong>{general.nombre_uso}</strong>
                <small>
                  {general.email}
                  {general.facultad_nombre ? ` · ${general.facultad_nombre}` : ''}
                </small>
              </span>
              <button
                type="button"
                className="boton-peligro"
                disabled={eliminando === general.id}
                onClick={() => alEliminar(general)}
              >
                {eliminando === general.id ? 'Eliminando…' : 'Eliminar'}
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function Bitacora({ logs }) {
  return (
    <section aria-labelledby="titulo-logs">
      <h2 id="titulo-logs">Bitácora de seguridad y administración</h2>
      {logs.length === 0 ? (
        <p>No hay logs registrados.</p>
      ) : (
        <div className="panel-tabla-scroll">
          <table className="panel-auditoria">
            <thead>
              <tr>
                <th scope="col">Fecha</th>
                <th scope="col">Categoría</th>
                <th scope="col">Evento</th>
                <th scope="col">Actor</th>
                <th scope="col">IP</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((log) => (
                <tr key={log.id}>
                  <td>{formatearFecha(log.creado_en)}</td>
                  <td>{ETIQUETAS_CATEGORIA[log.categoria] ?? log.categoria}</td>
                  <td>{ETIQUETAS_EVENTO[log.evento] ?? log.evento}</td>
                  <td>
                    {log.actor_id || 'Sistema'}
                    <small>{log.actor_rol}</small>
                  </td>
                  <td>{log.ip || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function formatearFecha(fecha) {
  if (!fecha) return '—';
  const valor = new Date(fecha);
  if (Number.isNaN(valor.getTime())) return '—';
  return valor.toLocaleString('es-PE', {
    timeZone: 'America/Lima',
    dateStyle: 'short',
    timeStyle: 'short',
  });
}
