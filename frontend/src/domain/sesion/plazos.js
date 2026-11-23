// Estado de un plazo publicado por el backend (fuente única de la regla I11).
// La interfaz solo informa y evita peticiones inútiles; el servidor decide.
export const EstadoPlazo = Object.freeze({
  PENDIENTE: 'pendiente',
  ABIERTO: 'abierto',
  CERRADO: 'cerrado',
});

/**
 * @param {{ desde: string, hasta: string } | undefined} plazo
 * @param {Date} ahora reloj inyectable para pruebas
 */
export function estadoPlazo(plazo, ahora = new Date()) {
  // Sin plazo publicado (API anterior) no se bloquea: el backend valida.
  if (!plazo?.desde || !plazo?.hasta) return EstadoPlazo.ABIERTO;

  const momento = ahora.getTime();
  if (momento < new Date(plazo.desde).getTime()) return EstadoPlazo.PENDIENTE;
  if (momento > new Date(plazo.hasta).getTime()) return EstadoPlazo.CERRADO;
  return EstadoPlazo.ABIERTO;
}

export function plazoAbierto(plazo, ahora = new Date()) {
  return estadoPlazo(plazo, ahora) === EstadoPlazo.ABIERTO;
}
