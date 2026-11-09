import { Link } from 'react-router-dom';
import { salir } from '../../../application/acceso/sesion.js';
import { etiquetaRol, puedeAcceder } from '../../../domain/acceso/roles.js';
import { useSesion } from '../../hooks/useSesion.js';

const ENLACES_INICIO = [
  {
    ruta: '/catalogo',
    texto: 'Explorar cursos y perfiles públicos de tutores',
  },
  {
    ruta: '/notificaciones',
    texto: 'Mis notificaciones',
    roles: ['tutor', 'escolar'],
  },
  {
    ruta: '/sesiones/sesion-i05-demo',
    texto: 'Ver sesión de demostración',
    roles: ['tutor', 'escolar'],
    soloDemo: true,
  },
  {
    ruta: '/coordinacion',
    texto: 'Zona de coordinación',
    roles: ['coordinador'],
  },
  {
    ruta: '/reservar',
    texto: 'Reservar una tutoría',
    roles: ['escolar'],
  },
  {
    ruta: '/agenda',
    texto: 'Mi agenda',
    roles: ['escolar', 'tutor'],
  },
  {
    ruta: '/mi-apoderado',
    texto: 'Mi apoderado',
    roles: ['escolar'],
  },
  {
    ruta: '/expediente',
    texto: 'Mi expediente de tutor',
    roles: ['tutor'],
  },
  {
    ruta: '/tutor/disponibilidad',
    texto: 'Gestionar disponibilidad semanal',
    roles: ['tutor'],
  },
  {
    ruta: '/panel_escolares',
    texto: 'Panel de escolares',
    roles: ['apoderado'],
  },
  {
    ruta: '/apoderado/vincular',
    texto: 'Vincular a un escolar',
    roles: ['apoderado'],
  },
];

export default function Inicio() {
  const { usuario } = useSesion();
  const modoDemoActivo = import.meta.env.VITE_I05_DEMO === 'true';
  const enlacesVisibles = enlacesParaRol(usuario.rol, modoDemoActivo);

  return (
    <main>
      <h1>Hola, {usuario.nombre_uso}</h1>
      <p>Rol: <strong>{etiquetaRol(usuario.rol)}</strong></p>

      <ul>
        {enlacesVisibles.map(({ ruta, texto }) => (
          <li key={ruta}>
            <Link to={ruta}>{texto}</Link>
          </li>
        ))}
      </ul>

      <button onClick={salir}>Cerrar sesión</button>
    </main>
  );
}

function enlacesParaRol(rol, modoDemoActivo) {
  return ENLACES_INICIO.filter(({ roles, soloDemo }) =>
    (!soloDemo || modoDemoActivo) && puedeAcceder(rol, roles)
  );
}
