import { useState } from 'react';
import { Link } from 'react-router-dom';
import Campo from '../../components/Campo.jsx';
import { useFormulario } from '../../hooks/useFormulario.js';
import { registrar } from '../../../application/acceso/registro.js';
import { reenviarConfirmacion } from '../../../application/acceso/confirmacion.js';
import { ROLES_REGISTRO, etiquetaRol, requiereCorreoInstitucional } from '../../../domain/acceso/roles.js';
import { correosInstitucionales, validarRegistro } from '../../../domain/acceso/validadores.js';

const DOMINIOS_INSTITUCIONALES = import.meta.env?.VITE_TUTOR_EMAIL_DOMAIN || 'uni.edu.pe,uni.pe';

const FORMULARIO_INICIAL = {
  email: '',
  nombre_uso: '',
  clave: '',
  confirmar: '',
  rol: 'escolar',
};

export default function Registro() {
  const { formulario, actualizarCampo } = useFormulario(FORMULARIO_INICIAL);
  const [erroresFormulario, setErroresFormulario] = useState({});
  const [mensajeError, setMensajeError] = useState('');
  const [registroCreado, setRegistroCreado] = useState(false);
  const [avisoReenvio, setAvisoReenvio] = useState('');
  const [enviandoReenvio, setEnviandoReenvio] = useState(false);
  const requiereCorreo = requiereCorreoInstitucional(formulario.rol);
  const ayudaCorreo = requiereCorreo
    ? `Correo institucional (${correosInstitucionales(DOMINIOS_INSTITUCIONALES)})`
    : undefined;

  async function enviarRegistro(evento) {
    evento.preventDefault();
    const erroresValidacion = validarRegistro(formulario, DOMINIOS_INSTITUCIONALES);
    setErroresFormulario(erroresValidacion);
    setMensajeError('');
    if (Object.keys(erroresValidacion).length) return;

    try {
      await registrar(formulario);
      setRegistroCreado(true);
    } catch (errorRegistro) {
      setErroresFormulario(errorRegistro.campos || {});
      setMensajeError(errorRegistro.message);
    }
  }

  async function reenviarCorreo() {
    setEnviandoReenvio(true);
    setAvisoReenvio('');
    try {
      const respuestaReenvio = await reenviarConfirmacion(formulario.email.trim());
      setAvisoReenvio(respuestaReenvio.mensaje);
    } catch (errorReenvio) {
      setAvisoReenvio(errorReenvio.message);
    } finally {
      setEnviandoReenvio(false);
    }
  }

  if (registroCreado) {
    return (
      <PantallaCorreoEnviado
        email={formulario.email}
        reenvio={{ aviso: avisoReenvio, enviando: enviandoReenvio, alReenviar: reenviarCorreo }}
      />
    );
  }

  return (
    <main>
      <h1>Crear cuenta</h1>
      <form onSubmit={enviarRegistro} noValidate>
        <div className="campo">
          <label htmlFor="rol">Soy</label>
          <select id="rol" value={formulario.rol} onChange={actualizarCampo('rol')}>
            {ROLES_REGISTRO.map((rol) => <option key={rol} value={rol}>{etiquetaRol(rol)}</option>)}
          </select>
        </div>
        <Campo
          etiqueta="Nombre de uso"
          autoComplete="nickname"
          value={formulario.nombre_uso}
          onChange={actualizarCampo('nombre_uso')}
          error={erroresFormulario.nombre_uso}
          ayuda="No pedimos DNI ni dirección."
        />
        <Campo
          etiqueta="Correo"
          type="email"
          autoComplete="email"
          value={formulario.email}
          onChange={actualizarCampo('email')}
          error={erroresFormulario.email}
          ayuda={ayudaCorreo}
        />
        <Campo
          etiqueta="Contraseña"
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
        <p><small>Te enviaremos un enlace de confirmación válido por 30 minutos.</small></p>
        <button>Crear cuenta</button>
      </form>
      <p>¿Ya tienes cuenta? <Link to="/ingreso">Ingresar</Link></p>
    </main>
  );
}

function PantallaCorreoEnviado({ email, reenvio }) {
  return (
    <main>
      <h1>Revisa tu correo</h1>
      <p>
        Te enviamos un enlace a <strong>{email}</strong> para confirmar la cuenta.
        Ábrelo para poder ingresar.
      </p>
      <p>
        El enlace vence en <strong>30 minutos</strong>. Si no lo confirmas, el correo
        quedará libre y podrás registrarte de nuevo.
      </p>
      {reenvio.aviso && <p role="status">{reenvio.aviso}</p>}
      <p>
        <button type="button" onClick={reenvio.alReenviar} disabled={reenvio.enviando}>
          {reenvio.enviando ? 'Enviando…' : 'Reenviar enlace'}
        </button>
      </p>
      <p><Link to="/ingreso">Ya confirmé mi correo</Link></p>
    </main>
  );
}
