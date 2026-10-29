import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import Campo from '../../components/Campo.jsx';
import { restablecer } from '../../../application/acceso/acciones.js';
import { validarRestablecer } from '../../../domain/acceso/validadores.js';

const FORMULARIO_INICIAL = { clave: '', confirmar: '' };

export default function Restablecer() {
  const [parametros] = useSearchParams();
  const token = parametros.get('token') || '';
  const [formulario, setFormulario] = useState(FORMULARIO_INICIAL);
  const [errores, setErrores] = useState({});
  const [mensaje, setMensaje] = useState('');
  const [actualizado, setActualizado] = useState(false);
  const cambiarCampo = (campo) => (evento) =>
    setFormulario({ ...formulario, [campo]: evento.target.value });

  async function enviarFormulario(evento) {
    evento.preventDefault();
    const erroresValidacion = validarRestablecer(formulario);
    setErrores(erroresValidacion);
    setMensaje('');
    if (Object.keys(erroresValidacion).length) return;
    try {
      await restablecer(token, formulario.clave);
      setActualizado(true);
    } catch (error) {
      setErrores(error.campos || {});
      setMensaje(error.message);
    }
  }

  if (!token) return <main><h1>Enlace no válido</h1><p><Link to="/recuperar">Solicita uno nuevo</Link>.</p></main>;
  if (actualizado) return <main><h1>Contraseña actualizada</h1><p><Link to="/ingreso">Ingresar</Link></p></main>;
  return (
    <main>
      <h1>Nueva contraseña</h1>
      <form onSubmit={enviarFormulario} noValidate>
        <Campo etiqueta="Nueva contraseña" type="password" autoComplete="new-password" value={formulario.clave} onChange={cambiarCampo('clave')} error={errores.clave} />
        <Campo etiqueta="Repite la contraseña" type="password" autoComplete="new-password" value={formulario.confirmar} onChange={cambiarCampo('confirmar')} error={errores.confirmar} />
        {mensaje && <p className="error" role="alert">{mensaje}</p>}
        <button>Guardar</button>
      </form>
    </main>
  );
}
