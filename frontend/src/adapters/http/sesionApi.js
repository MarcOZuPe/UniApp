import { crearClienteHttp } from '../../infrastructure/http/fetchClient.js';

const cliente = crearClienteHttp();

export function obtenerSesion(id) {
  return cliente.get(`/api/sesion/${encodeURIComponent(id)}`);
}

export function declararAsistencia(id) {
  return cliente.post(`/api/sesion/${encodeURIComponent(id)}/asistencias`);
}
