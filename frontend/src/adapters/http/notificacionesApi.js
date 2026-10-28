import { crearClienteHttp } from '../../infrastructure/http/fetchClient.js';
const cliente = crearClienteHttp();
export async function obtenerNotificaciones() {
  const datos = await cliente.get('/api/notificaciones');
  return datos.notificaciones;
}
