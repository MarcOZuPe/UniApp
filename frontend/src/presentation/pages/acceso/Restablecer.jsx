import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import Campo from '../../components/Campo.jsx';
import { useFormulario } from '../../hooks/useFormulario.js';
import { restablecerClave } from '../../../application/acceso/recuperacion.js';
import { validarRestablecer } from '../../../domain/acceso/validadores.js';

const FORMULARIO_INICIAL = { clave: '', confirmar: '' };

export default function Restablecer() {
  const [parametros] = useSearchParams();
  const tokenRestablecimiento = parametros.get('token') || '';
  const { formulario, actualizarCampo } = useFormulario(FORMULARIO_INICIAL);
  const [erroresFormulario, setErroresFormulario] = useState({});
  const [mensajeError, setMensajeError] = useState('');
  const [claveActualizada, setClaveActualizada] = useState(false);

  async function enviarRestablecimiento(evento) {
    evento.preventDefault();
    const erroresValidacion = validarRestablecer(formulario);
    setErroresFormulario(erroresValidacion);
    setMensajeError('');
    if (Object.keys(erroresValidacion).length) return;

    try {
      await restablecerClave({
        token: tokenRestablecimiento,
        clave: formulario.clave,
      });
      setClaveActualizada(true);
    } catch (errorRestablecimiento) {
      setErroresFormulario(errorRestablecimiento.campos || {});
      setMensajeError(errorRestablecimiento.message);
    }
  }

  if (!tokenRestablecimiento) {
    return <EnlaceRestablecimientoInvalido />;
  }

  if (claveActualizada) {
    return <ClaveRestablecida />;
  }

  return (
    <main>
      <h1>Nueva contraseña</h1>
      <form onSubmit={enviarRestablecimiento} noValidate>
        <Campo
          etiqueta="Nueva contraseña"
          type="password"
          autoComplete="new-password"
          value={formulario.clave}
          onChange={actualizarCampo('clave')}
          error={erroresFormulario.clave}
        />
        <Campo
          etiqueta="Repite la contraseña"
          type="password"
          autoComplete="new-password"
          value={formulario.confirmar}
          onChange={actualizarCampo('confirmar')}
          error={erroresFormulario.confirmar}
        />
        {mensajeError && <p className="error" role="alert">{mensajeError}</p>}
        <button>Guardar</button>
      </form>
    </main>
  );
}

function EnlaceRestablecimientoInvalido() {
  return (
    <main>
      <h1>Enlace no válido</h1>
      <p><Link to="/recuperar">Solicita uno nuevo</Link>.</p>
    </main>
  );
}

function ClaveRestablecida() {
  return (
    <main>
      <h1>Contraseña actualizada</h1>
      <p><Link to="/ingreso">Ingresar</Link></p>
    </main>
  );
}
