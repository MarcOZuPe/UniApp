import { crearClienteHttp } from '../../infrastructure/http/fetchClient.js';

// I17: minutos verificados. La identidad y la hora las fija el servidor.
export function crearMinutosApi(cliente = crearClienteHttp()) {
  const sesion = (id) => `/api/sesion/${encodeURIComponent(id)}/minutos`;
  const coordinacion = (id) => `/api/minutos/${encodeURIComponent(id)}`;
  return {
    obtener: (sesionID) => cliente.get(sesion(sesionID)),
    proponer: (sesionID, minutos) => cliente.post(sesion(sesionID), { minutos }),
    disputar: (sesionID, motivo) => cliente.post(`${sesion(sesionID)}/disputa`, { motivo }),
    pendientes: () => cliente.get('/api/minutos/pendientes'),
    aprobar: (sesionID) => cliente.post(`${coordinacion(sesionID)}/aprobar`),
    ajustar: (sesionID, minutos, motivo) => cliente.post(`${coordinacion(sesionID)}/ajustar`, { minutos, motivo }),
    resolver: (sesionID, minutos, motivo) => cliente.post(`${coordinacion(sesionID)}/resolver`, { minutos, motivo }),
    auditoria: (sesionID) => cliente.get(`${coordinacion(sesionID)}/auditoria`),
    totalPropio: () => cliente.get('/api/minutos/total'),
  };
}

export const minutosApi = crearMinutosApi();
