import { useState } from 'react';
import { Link } from 'react-router-dom';
import Campo from '../../components/Campo.jsx';
import { useFormulario } from '../../hooks/useFormulario.js';
import { solicitarRecuperacion } from '../../../application/acceso/recuperacion.js';
import { validarRecuperar } from '../../../domain/acceso/validadores.js';

const FORMULARIO_INICIAL = { email: '' };

export default function Recuperar() {
  const { formulario, actualizarCampo } = useFormulario(FORMULARIO_INICIAL);
  const [errorCorreo, setErrorCorreo] = useState('');
  const [mensajeConfirmacion, setMensajeConfirmacion] = useState('');

  async function enviarSolicitud(evento) {
    evento.preventDefault();
    const erroresFormulario = validarRecuperar(formulario);
    setErrorCorreo(erroresFormulario.email || '');
    if (erroresFormulario.email) return;

    try {
      const respuestaSolicitud = await solicitarRecuperacion(formulario.email.trim());
      setMensajeConfirmacion(respuestaSolicitud.mensaje);
    } catch (errorSolicitud) {
      setErrorCorreo(errorSolicitud.message);
    }
  }

  return (
    <main>
      <h1>Recuperar contraseña</h1>
      {mensajeConfirmacion ? <p role="status">{mensajeConfirmacion}</p> : (
        <form onSubmit={enviarSolicitud} noValidate>
          <Campo
            etiqueta="Correo"
            type="email"
            autoComplete="email"
            value={formulario.email}
            onChange={actualizarCampo('email')}
            error={errorCorreo}
          />
          <button>Enviar enlace</button>
        </form>
      )}
      <p><Link to="/ingreso">Volver a ingresar</Link></p>
    </main>
  );
}
