import { crearClienteHttp, ErrorApi } from '../../infrastructure/http/fetchClient.js';

// I18: reporte de minutos. La descarga devuelve un archivo (no JSON), por eso
// usa fetch directo; los errores llegan como JSON igual que el resto de la API.
export function crearReporteApi(
  implementacionFetch = (...argumentos) => fetch(...argumentos),
  cliente = crearClienteHttp(implementacionFetch),
) {
  async function descargar({ desde, hasta, formato, tutorId }) {
    const parametros = new URLSearchParams({ desde, hasta, formato });
    if (tutorId) parametros.set('tutor_id', tutorId);

    let respuesta;
    try {
      respuesta = await implementacionFetch(`/api/reportes/minutos?${parametros}`, { credentials: 'same-origin' });
    } catch {
      throw new ErrorApi(0, { codigo: 'red', mensaje: 'No hay conexión con el servidor.' });
    }
    if (!respuesta.ok) {
      let cuerpo = null;
      try { cuerpo = await respuesta.json(); } catch { /* sin cuerpo JSON */ }
      throw new ErrorApi(respuesta.status, cuerpo);
    }
    const disposicion = respuesta.headers.get('Content-Disposition') || '';
    const nombre = /filename="([^"]+)"/.exec(disposicion)?.[1] || `reporte-minutos.${formato}`;
    return {
      archivo: await respuesta.blob(),
      nombre,
      id: respuesta.headers.get('X-Reporte-Id') || '',
      sha256: respuesta.headers.get('X-Reporte-Sha256') || '',
    };
  }

  return {
    descargar,
    registro: () => cliente.get('/api/reportes/registro'),
  };
}

export const reporteApi = crearReporteApi();
