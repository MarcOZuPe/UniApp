export const TIPOS_INCIDENTE = [
  { valor: 'trato_inadecuado', etiqueta: 'Trato inadecuado' },
  { valor: 'discriminacion', etiqueta: 'Discriminación' },
  { valor: 'acoso', etiqueta: 'Acoso o hostigamiento' },
  { valor: 'seguridad', etiqueta: 'Riesgo de seguridad' },
  { valor: 'otro', etiqueta: 'Otro' },
];

export const MAX_EVIDENCIAS = 5;
export const TAMANO_MAXIMO_EVIDENCIA = 10 * 1024 * 1024;
export const EXTENSIONES_EVIDENCIA = ['.pdf', '.jpg', '.jpeg', '.png'];

export function validarEvidencia(archivo) {
  if (!archivo) return 'Selecciona un archivo.';
  const nombre = String(archivo.name || '').toLowerCase();
  if (!EXTENSIONES_EVIDENCIA.some((extension) => nombre.endsWith(extension))) {
    return 'La evidencia debe ser PDF, JPG o PNG.';
  }
  if (!archivo.size) return 'La evidencia está vacía.';
  if (archivo.size > TAMANO_MAXIMO_EVIDENCIA) return 'Cada evidencia debe pesar como máximo 10 MB.';
  return '';
}

export function validarIncidente({ reserva_id, tipo, descripcion, archivos }) {
  const errores = {};
  if (!reserva_id) {
    errores.reserva_id = 'Elige la tutoría sobre la que reportas el incidente.';
  }
  if (!TIPOS_INCIDENTE.some((opcion) => opcion.valor === tipo)) {
    errores.tipo = 'Elige un tipo de incidente.';
  }
  const longitud = (descripcion || '').trim().length;
  if (longitud < 20 || longitud > 2000) {
    errores.descripcion = 'Describe lo ocurrido entre 20 y 2000 caracteres.';
  }
  const lista = archivos || [];
  if (lista.length === 0) {
    errores.evidencias = 'Adjunta al menos una evidencia del incidente.';
  } else if (lista.length > MAX_EVIDENCIAS) {
    errores.evidencias = `Adjunta como máximo ${MAX_EVIDENCIAS} evidencias.`;
  } else {
    const invalida = lista.map(validarEvidencia).find(Boolean);
    if (invalida) errores.evidencias = invalida;
  }
  return errores;
}

export function validarRevision({ resultado, comentario }) {
  const errores = {};
  if (resultado !== 'procede' && resultado !== 'no_procede') {
    errores.resultado = 'Indica si el incidente procede o no procede.';
  }
  const longitud = (comentario || '').trim().length;
  if (longitud < 5 || longitud > 1000) {
    errores.comentario = 'Explica la revisión entre 5 y 1000 caracteres.';
  }
  return errores;
}

export function formatearTamano(bytes) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export const etiquetaTipo = (valor) =>
  TIPOS_INCIDENTE.find((opcion) => opcion.valor === valor)?.etiqueta ?? valor;
