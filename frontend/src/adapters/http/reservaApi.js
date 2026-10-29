import { ErrorApi } from '../../infrastructure/http/fetchClient.js';

// Cliente del flujo D. Reutiliza el ErrorApi compartido, pero maneja también multipart y
// la cabecera Idempotency-Key, que el cliente de acceso no necesita.
async function interpretar(respuesta) {
  const texto = await respuesta.text();
  if (!texto) return null;
  try {
    return JSON.parse(texto);
  } catch {
    return null;
  }
}

export function crearApiReserva(implementacionFetch = (...argumentos) => fetch(...argumentos)) {
  async function pedir(metodo, ruta, { cuerpo, cabeceras = {} } = {}) {
    const opciones = { method: metodo, credentials: 'same-origin', headers: { ...cabeceras } };
    if (cuerpo instanceof FormData) {
      opciones.body = cuerpo;
    } else if (cuerpo !== undefined) {
      opciones.headers['Content-Type'] = 'application/json';
      opciones.body = JSON.stringify(cuerpo);
    }
    let respuesta;
    try {
      respuesta = await implementacionFetch(ruta, opciones);
    } catch {
      throw new ErrorApi(0, { codigo: 'red', mensaje: 'No hay conexión con el servidor.' });
    }
    const datos = await interpretar(respuesta);
    if (!respuesta.ok) throw new ErrorApi(respuesta.status, datos);
    return datos;
  }

  const ruta = (id, sufijo = '') => `/api/reserva/${encodeURIComponent(id)}${sufijo}`;

  return {
    listarSlots: () => pedir('GET', '/api/reserva/slots'),

    // La clave de idempotencia evita duplicar la reserva si el usuario reintenta.
    crearReserva: (datos, claveIdempotencia) =>
      pedir('POST', '/api/reserva', {
        cuerpo: {
          slot_id: datos.slot_id,
          tema: datos.tema,
          objetivo: datos.objetivo,
          grado: datos.grado,
        },
        cabeceras: claveIdempotencia ? { 'Idempotency-Key': claveIdempotencia } : {},
      }),

    obtenerReserva: (id) => pedir('GET', ruta(id)),

    listarAdjuntos: (id) => pedir('GET', ruta(id, '/adjuntos')),

    // El archivo se sube después de crear la reserva: si falla, la reserva sigue viva.
    subirAdjunto: (id, archivo) => {
      const formulario = new FormData();
      formulario.append('archivo', archivo, archivo.name);
      return pedir('POST', ruta(id, '/adjuntos'), { cuerpo: formulario });
    },

    urlAdjunto: (id, adjuntoId) => ruta(id, `/adjuntos/${encodeURIComponent(adjuntoId)}`),
  };
}

export const apiReserva = crearApiReserva();

export const listarSlots = () => apiReserva.listarSlots();
export const crearReserva = (datos, claveIdempotencia) => apiReserva.crearReserva(datos, claveIdempotencia);
export const obtenerReserva = (id) => apiReserva.obtenerReserva(id);
export const listarAdjuntos = (id) => apiReserva.listarAdjuntos(id);
export const subirAdjunto = (id, archivo) => apiReserva.subirAdjunto(id, archivo);
export const urlAdjunto = (id, adjuntoId) => apiReserva.urlAdjunto(id, adjuntoId);
