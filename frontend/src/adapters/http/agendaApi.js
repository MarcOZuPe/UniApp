import { api } from '../../application/acceso/acciones.js';

// Reutilizamos el cliente HTTP autenticado de UniApp.
export async function obtenerAgenda() {
  const respuesta = await api.get('/api/agenda');
  return respuesta?.reservas ?? [];
}
