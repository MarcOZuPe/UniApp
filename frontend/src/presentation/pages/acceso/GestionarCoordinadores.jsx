import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import {
  crearCoordinador,
  eliminarCoordinador,
  listarCoordinadores,
} from '../../../application/acceso/coordinacion.js';
import '../../styles/panel-coordinacion.css';

const FORMULARIO_INICIAL = { email: '', nombre_uso: '', clave: '' };

export default function GestionarCoordinadores() {
  const [coordinadores, setCoordinadores] = useState([]);
  const [formulario, setFormulario] = useState(FORMULARIO_INICIAL);
  const [errores, setErrores] = useState({});
  const [cargando, setCargando] = useState(true);
  const [enviando, setEnviando] = useState(false);
  const [eliminando, setEliminando] = useState('');
  const [mensaje, setMensaje] = useState('');
  const [errorGeneral, setErrorGeneral] = useState('');

  async function cargar() {
    setErrorGeneral('');
    try {
      const respuesta = await listarCoordinadores();
      setCoordinadores(respuesta?.coordinadores ?? []);
    } catch (error) {
      setErrorGeneral(error.message);
    } finally {
      setCargando(false);
    }
  }

  useEffect(() => {
    let vigente = true;
    listarCoordinadores()
      .then((respuesta) => {
        if (vigente) setCoordinadores(respuesta?.coordinadores ?? []);
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
      await crearCoordinador(formulario);
      setFormulario(FORMULARIO_INICIAL);
      setMensaje('Coordinador creado.');
      await cargar();
    } catch (error) {
      setErrores(error.campos || {});
      setErrorGeneral(error.message);
    } finally {
      setEnviando(false);
    }
  }

  async function eliminar(coordinador) {
    if (!window.confirm(`¿Eliminar el acceso de ${coordinador.nombre_uso}?`)) return;
    setMensaje('');
    setErrorGeneral('');
    setEliminando(coordinador.id);
    try {
      await eliminarCoordinador(coordinador.id);
      setMensaje('Coordinador eliminado.');
      await cargar();
    } catch (error) {
      setErrorGeneral(error.message);
    } finally {
      setEliminando('');
    }
  }

  return (
    <main className="pagina-coordinacion">
      <h1>Gestionar coordinadores</h1>
      <p>
        Solo la coordinación general puede crear o eliminar cuentas de coordinación.
        Al eliminar una cuenta se revoca su acceso y sus sesiones activas.
      </p>
      <p><Link to="/coordinacion">Volver al panel</Link></p>

      {mensaje && <p className="incidente-exito" role="status">{mensaje}</p>}
      {errorGeneral && <p className="error" role="alert">{errorGeneral}</p>}

      <section aria-labelledby="titulo-nuevo-coordinador">
        <h2 id="titulo-nuevo-coordinador">Nuevo coordinador</h2>
        <form onSubmit={crear} noValidate>
          <label className="campo">
            <span>Correo</span>
            <input
              type="email"
              autoComplete="email"
              value={formulario.email}
              onChange={(evento) => cambiarCampo('email', evento.target.value)}
              aria-invalid={Boolean(errores.email)}
            />
          </label>
          {errores.email && <p className="error">{errores.email}</p>}

          <label className="campo">
            <span>Nombre de uso</span>
            <input
              value={formulario.nombre_uso}
              maxLength={60}
              onChange={(evento) => cambiarCampo('nombre_uso', evento.target.value)}
              aria-invalid={Boolean(errores.nombre_uso)}
            />
          </label>
          {errores.nombre_uso && <p className="error">{errores.nombre_uso}</p>}

          <label className="campo">
            <span>Clave inicial</span>
            <input
              type="password"
              autoComplete="new-password"
              value={formulario.clave}
              onChange={(evento) => cambiarCampo('clave', evento.target.value)}
              aria-invalid={Boolean(errores.clave)}
            />
          </label>
          {errores.clave && <p className="error">{errores.clave}</p>}
          <small>Se aplica la política estricta de claves de coordinación.</small>

          <button type="submit" disabled={enviando}>
            {enviando ? 'Creando…' : 'Crear coordinador'}
          </button>
        </form>
      </section>

      <section aria-labelledby="titulo-lista-coordinadores">
        <h2 id="titulo-lista-coordinadores">Coordinadores activos</h2>
        {cargando ? (
          <p role="status">Cargando coordinadores…</p>
        ) : coordinadores.length === 0 ? (
          <p>No hay coordinadores activos.</p>
        ) : (
          <ul className="panel-coordinadores">
            {coordinadores.map((coordinador) => (
              <li key={coordinador.id}>
                <span>
                  <strong>{coordinador.nombre_uso}</strong>
                  <small>
                    {coordinador.email}
                    {coordinador.facultad_nombre ? ` · ${coordinador.facultad_nombre}` : ''}
                  </small>
                </span>
                <button
                  type="button"
                  className="boton-peligro"
                  disabled={eliminando === coordinador.id}
                  onClick={() => eliminar(coordinador)}
                >
                  {eliminando === coordinador.id ? 'Eliminando…' : 'Eliminar'}
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>

      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}
