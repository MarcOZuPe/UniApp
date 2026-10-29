import { describe, expect, it } from 'vitest';
import { validarNuevoSlot } from './ofertaService.js';

describe('controlador de oferta: validación de slots', () => {
  const lunes = { dia_semana: 1, inicio: '09:00', fin: '10:00' };

  it('permite slots contiguos por ser intervalos semiabiertos', () => {
    expect(validarNuevoSlot(
      { dia_semana: 1, inicio: '10:00', fin: '11:00' },
      [lunes],
      180,
    )).toBe('');
  });

  it('rechaza slots solapados el mismo día', () => {
    expect(validarNuevoSlot(
      { dia_semana: 1, inicio: '09:30', fin: '10:30' },
      [lunes],
      180,
    )).toMatch(/solapa/i);
  });

  it('rechaza slots que exceden el cupo semanal', () => {
    expect(validarNuevoSlot(
      { dia_semana: 2, inicio: '10:00', fin: '11:00' },
      [
        lunes,
        { dia_semana: 3, inicio: '09:00', fin: '10:00' },
        { dia_semana: 5, inicio: '09:00', fin: '10:00' },
      ],
      180,
    )).toMatch(/cupo/i);
  });

  it('valida duración y formato de hora', () => {
    expect(validarNuevoSlot(
      { dia_semana: 1, inicio: '09:00', fin: '09:30' },
      [],
      180,
    )).toMatch(/60 minutos/i);
  });
});
