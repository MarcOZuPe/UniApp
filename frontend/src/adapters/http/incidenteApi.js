import { api } from './apiAcceso.js';

export const listarTutoriasIncidentes = () => api.get('/api/incidentes/tutorias');

export const listarIncidentes = () => api.get('/api/incidentes');

export const crearIncidente = (formulario) => api.postFormulario('/api/incidentes', formulario);

export const revisarIncidente = (incidenteID, revision) =>
  api.post(`/api/incidentes/${encodeURIComponent(incidenteID)}/revision`, revision);

export const urlEvidencia = (incidenteID, evidenciaID) =>
  `/api/incidentes/${encodeURIComponent(incidenteID)}/evidencias/${encodeURIComponent(evidenciaID)}`;
