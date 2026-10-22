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
        {usuario.rol === 'coordinador' && <li><Link to="/coordinacion">Zona de coordinación</Link></li>}
        {usuario.rol === 'escolar' && <li><Link to="/reservar">Reservar una tutoría</Link></li>}
        {usuario.rol === 'tutor' && <li><Link to="/expediente">Mi expediente de tutor</Link></li>}
        {usuario.rol === 'tutor' && <li><Link to="/tutor/disponibilidad">Gestionar disponibilidad semanal</Link></li>}
        {usuario.rol === 'apoderado' && <li>Próximamente: autorización del menor (I06)</li>}
      </ul>
      <button onClick={salir}>Cerrar sesión</button>
    </main>
  );
}
