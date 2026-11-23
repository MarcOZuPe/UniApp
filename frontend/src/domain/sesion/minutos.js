// I17: reglas de presentacion de minutos. La autoridad es el servidor; esto
// solo decide que acciones ofrecer para no mostrar botones que se rechazaran.

export const MAX_MINUTOS = 60;
export const MAX_MOTIVO = 500;

export const ETIQUETAS_ESTADO = {
  propuesto: 'Propuesto, pendiente de verificación',
  en_disputa: 'En disputa (congelado, suma 0)',
  verificado: 'Verificado',
  anulado: 'Anulado por cancelación (suma 0)',
};

export function minutosValidos(valor, minimo = 1) {
  const numero = Number(valor);
  return Number.isInteger(numero) && numero >= minimo && numero <= MAX_MINUTOS;
}

export function motivoValido(motivo) {
  const limpio = (motivo || '').trim();
  return limpio.length > 0 && Array.from(limpio).length <= MAX_MOTIVO;
}

// El escolar disputa lo propuesto o lo verificado; el tutor solo un
// verificado menor que lo que propuso. Tras una resolucion no hay mas disputas
// (el servidor lo vuelve a comprobar).
export function puedeDisputar(registro, usuarioID) {
  if (!registro || !usuarioID) return false;
  if (usuarioID === registro.escolar_id) {
    return registro.estado === 'propuesto' || registro.estado === 'verificado';
  }
  if (usuarioID === registro.tutor_id) {
    return registro.estado === 'verificado' &&
      registro.minutos_verificados !== null &&
      registro.minutos_verificados < registro.minutos_propuestos;
  }
  return false;
}
