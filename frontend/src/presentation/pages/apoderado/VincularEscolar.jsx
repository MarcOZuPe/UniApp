import { useState } from 'react';
import { Link } from 'react-router-dom';
import Campo from '../../components/Campo.jsx';
import { useFormulario } from '../../hooks/useFormulario.js';
import { solicitarVinculacion } from '../../../adapters/http/apoderadoApi.js';

const FORMULARIO_INICIAL = { email_escolar: '' };

export default function VincularEscolar() {
  const { formulario, actualizarCampo, restablecerFormulario } =
    useFormulario(FORMULARIO_INICIAL);
  const [erroresFormulario, setErroresFormulario] = useState({});
  const [mensajeEstado, setMensajeEstado] = useState('');
  const [enviandoSolicitud, setEnviandoSolicitud] = useState(false);
  const actualizarEmailEscolar = actualizarCampo('email_escolar');

  async function enviarSolicitudVinculacion(evento) {
    evento.preventDefault();
    const emailEscolar = formulario.email_escolar.trim();
    if (!emailEscolar) {
      setErroresFormulario({ email_escolar: 'Ingresa el correo del escolar.' });
      return;
    }

    setErroresFormulario({});
    setMensajeEstado('');
    setEnviandoSolicitud(true);
    try {
      await solicitarVinculacion(emailEscolar);
      restablecerFormulario();
      setMensajeEstado('Solicitud enviada. El escolar recibirá un enlace válido por 24 horas.');
    } catch (errorSolicitud) {
      setMensajeEstado(errorSolicitud.message);
      setErroresFormulario(errorSolicitud.campos || {});
    } finally {
      setEnviandoSolicitud(false);
    }
  }

  return (
    <main>
      <h1>Vincular a un escolar</h1>
      <p>
        Ingresa el correo del escolar. Le enviaremos un enlace para que acepte la
        vinculación. El enlace vence en 24 horas.
      </p>
      <form onSubmit={enviarSolicitudVinculacion} noValidate>
        <Campo
          etiqueta="Correo del escolar"
          type="email"
          autoComplete="email"
          value={formulario.email_escolar}
          onChange={actualizarEmailEscolar}
          error={erroresFormulario.email_escolar}
        />
        {mensajeEstado && <p role="alert">{mensajeEstado}</p>}
        <button disabled={enviandoSolicitud}>
          {enviandoSolicitud ? 'Enviando…' : 'Enviar solicitud'}
        </button>
      </form>

      <p>
        <Link to="/panel_escolares">Ir al panel de escolares</Link>
      </p>
      <p>
        <Link to="/">Volver al inicio</Link>
      </p>
    </main>
  );
}
