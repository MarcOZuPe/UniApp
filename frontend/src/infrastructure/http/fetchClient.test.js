import { describe, expect, it, vi } from 'vitest';

import { crearClienteHttp, ErrorApi } from './fetchClient.js';

describe('F-18: fetch sin timeout', () => {
  it('debe enviar un signal para poder abortar una petición sin respuesta', () => {
    const implementacionFetch = vi.fn(() => new Promise(() => {}));
    const api = crearClienteHttp(implementacionFetch);

    api.get('/api/lento');

    const opciones = implementacionFetch.mock.calls[0][1];
    expect(opciones.signal).toBeInstanceOf(AbortSignal);
  });
});

describe('respuestas exitosas con cuerpo ilegible', () => {
  it('no debe convertir un 200 con JSON inválido en un resultado nulo', async () => {
    const implementacionFetch = vi.fn(async () => ({
      ok: true,
      status: 200,
      text: async () => '<html>respuesta del proxy</html>',
    }));
    const api = crearClienteHttp(implementacionFetch);

    await expect(api.get('/api/datos')).rejects.toBeInstanceOf(ErrorApi);
  });
});
