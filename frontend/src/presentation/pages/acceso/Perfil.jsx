import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { cambiarClavePerfil } from '../../../application/acceso/perfil.js';
import { salir } from '../../../application/acceso/sesion.js';
import { etiquetaRol } from '../../../domain/acceso/roles.js';
import { validarCambioClave } from '../../../domain/acceso/validadores.js';
import { useSesion } from '../../hooks/useSesion.js';
import '../../styles/perfil.css';

const FORMULARIO_INICIAL = { clave_actual: '', clave_nueva: '', confirmar: '' };

export default function Perfil() {
  const { usuario } = useSesion();
  const navegar = useNavigate();
  const [formulario, setFormulario] = useState(FORMULARIO_INICIAL);
  const [errores, setErrores] = useState({});
  const [errorGeneral, setErrorGeneral] = useState('');
  const [enviando, setEnviando] = useState(false);

  function cambiarCampo(campo, valor) {
    setFormulario((actual) => ({ ...actual, [campo]: valor }));
    setErrores((actual) => ({ ...actual, [campo]: '' }));
  }

  async function enviar(evento) {
    evento.preventDefault();
    setErrorGeneral('');
    const encontrados = validarCambioClave(formulario, {
      rol: usuario?.rol,
      email: usuario?.email,
    });
    if (Object.keys(encontrados).length > 0) {
      setErrores(encontrados);
      return;
    }
    setEnviando(true);
    try {
      await cambiarClavePerfil({
        clave_actual: formulario.clave_actual,
        clave_nueva: formulario.clave_nueva,
      });
      try {
        await salir();
      } catch {
      }
      navegar('/ingreso', {
        replace: true,
        state: {
          mensaje: 'Contraseña actualizada. Por seguridad, inicia sesión nuevamente.',
        },
      });
    } catch (error) {
      setErrores(error.campos || {});
      setErrorGeneral(error.message);
    } finally {
      setEnviando(false);
    }
  }

  return (
    <main className="pagina-perfil">
      <h1>Mi perfil</h1>
      <p>Administra los datos de tu cuenta y tu contraseña.</p>

      <section aria-labelledby="titulo-datos-cuenta">
        <h2 id="titulo-datos-cuenta">Datos de la cuenta</h2>
        <dl className="perfil-datos">
          <dt>Nombre de identificación</dt>
          <dd>{usuario?.nombre_uso}</dd>
          <dt>Correo</dt>
          <dd>{usuario?.email}</dd>
          <dt>Rol</dt>
          <dd>{etiquetaRol(usuario?.rol)}</dd>
          <dt>Correo confirmado</dt>
          <dd>{usuario?.correo_confirmado ? 'Sí' : 'No'}</dd>
        </dl>
      </section>

      <section aria-labelledby="titulo-clave">
        <h2 id="titulo-clave">Cambiar contraseña</h2>
        {errorGeneral && <p className="error" role="alert">{errorGeneral}</p>}
        <form onSubmit={enviar} noValidate>
          <label className="campo">
            <span>Contraseña actual</span>
            <input
              type="password"
              autoComplete="current-password"
              value={formulario.clave_actual}
              onChange={(evento) => cambiarCampo('clave_actual', evento.target.value)}
              aria-invalid={Boolean(errores.clave_actual)}
            />
          </label>
          {errores.clave_actual && <p className="error">{errores.clave_actual}</p>}

          <label className="campo">
            <span>Nueva contraseña</span>
            <input
              type="password"
              autoComplete="new-password"
              value={formulario.clave_nueva}
              onChange={(evento) => cambiarCampo('clave_nueva', evento.target.value)}
              aria-invalid={Boolean(errores.clave_nueva)}
            />
          </label>
          {errores.clave_nueva && <p className="error">{errores.clave_nueva}</p>}

          <label className="campo">
            <span>Repite la nueva contraseña</span>
            <input
              type="password"
              autoComplete="new-password"
              value={formulario.confirmar}
              onChange={(evento) => cambiarCampo('confirmar', evento.target.value)}
              aria-invalid={Boolean(errores.confirmar)}
            />
          </label>
          {errores.confirmar && <p className="error">{errores.confirmar}</p>}

          <button type="submit" disabled={enviando}>
            {enviando ? 'Guardando…' : 'Cambiar contraseña'}
          </button>
        </form>
      </section>

      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}
