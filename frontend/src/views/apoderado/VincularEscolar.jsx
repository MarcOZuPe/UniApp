import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import Campo from '../../components/Campo.jsx';
import {
  listarVinculaciones,
  solicitarVinculacion,
} from '../../controllers/apoderado/apiApoderado.js';

const FORMULARIO_INICIAL = { email_escolar: '' };

export default function VincularEscolar() {
  const [formulario, setFormulario] = useState(FORMULARIO_INICIAL);
  const [vinculaciones, setVinculaciones] = useState([]);
  const [errores, setErrores] = useState({});
  const [mensaje, setMensaje] = useState('');
  const [cargando, setCargando] = useState(true);
  const [enviando, setEnviando] = useState(false);

  async function cargarVinculaciones() {
    try {
      const respuesta = await listarVinculaciones();
      setVinculaciones(respuesta.vinculaciones || []);
    } catch (error) {
      setMensaje(error.message);
    } finally {
      setCargando(false);
    }
  }

  useEffect(() => {
    cargarVinculaciones();
  }, []);

  const cambiarCampo = (evento) =>
    setFormulario({ ...formulario, email_escolar: evento.target.value });

  async function enviarFormulario(evento) {
    evento.preventDefault();
    const email = formulario.email_escolar.trim();
    if (!email) {
      setErrores({ email_escolar: 'Ingresa el correo del escolar.' });
      return;
    }
    setErrores({});
    setMensaje('');
    setEnviando(true);
    try {
      await solicitarVinculacion(email);
      setFormulario(FORMULARIO_INICIAL);
      setMensaje('Solicitud enviada. El escolar recibirá un enlace válido por 24 horas.');
      await cargarVinculaciones();
    } catch (error) {
      setMensaje(error.message);
      setErrores(error.campos || {});
    } finally {
      setEnviando(false);
    }
  }

  return (
    <main>
      <h1>Vincular a un escolar</h1>
      <p>
        Ingresa el correo del escolar. Le enviaremos un enlace para que acepte la
        vinculación. El enlace vence en 24 horas.
      </p>
      <form onSubmit={enviarFormulario} noValidate>
        <Campo
          etiqueta="Correo del escolar"
          type="email"
          autoComplete="email"
          value={formulario.email_escolar}
          onChange={cambiarCampo}
          error={errores.email_escolar}
        />
        {mensaje && <p role="alert">{mensaje}</p>}
        <button disabled={enviando}>
          {enviando ? 'Enviando…' : 'Enviar solicitud'}
        </button>
      </form>

      <section aria-labelledby="vinculaciones-titulo">
        <h2 id="vinculaciones-titulo">Mis escolares</h2>
        {cargando && <p role="status">Cargando…</p>}
        {!cargando && vinculaciones.length === 0 && (
          <p>Todavía no has vinculado a ningún escolar.</p>
        )}
        {!cargando && vinculaciones.length > 0 && (
          <ul>
            {vinculaciones.map((vinculacion) => (
              <li key={vinculacion.id}>
                <strong>{vinculacion.escolar.nombre_uso}</strong>{' '}
                <span>({vinculacion.escolar.email})</span>
                <br />
                <small>
                  {vinculacion.estado === 'aceptada'
                    ? 'Vinculación aceptada'
                    : 'Pendiente de aceptación'}
                </small>
              </li>
            ))}
          </ul>
        )}
      </section>

      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}
