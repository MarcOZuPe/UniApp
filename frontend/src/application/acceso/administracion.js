import { api } from '../../adapters/http/apiAcceso.js';

export const listarGenerales = () =>
  api.get('/api/administracion/coordinadores-generales');

export const crearGeneral = (datos) =>
  api.post('/api/administracion/coordinadores-generales', datos);

export const eliminarGeneral = (id) =>
  api.delete(`/api/administracion/coordinadores-generales/${encodeURIComponent(id)}`);

export const listarLogsAdministracion = (limite = 100) =>
  api.get(`/api/administracion/logs?limite=${encodeURIComponent(limite)}`);
