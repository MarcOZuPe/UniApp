import { describe, expect, it, vi } from 'vitest';
import { crearApiReserva } from './apiReserva.js';
import {
  adjuntosDeArchivos,
  cumpleAnticipacion,
  formatearEnLima,
  validarReserva,
} from '../../models/reserva/validadores.js';

const respuesta = (estado, cuerpo) => ({
  ok: estado < 400,
  status: estado,
  text: async () => (cuerpo ? JSON.stringify(cuerpo) : ''),
});

const reservaValida = {
  slot_id: 'slot-i04-01',
  tema: 'Repaso de factorización',
  objetivo: 'Resolver ejercicios de factorización de trinomios sin ayuda del tutor.',
  grado: '5to de secundaria',
};

describe('modelo D: formulario de reserva', () => {
  it('acepta una reserva con los datos mínimos y sin adjuntos', () => {
    expect(validarReserva(reservaValida)).toEqual({});
  });

  it('explica qué campo falta o es demasiado corto', () => {
    const errores = validarReserva({ slot_id: '', tema: 'corto', objetivo: 'corto', grado: '  ' });
    expect(Object.keys(errores).sort()).toEqual(['grado', 'objetivo', 'slot_id', 'tema']);
  });

  it('limita a cinco adjuntos y a 10 MB por archivo', () => {
    const seis = Array.from({ length: 6 }, (_, i) => ({ nombre: `p${i}.pdf`, tamano_bytes: 10 }));
    expect(validarReserva({ ...reservaValida, adjuntos: seis }).adjuntos).toBeTruthy();
    const pesado = [{ nombre: 'p.pdf', tamano_bytes: 10 * 1024 * 1024 + 1 }];
    expect(validarReserva({ ...reservaValida, adjuntos: pesado }).adjuntos).toContain('10 MB');
  });

  it('convierte archivos en metadatos sin subir el contenido', () => {
    const archivos = [{ name: 'practica.pdf', type: 'application/pdf', size: 2048 }, { name: 'foto.png', type: '', size: 100 }];
    expect(adjuntosDeArchivos(archivos)).toEqual([
      { nombre: 'practica.pdf', tipo: 'application/pdf', tamano_bytes: 2048 },
      { nombre: 'foto.png', tipo: 'application/octet-stream', tamano_bytes: 100 },
    ]);
  });

  it('exige 48 horas exactas antes del inicio (RN-03)', () => {
    const ahora = new Date('2026-10-20T12:00:00Z');
    expect(cumpleAnticipacion('2026-10-22T12:00:00Z', ahora)).toBe(true);
    expect(cumpleAnticipacion('2026-10-22T11:59:59Z', ahora)).toBe(false);
  });

  it('presenta la hora en America/Lima', () => {
    expect(formatearEnLima('2026-10-23T20:00:00Z')).toContain('15:00');
  });
});

describe('controlador D: cliente de reservas', () => {
  it('pide los horarios al endpoint del flujo D', async () => {
    const fetchSimulado = vi.fn().mockResolvedValue(respuesta(200, { slots: [] }));
    const api = crearApiReserva(fetchSimulado);
    await expect(api.listarSlots()).resolves.toEqual({ slots: [] });
    expect(fetchSimulado).toHaveBeenCalledWith('/api/reserva/slots', expect.objectContaining({ method: 'GET' }));
  });

  it('envía solo los campos del formulario y los metadatos de adjuntos', async () => {
    const fetchSimulado = vi.fn().mockResolvedValue(respuesta(201, { id: 'r-1', estado: 'confirmada' }));
    const api = crearApiReserva(fetchSimulado);
    await api.crearReserva({ ...reservaValida, adjuntos: [{ nombre: 'p.pdf', tipo: 'application/pdf', tamano_bytes: 10, extra: 'x' }] });
    const [ruta, opciones] = fetchSimulado.mock.calls[0];
    expect(ruta).toBe('/api/reserva');
    expect(opciones.method).toBe('POST');
    expect(JSON.parse(opciones.body)).toEqual({
      slot_id: 'slot-i04-01',
      tema: 'Repaso de factorización',
      objetivo: 'Resolver ejercicios de factorización de trinomios sin ayuda del tutor.',
      grado: '5to de secundaria',
      adjuntos: [{ nombre: 'p.pdf', tipo: 'application/pdf', tamano_bytes: 10 }],
    });
  });

  it('propaga el motivo del rechazo de las 48 horas', async () => {
    const fetchSimulado = vi.fn().mockResolvedValue(
      respuesta(422, { codigo: 'anticipacion_insuficiente', mensaje: 'La sesión debe empezar al menos 48 horas después', campos: { slot_id: 'x' } }),
    );
    const api = crearApiReserva(fetchSimulado);
    await expect(api.crearReserva(reservaValida)).rejects.toMatchObject({ estado: 422, codigo: 'anticipacion_insuficiente' });
  });

  it('codifica el identificador al consultar una reserva', async () => {
    const fetchSimulado = vi.fn().mockResolvedValue(respuesta(200, { id: 'r 1' }));
    const api = crearApiReserva(fetchSimulado);
    await api.obtenerReserva('r 1');
    expect(fetchSimulado).toHaveBeenCalledWith('/api/reserva/r%201', expect.any(Object));
  });
});
