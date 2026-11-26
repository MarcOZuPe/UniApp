import { afterEach, describe, expect, it, vi } from 'vitest';

import { obtenerAgenda } from './agendaApi.js';

const respuesta = (estado, cuerpo) => ({
  ok: estado >= 200 && estado < 300,
  status: estado,
  text: async () => JSON.stringify(cuerpo),
});

describe('F-16: la agenda debe poder paginarse', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('debe propagar la página y el límite solicitados', async () => {
    const fetchSimulado = vi.fn().mockResolvedValue(respuesta(200, { reservas: [] }));
    vi.stubGlobal('fetch', fetchSimulado);

    await obtenerAgenda({ pagina: 2, limite: 50 });

    expect(fetchSimulado).toHaveBeenCalledWith(
      '/api/agenda?pagina=2&limite=50',
      expect.objectContaining({ method: 'GET' }),
    );
  });
});
