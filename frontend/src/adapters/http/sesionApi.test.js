import {
  afterEach,
  describe,
  expect,
  it,
  vi,
} from 'vitest';

import {
  obtenerSesion,
  declararAsistencia,
} from './sesionApi.js';

afterEach(() => {
  vi.unstubAllGlobals();
});

function simularRespuesta(estado, datos) {
  return {
    ok: estado >= 200 && estado < 300,
    status: estado,
    text: async () => JSON.stringify(datos),
  };
}

describe('I05: cliente API de sesiones', () => {
  it('consulta la sesion mediante GET', async () => {
    const sesion = {
      id: 'sesion-001',
      reserva_id: 'reserva-001',
      asistencias: [],
    };

    const fetchSimulado = vi.fn().mockResolvedValue(
      simularRespuesta(200, sesion)
    );

    vi.stubGlobal('fetch', fetchSimulado);

    const resultado = await obtenerSesion('sesion-001');

    expect(resultado).toEqual(sesion);

    expect(fetchSimulado).toHaveBeenCalledWith(
      '/api/sesion/sesion-001',
      expect.objectContaining({
        method: 'GET',
        credentials: 'same-origin',
      })
    );
  });

  it('declara asistencia mediante POST sin enviar usuario_id', async () => {
    const asistencia = {
      usuario_id: 'usuario-001',
      rol: 'tutor',
      declarada_en: '2026-10-20T15:00:00Z',
    };

    const fetchSimulado = vi.fn().mockResolvedValue(
      simularRespuesta(201, asistencia)
    );

    vi.stubGlobal('fetch', fetchSimulado);

    const resultado = await declararAsistencia('sesion-001');

    expect(resultado).toEqual(asistencia);

    expect(fetchSimulado).toHaveBeenCalledWith(
      '/api/sesion/sesion-001/asistencias',
      expect.objectContaining({
        method: 'POST',
        credentials: 'same-origin',
        body: undefined,
      })
    );
  });

  it('codifica correctamente identificadores en la URL', async () => {
    const fetchSimulado = vi.fn().mockResolvedValue(
      simularRespuesta(200, { id: 'sesion de prueba' })
    );

    vi.stubGlobal('fetch', fetchSimulado);

    await obtenerSesion('sesion de prueba');

    expect(fetchSimulado).toHaveBeenCalledWith(
      '/api/sesion/sesion%20de%20prueba',
      expect.any(Object)
    );
  });

  it('propaga el error cuando la sesion no existe', async () => {
    const fetchSimulado = vi.fn().mockResolvedValue(
      simularRespuesta(404, {
        codigo: 'sesion_no_encontrada',
        mensaje: 'Sesion no encontrada',
      })
    );

    vi.stubGlobal('fetch', fetchSimulado);

    await expect(
      obtenerSesion('inexistente')
    ).rejects.toMatchObject({
      estado: 404,
      codigo: 'sesion_no_encontrada',
    });
  });

  it('propaga el error cuando la asistencia esta duplicada', async () => {
    const fetchSimulado = vi.fn().mockResolvedValue(
      simularRespuesta(409, {
        codigo: 'asistencia_duplicada',
        mensaje: 'Asistencia ya declarada',
      })
    );

    vi.stubGlobal('fetch', fetchSimulado);

    await expect(
      declararAsistencia('sesion-001')
    ).rejects.toMatchObject({
      estado: 409,
      codigo: 'asistencia_duplicada',
    });
  });
});
