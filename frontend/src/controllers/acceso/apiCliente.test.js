import { describe, expect, it, vi } from 'vitest';
import { crearClienteHttp } from './apiCliente.js';

describe('cliente HTTP', () => {
  it('envía formularios multipart sin fijar Content-Type', async () => {
    const formulario = new FormData();
    formulario.append('presentacion', 'Tutor sintético');
    const fetch = vi.fn(async (_url, opciones) => ({
      ok: true,
      text: async () => '{"estado":"pendiente"}',
      opciones,
    }));
    const api = crearClienteHttp(fetch);

    await expect(api.postFormulario('/api/tutor/expediente', formulario))
      .resolves.toEqual({ estado: 'pendiente' });
    expect(fetch).toHaveBeenCalledWith('/api/tutor/expediente', expect.objectContaining({
      method: 'POST',
      credentials: 'same-origin',
      body: formulario,
    }));
    expect(fetch.mock.calls[0][1].headers).toEqual({});
  });
});
