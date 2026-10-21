import { useState } from 'react';
import { Link } from 'react-router-dom';
import Campo from '../../components/Campo.jsx';
import { recuperar } from '../../controllers/acceso/acciones.js';
import { validarRecuperar } from '../../models/acceso/validadores.js';

export default function Recuperar() {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [confirmacion, setConfirmacion] = useState('');

  async function enviarFormulario(evento) {
    evento.preventDefault();
    const erroresValidacion = validarRecuperar({ email });
    setError(erroresValidacion.email || '');
    if (erroresValidacion.email) return;
    try {
      const respuesta = await recuperar(email.trim());
      setConfirmacion(respuesta.mensaje);
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <main>
      <h1>Recuperar contraseña</h1>
      {confirmacion ? <p role="status">{confirmacion}</p> : (
        <form onSubmit={enviarFormulario} noValidate>
          <Campo etiqueta="Correo" type="email" autoComplete="email" value={email} onChange={(evento) => setEmail(evento.target.value)} error={error} />
          <button>Enviar enlace</button>
        </form>
      )}
      <p><Link to="/ingreso">Volver a ingresar</Link></p>
    </main>
  );
}
