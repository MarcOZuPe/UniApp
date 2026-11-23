import { describe, expect, it } from 'vitest';
import { EstadoPlazo, estadoPlazo, plazoAbierto } from './plazos.js';

const plazo = { desde: '2026-10-20T14:45:00Z', hasta: '2026-10-21T16:00:00Z' };

describe('I11: estado de plazos', () => {
  it.each([
    ['antes del inicio', '2026-10-20T14:44:59Z', EstadoPlazo.PENDIENTE],
    ['límite inicial', '2026-10-20T14:45:00Z', EstadoPlazo.ABIERTO],
    ['límite final', '2026-10-21T16:00:00Z', EstadoPlazo.ABIERTO],
    ['después del final', '2026-10-21T16:00:01Z', EstadoPlazo.CERRADO],
  ])('%s', (_, momento, esperado) => {
    expect(estadoPlazo(plazo, new Date(momento))).toBe(esperado);
  });

  it('no bloquea cuando la API no publica plazos', () => {
    expect(estadoPlazo(undefined)).toBe(EstadoPlazo.ABIERTO);
    expect(plazoAbierto({ desde: '' })).toBe(true);
  });
});
