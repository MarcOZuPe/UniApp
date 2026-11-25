import { describe, expect, it } from 'vitest';

import { minutosValidos, motivoValido } from './minutos.js';

describe('validación de minutos: entradas no numéricas', () => {
  it('no debe aceptar una entrada vacía como cero minutos', () => {
    expect(minutosValidos('', 0)).toBe(false);
    expect(minutosValidos('   ', 0)).toBe(false);
  });

  it('no debe aceptar booleanos ni arreglos como minutos', () => {
    expect(minutosValidos(true)).toBe(false);
    expect(minutosValidos([5])).toBe(false);
  });

  it('no debe aceptar un motivo formado solo por espacios de ancho cero', () => {
    expect(motivoValido('\u200b')).toBe(false);
  });
});
