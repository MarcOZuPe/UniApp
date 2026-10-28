// Reglas del flujo D en el navegador: duplican a propósito RN-07 y el límite de adjuntos
// para avisar antes de enviar. La autoridad sigue siendo el modelo Go del servidor.
export const TEMA_MIN = 10;
export const TEMA_MAX = 200;
export const OBJETIVO_MIN = 20;
export const OBJETIVO_MAX = 1000;
export const GRADO_MAX = 80;
export const MAX_ADJUNTOS = 5;
export const ADJUNTO_MAX_BYTES = 10 * 1024 * 1024;
export const ANTICIPACION_MINIMA_MS = 48 * 60 * 60 * 1000;

export const contarCaracteres = (texto) => [...(texto ?? '')].length;

export function validarReserva(datos) {
  const errores = {};
  if (!(datos.slot_id ?? '').trim()) errores.slot_id = 'Elige un horario disponible';

  const tema = (datos.tema ?? '').trim();
  if (contarCaracteres(tema) < TEMA_MIN || contarCaracteres(tema) > TEMA_MAX) {
    errores.tema = 'El tema debe tener entre 10 y 200 caracteres';
  }

  const objetivo = (datos.objetivo ?? '').trim();
  if (contarCaracteres(objetivo) < OBJETIVO_MIN || contarCaracteres(objetivo) > OBJETIVO_MAX) {
    errores.objetivo = 'El objetivo debe tener entre 20 y 1000 caracteres';
  }

  const grado = (datos.grado ?? '').trim();
  if (!grado || contarCaracteres(grado) > GRADO_MAX) {
    errores.grado = 'Indica tu grado (por ejemplo, 5to de secundaria)';
  }

  const adjuntos = datos.adjuntos ?? [];
  if (adjuntos.length > MAX_ADJUNTOS) {
    errores.adjuntos = 'Puedes adjuntar como máximo 5 archivos';
  } else {
    const problema = adjuntos.map(validarAdjunto).find(Boolean);
    if (problema) errores.adjuntos = problema;
  }
  return errores;
}

export function validarAdjunto(adjunto) {
  if (!(adjunto?.nombre ?? '').trim()) return 'Cada adjunto necesita un nombre';
  if (!(adjunto?.tamano_bytes > 0)) return 'El adjunto está vacío';
  if (adjunto.tamano_bytes > ADJUNTO_MAX_BYTES) return 'Cada adjunto debe pesar como máximo 10 MB';
  return '';
}

// El archivo no se sube en I04: se envían sus metadatos y la carga protegida llega en I09.
export const adjuntosDeArchivos = (archivos) =>
  Array.from(archivos ?? []).map((archivo) => ({
    nombre: archivo.name,
    tipo: archivo.type || 'application/octet-stream',
    tamano_bytes: archivo.size,
  }));

// Se usa hora de 24 para que coincida con inicio_lima del servidor (RN-04).
export const formatearEnLima = (instante) =>
  new Intl.DateTimeFormat('es-PE', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
    timeZone: 'America/Lima',
  }).format(new Date(instante));

export const cumpleAnticipacion = (inicio, ahora = new Date()) =>
  new Date(inicio).getTime() - ahora.getTime() >= ANTICIPACION_MINIMA_MS;

export const formatearTamano = (bytes) => {
  if (!bytes) return '0 B';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};
