import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { confirmar } from '../../controllers/acceso/acciones.js';

export default function Confirmar() {
  const [parametros] = useSearchParams();
  const token = parametros.get('token') || '';
  const [estado, setEstado] = useState('inicial');
  const [mensaje, setMensaje] = useState('');

  async function enviarFormulario(evento) {
    evento.preventDefault();
    setEstado('enviando');
    setMensaje('');
    try {
      const respuesta = await confirmar(token);
      setMensaje(respuesta.mensaje);
      setEstado('listo');
    } catch (error) {
      setMensaje(error.message);
      setEstado('error');
    }
  }

  if (!token) {
    return (
      <main>
        <h1>Enlace no válido</h1>
        <p>Falta el token de confirmación.</p>
        <p><Link to="/registro">Crear cuenta</Link> · <Link to="/ingreso">Ingresar</Link></p>
      </main>
    );
  }
  if (estado === 'listo') {
    return (
      <main>
        <h1>Correo confirmado</h1>
        <p role="status">{mensaje}</p>
        <p><Link to="/ingreso">Ingresar</Link></p>
      </main>
    );
  }
  return (
    <main>
      <h1>Confirmar correo</h1>
      <p>Pulsa el botón para activar tu cuenta.</p>
      <form onSubmit={enviarFormulario}>
        {mensaje && <p className="error" role="alert">{mensaje}</p>}
        <button disabled={estado === 'enviando'}>
          {estado === 'enviando' ? 'Confirmando…' : 'Confirmar mi correo'}
        </button>
      </form>
      <p><Link to="/ingreso">Volver a ingresar</Link></p>
    </main>
  );
}
