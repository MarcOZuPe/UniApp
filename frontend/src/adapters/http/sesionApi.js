import { crearClienteHttp } from '../../infrastructure/http/fetchClient.js';

const cliente = crearClienteHttp();

export function obtenerSesion(id) {
  return cliente.get(`/api/sesion/${encodeURIComponent(id)}`);
}

export function declararAsistencia(id) {
  return cliente.post(`/api/sesion/${encodeURIComponent(id)}/asistencias`);
}

export function obtenerSesionPorReserva(reservaID) {
  return cliente.get(
    `/api/sesion/por-reserva/${encodeURIComponent(reservaID)}`
  );
}

export function reportarAusencia(id) {
  return cliente.post(`/api/sesion/${encodeURIComponent(id)}/ausencias`);
}

export function registrarResultado(id, resumen) {
  return cliente.post(`/api/sesion/${encodeURIComponent(id)}/resultado`, { resumen });
}
