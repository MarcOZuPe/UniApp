import { describe, expect, it } from 'vitest';

import { esFuenteSintetica } from './fuente.js';

describe('detección de fuentes sintéticas', () => {
  it('reconoce la etiqueta canónica', () => {
    expect(esFuenteSintetica('sintetico: fixture inicial I07')).toBe(true);
  });

  it('tolera mayúsculas, tildes y espacios exteriores', () => {
    expect(esFuenteSintetica('  SINTÉTICO: prueba  ')).toBe(true);
    expect(esFuenteSintetica('Sintética: muestra')).toBe(true);
  });

  it('no confunde una fuente institucional con datos sintéticos', () => {
    expect(esFuenteSintetica('Padrón institucional 2026')).toBe(false);
    expect(esFuenteSintetica('')).toBe(false);
    expect(esFuenteSintetica(null)).toBe(false);
  });
});
