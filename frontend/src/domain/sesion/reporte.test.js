import { describe, expect, it } from 'vitest';
import { huellaCorta, mesAnterior, validarPeriodo } from './reporte.js';

describe('I18: periodo del reporte', () => {
  it('acepta hasta 31 dias con el dia de fin incluido', () => {
    expect(validarPeriodo({ desde: '2026-10-01', hasta: '2026-10-31' })).toEqual({});
    expect(validarPeriodo({ desde: '2026-10-20', hasta: '2026-10-20' })).toEqual({});
  });

  it('rechaza mas de 31 dias, fin anterior y fechas invalidas', () => {
    expect(validarPeriodo({ desde: '2026-10-01', hasta: '2026-11-01' }).hasta).toMatch(/31 días/);
    expect(validarPeriodo({ desde: '2026-10-10', hasta: '2026-10-09' }).hasta).toMatch(/anterior/);
    expect(validarPeriodo({ desde: '', hasta: '2026-02-30' })).toEqual({
      desde: 'Indica la fecha de inicio.',
      hasta: 'Indica la fecha de fin.',
    });
  });

  it('propone el mes anterior completo', () => {
    expect(mesAnterior(new Date(2026, 10, 22))).toEqual({ desde: '2026-10-01', hasta: '2026-10-31' });
    expect(mesAnterior(new Date(2026, 0, 5))).toEqual({ desde: '2025-12-01', hasta: '2025-12-31' });
  });

  it('abrevia la huella', () => {
    expect(huellaCorta('a'.repeat(64))).toBe('aaaaaaaaaaaa…');
    expect(huellaCorta('')).toBe('');
  });
});
