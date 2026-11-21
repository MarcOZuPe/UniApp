import { describe, expect, it } from 'vitest';

import { minutosValidos, motivoValido, puedeDisputar } from './minutos.js';

const base = { tutor_id: 't', escolar_id: 'e', minutos_propuestos: 50, minutos_verificados: null };

describe('I17: reglas de presentación de minutos', () => {
  it('valida minutos enteros entre el mínimo y 60', () => {
    expect(minutosValidos(45)).toBe(true);
    expect(minutosValidos('45')).toBe(true);
    expect(minutosValidos(0)).toBe(false);
    expect(minutosValidos(0, 0)).toBe(true);
    expect(minutosValidos(61)).toBe(false);
    expect(minutosValidos(4.5)).toBe(false);
    expect(minutosValidos('')).toBe(false);
  });

  it('exige un motivo no vacío de hasta 500 caracteres', () => {
    expect(motivoValido('  Retraso  ')).toBe(true);
    expect(motivoValido('   ')).toBe(false);
    expect(motivoValido('a'.repeat(501))).toBe(false);
  });

  it('el escolar disputa lo propuesto o verificado', () => {
    expect(puedeDisputar({ ...base, estado: 'propuesto' }, 'e')).toBe(true);
    expect(puedeDisputar({ ...base, estado: 'verificado', minutos_verificados: 50 }, 'e')).toBe(true);
    expect(puedeDisputar({ ...base, estado: 'en_disputa' }, 'e')).toBe(false);
    expect(puedeDisputar({ ...base, estado: 'anulado' }, 'e')).toBe(false);
  });

  it('el tutor solo disputa un ajuste a la baja', () => {
    expect(puedeDisputar({ ...base, estado: 'propuesto' }, 't')).toBe(false);
    expect(puedeDisputar({ ...base, estado: 'verificado', minutos_verificados: 50 }, 't')).toBe(false);
    expect(puedeDisputar({ ...base, estado: 'verificado', minutos_verificados: 30 }, 't')).toBe(true);
  });

  it('nadie más puede disputar', () => {
    expect(puedeDisputar({ ...base, estado: 'propuesto' }, 'otro')).toBe(false);
    expect(puedeDisputar(null, 'e')).toBe(false);
  });
});
