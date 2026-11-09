import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { confirmarCorreo } from '../../../application/acceso/confirmacion.js';

const ESTADOS_CONFIRMACION = Object.freeze({
  INICIAL: 'inicial',
  ENVIANDO: 'enviando',
  LISTO: 'listo',
  ERROR: 'error',
});

export default function Confirmar() {
  const [parametros] = useSearchParams();
  const token = parametros.get('token') || '';
  const [estadoConfirmacion, setEstadoConfirmacion] = useState(ESTADOS_CONFIRMACION.INICIAL);
  const [mensajeConfirmacion, setMensajeConfirmacion] = useState('');
  const estaEnviando = estadoConfirmacion === ESTADOS_CONFIRMACION.ENVIANDO;

  async function enviarConfirmacion(evento) {
    evento.preventDefault();
    setEstadoConfirmacion(ESTADOS_CONFIRMACION.ENVIANDO);
    setMensajeConfirmacion('');
    try {
      const respuestaConfirmacion = await confirmarCorreo(token);
      setMensajeConfirmacion(respuestaConfirmacion.mensaje);
      setEstadoConfirmacion(ESTADOS_CONFIRMACION.LISTO);
    } catch (errorConfirmacion) {
      setMensajeConfirmacion(errorConfirmacion.message);
      setEstadoConfirmacion(ESTADOS_CONFIRMACION.ERROR);
    }
  }

  if (!token) {
    return <EnlaceConfirmacionInvalido />;
  }

  if (estadoConfirmacion === ESTADOS_CONFIRMACION.LISTO) {
    return <ConfirmacionExitosa mensaje={mensajeConfirmacion} />;
  }

  return (
    <main>
      <h1>Confirmar correo</h1>
      <p>Pulsa el botón para activar tu cuenta.</p>
      <form onSubmit={enviarConfirmacion}>
        {mensajeConfirmacion && <p className="error" role="alert">{mensajeConfirmacion}</p>}
        <button disabled={estaEnviando}>
          {estaEnviando ? 'Confirmando…' : 'Confirmar mi correo'}
        </button>
      </form>
      <p><Link to="/ingreso">Volver a ingresar</Link></p>
    </main>
  );
}

function EnlaceConfirmacionInvalido() {
  return (
    <main>
      <h1>Enlace no válido</h1>
      <p>Falta el token de confirmación.</p>
      <p><Link to="/registro">Crear cuenta</Link> · <Link to="/ingreso">Ingresar</Link></p>
    </main>
  );
}

function ConfirmacionExitosa({ mensaje }) {
  return (
    <main>
      <h1>Correo confirmado</h1>
      <p role="status">{mensaje}</p>
      <p><Link to="/ingreso">Ingresar</Link></p>
    </main>
  );
}
