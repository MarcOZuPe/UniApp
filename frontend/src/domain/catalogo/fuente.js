/**
 * Indica si la fuente de una importación se identifica como sintética.
 * Acepta variantes con mayúsculas, tildes y espacios exteriores.
 */
export function esFuenteSintetica(fuente) {
  const normalizada = String(fuente ?? '')
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .trim()
    .toLowerCase();

  return normalizada.startsWith('sintetico') || normalizada.startsWith('sintetica');
}
