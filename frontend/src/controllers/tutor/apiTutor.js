import { api } from '../acceso/acciones.js';

export const obtenerMiExpediente = () => api.get('/api/tutor/mi-expediente');
export const presentarExpediente = (formulario) =>
  api.postFormulario('/api/tutor/expediente', formulario);
export const obtenerPendientes = () => api.get('/api/tutor/revision');
export const revisarExpediente = (tutorId, decision) =>
  api.post(`/api/tutor/${encodeURIComponent(tutorId)}/revision`, decision);
