import { afterEach, describe, expect, it, vi } from 'vitest';

import { listarIncidentes } from './incidenteApi.js';

const respuesta = (estado, cuerpo) => ({
  ok: estado >= 200 && estado < 300,
  status: estado,
  text: async () => JSON.stringify(cuerpo),
});

describe('F-16: el listado de incidentes debe poder paginarse', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('debe propagar la página y el límite solicitados', async () => {
    const fetchSimulado = vi.fn().mockResolvedValue(respuesta(200, { incidentes: [] }));
    vi.stubGlobal('fetch', fetchSimulado);

    await listarIncidentes({ pagina: 2, limite: 50 });

    expect(fetchSimulado).toHaveBeenCalledWith(
      '/api/incidentes?pagina=2&limite=50',
      expect.objectContaining({ method: 'GET' }),
    );
  });
});
