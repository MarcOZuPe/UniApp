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
        {usuario.rol === 'escolar' && <li><Link to="/mi-apoderado">Mi apoderado</Link></li>}
        {usuario.rol === 'tutor' && <li>Próximamente: expediente y disponibilidad (flujos B y C)</li>}
        {usuario.rol === 'apoderado' && <li><Link to="/apoderado/vincular">Vincular a un escolar</Link></li>}
      </ul>
      <button onClick={salir}>Cerrar sesión</button>
    </main>
  );
}
