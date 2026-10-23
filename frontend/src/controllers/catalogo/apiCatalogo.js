import { api } from '../acceso/acciones.js';

export const obtenerCatalogo = () => api.get('/api/catalogo');
export const importarCatalogo = (datos) => api.post('/api/catalogo/importaciones', datos);
export const obtenerCursosPendientes = () => api.get('/api/catalogo/cursos/pendientes');
export const revisarCurso = (cursoId, decision) =>
  api.post(`/api/catalogo/cursos/${encodeURIComponent(cursoId)}/revision`, decision);
export const solicitarColegio = (datos) => api.post('/api/catalogo/colegios/solicitudes', datos);
export const obtenerSolicitudesColegio = () => api.get('/api/catalogo/colegios/solicitudes');
