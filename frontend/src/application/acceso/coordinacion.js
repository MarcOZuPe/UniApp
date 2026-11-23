import { api } from '../../adapters/http/apiAcceso.js';

export const obtenerPanelCoordinacion = () => api.get('/api/coordinacion/panel');

export const obtenerMetricasCoordinacion = () => api.get('/api/coordinacion/metricas');

export const listarAuditoriaCoordinacion = (limite = 50) =>
  api.get(`/api/coordinacion/auditoria?limite=${encodeURIComponent(limite)}`);

export const listarCoordinadores = () => api.get('/api/coordinacion/coordinadores');

export const crearCoordinador = (datos) =>
  api.post('/api/coordinacion/coordinadores', datos);

export const eliminarCoordinador = (coordinadorID) =>
  api.delete(`/api/coordinacion/coordinadores/${encodeURIComponent(coordinadorID)}`);
