import { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import Campo from '../../components/Campo.jsx';
import { useFormulario } from '../../hooks/useFormulario.js';
import { ingresar } from '../../../application/acceso/sesion.js';
import { validarIngreso } from '../../../domain/acceso/validadores.js';

const FORMULARIO_INICIAL = { email: '', clave: '' };

export default function Ingreso() {
  const { formulario, actualizarCampo } = useFormulario(FORMULARIO_INICIAL);
  const [erroresFormulario, setErroresFormulario] = useState({});
  const [mensajeError, setMensajeError] = useState('');
  const [enviandoIngreso, setEnviandoIngreso] = useState(false);
  const navegar = useNavigate();
  const [parametros] = useSearchParams();
  const destinoSolicitado = parametros.get('next');
  const destinoIngreso = obtenerDestinoSeguro(destinoSolicitado);

  async function enviarIngreso(evento) {
    evento.preventDefault();
    const erroresValidacion = validarIngreso(formulario);
    setErroresFormulario(erroresValidacion);
    setMensajeError('');
    if (Object.keys(erroresValidacion).length) return;

    setEnviandoIngreso(true);
    try {
      await ingresar({
        email: formulario.email.trim(),
        clave: formulario.clave,
      });
      navegar(destinoIngreso, { replace: true });
    } catch (errorIngreso) {
      setMensajeError(errorIngreso.message);
    } finally {
      setEnviandoIngreso(false);
    }
  }

  return (
    <main>
      <h1>Ingresar</h1>
      {destinoSolicitado && <p>Inicia sesión para continuar.</p>}
      <form onSubmit={enviarIngreso} noValidate>
        <Campo etiqueta="Correo" type="email" autoComplete="username" value={formulario.email} onChange={actualizarCampo('email')} error={erroresFormulario.email} />
        <Campo etiqueta="Contraseña" type="password" autoComplete="current-password" value={formulario.clave} onChange={actualizarCampo('clave')} error={erroresFormulario.clave} />
        {mensajeError && <p className="error" role="alert">{mensajeError}</p>}
        <button disabled={enviandoIngreso}>
          {enviandoIngreso ? 'Ingresando…' : 'Ingresar'}
        </button>
      </form>
      <p><Link to="/recuperar">¿Olvidaste tu contraseña?</Link> · <Link to="/registro">Crear cuenta</Link></p>
    </main>
  );
}

function obtenerDestinoSeguro(destinoSolicitado) {
  const esDestinoInterno = destinoSolicitado?.startsWith('/') &&
    !destinoSolicitado.startsWith('//') &&
    !destinoSolicitado.startsWith('/\\');
  return esDestinoInterno ? destinoSolicitado : '/';
}
