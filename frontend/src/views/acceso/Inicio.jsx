import { Link } from 'react-router-dom';
import { salir } from '../../controllers/acceso/acciones.js';
import { useSesion } from '../../controllers/acceso/useSesion.js';
import { ROLES } from '../../models/acceso/roles.js';

export default function Inicio() {
  const { usuario } = useSesion();
  return (
    <main>
      <h1>Hola, {usuario.nombre_uso}</h1>
      <p>Rol: <strong>{ROLES[usuario.rol]}</strong></p>
      <ul>
        {['tutor', 'escolar'].includes(usuario.rol) && <li><Link to="/notificaciones">Mis notificaciones</Link></li>}
        {import.meta.env.VITE_I05_DEMO === 'true' && ['tutor', 'escolar'].includes(usuario.rol) && (
          <li><Link to="/sesiones/sesion-i05-demo">Ver sesión de demostración</Link></li>
        )}
        {usuario.rol === 'coordinador' && <li><Link to="/coordinacion">Zona de coordinación</Link></li>}
        {usuario.rol === 'escolar' && <li><Link to="/reservar">Reservar una tutoría</Link></li>}
        {['escolar', 'tutor'].includes(usuario.rol) && ( <li> <Link to="/agenda">Mi agenda</Link></li> )}
        {usuario.rol === 'tutor' && <li><Link to="/expediente">Mi expediente de tutor</Link></li>}
        {usuario.rol === 'tutor' && <li><Link to="/tutor/disponibilidad">Gestionar disponibilidad semanal</Link></li>}
        {usuario.rol === 'apoderado' && <li>Próximamente: autorización del menor (I06)</li>}
      </ul>
      <button onClick={salir}>Cerrar sesión</button>
    </main>
  );
}
