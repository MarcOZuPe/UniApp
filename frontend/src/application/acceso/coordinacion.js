import { api } from '../../adapters/http/apiAcceso.js';

export const obtenerPanelCoordinacion = () => api.get('/api/coordinacion/panel');
