import { crearClienteHttp } from '../acceso/apiCliente.js';

// El cliente de acceso aporta el manejo de errores y las cookies de sesión.
export function crearApiReserva(implementacionFetch) {
  const cliente = crearClienteHttp(implementacionFetch);
  return {
    listarSlots: () => cliente.get('/api/reserva/slots'),
    crearReserva: (datos) =>
      cliente.post('/api/reserva', {
        slot_id: datos.slot_id,
        tema: datos.tema,
        objetivo: datos.objetivo,
        grado: datos.grado,
        adjuntos: (datos.adjuntos ?? []).map((adjunto) => ({
          nombre: adjunto.nombre,
          tipo: adjunto.tipo,
          tamano_bytes: adjunto.tamano_bytes,
        })),
      }),
    obtenerReserva: (id) => cliente.get(`/api/reserva/${encodeURIComponent(id)}`),
  };
}

export const apiReserva = crearApiReserva();

export const listarSlots = () => apiReserva.listarSlots();
export const crearReserva = (datos) => apiReserva.crearReserva(datos);
export const obtenerReserva = (id) => apiReserva.obtenerReserva(id);
