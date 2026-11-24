import { puedeAcceder } from '../domain/acceso/roles.js';

const ROLES_COORDINACION = ['coordinador', 'coordinador_general'];

// Accesos de la app por rol. Solo decide qué enlaces se muestran (barra lateral e inicio);
// los permisos reales siguen en App.jsx con RutaProtegida.
export const NAVEGACION = [
  { ruta: '/', texto: 'Inicio', icono: 'inicio', fin: true, soloMenu: true },
  { ruta: '/perfil', texto: 'Mi perfil', descripcion: 'Tus datos de cuenta y contraseña.', icono: 'persona' },
  { ruta: '/agenda', texto: 'Mi agenda', descripcion: 'Tus próximas tutorías y las pasadas.', icono: 'agenda', roles: ['escolar', 'tutor'] },
  { ruta: '/reservar', texto: 'Reservar tutoría', textoInicio: 'Reservar una tutoría', descripcion: 'Elige un horario publicado por un tutor.', icono: 'reservar', roles: ['escolar'] },
  { ruta: '/notificaciones', texto: 'Notificaciones', textoInicio: 'Mis notificaciones', descripcion: 'Avisos y recordatorios de tus sesiones.', icono: 'notificaciones', roles: ['escolar', 'tutor'] },
  { ruta: '/sesiones/sesion-i05-demo', texto: 'Ver sesión de demostración', descripcion: 'Prueba el detalle de una sesión.', icono: 'reloj', roles: ['tutor', 'escolar'], soloDemo: true, soloInicio: true },
  { ruta: '/expediente', texto: 'Mi expediente', textoInicio: 'Mi expediente de tutor', descripcion: 'Tus documentos de verificación como tutor.', icono: 'documento', roles: ['tutor'] },
  { ruta: '/tutor/disponibilidad', texto: 'Disponibilidad', textoInicio: 'Gestionar disponibilidad semanal', descripcion: 'Publica tus horarios semanales.', icono: 'reloj', roles: ['tutor'] },
  { ruta: '/panel_escolares', texto: 'Panel de escolares', descripcion: 'Sigue las tutorías de tus escolares.', icono: 'grupo', roles: ['apoderado'] },
  { ruta: '/apoderado/vincular', texto: 'Vincular escolar', textoInicio: 'Vincular a un escolar', descripcion: 'Envía una invitación para vincular a un escolar.', icono: 'enlace', roles: ['apoderado'] },
  { ruta: '/mi-apoderado', texto: 'Mi apoderado', descripcion: 'Quién te acompaña y qué autorizó.', icono: 'persona', roles: ['escolar'] },
  { ruta: '/coordinacion', texto: 'Coordinación', textoInicio: 'Zona de coordinación', descripcion: 'Resumen del piloto y pendientes.', icono: 'panel', roles: ROLES_COORDINACION, fin: true },
  { ruta: '/coordinacion/tutores', texto: 'Revisar tutores', descripcion: 'Aprueba o rechaza expedientes.', icono: 'verificado', roles: ROLES_COORDINACION },
  { ruta: '/coordinacion/catalogo', texto: 'Gestionar catálogo', descripcion: 'Cursos, grados y colegios del padrón.', icono: 'lista', roles: ROLES_COORDINACION },
  { ruta: '/coordinacion/minutos', texto: 'Verificar minutos', descripcion: 'Aprueba o ajusta minutos verificados.', icono: 'reloj', roles: ROLES_COORDINACION },
  { ruta: '/coordinacion/coordinadores', texto: 'Gestionar coordinadores', descripcion: 'Crea o elimina cuentas de coordinación.', icono: 'grupo', roles: ['coordinador_general'] },
  { ruta: '/administracion', texto: 'Administración', textoInicio: 'Administración del sistema', descripcion: 'Coordinadores generales y bitácora de seguridad.', icono: 'panel', roles: ['administrador'], fin: true },
  { ruta: '/catalogo', texto: 'Catálogo', descripcion: 'Cursos y perfiles públicos de tutores.', icono: 'catalogo' },
  { ruta: '/incidentes', texto: 'Incidentes', textoInicio: 'Incidentes confidenciales', descripcion: 'Reportes confidenciales y su seguimiento.', icono: 'bandera', roles: ['escolar', 'apoderado', ...ROLES_COORDINACION] },
];

const modoDemoActivo = () => import.meta.env.VITE_I05_DEMO === 'true';

export const enlacesDelMenu = (rol) =>
  NAVEGACION.filter(({ roles, soloInicio }) => !soloInicio && puedeAcceder(rol, roles));

export const accesosDeInicio = (rol) =>
  NAVEGACION.filter(({ roles, soloMenu, soloDemo }) =>
    !soloMenu && (!soloDemo || modoDemoActivo()) && puedeAcceder(rol, roles));
