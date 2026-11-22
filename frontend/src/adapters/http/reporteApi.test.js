import { describe, expect, it, vi } from 'vitest';
import { crearReporteApi } from './reporteApi.js';

function respuesta({ ok = true, status = 200, headers = {}, cuerpo = '' }) {
  return {
    ok,
    status,
    headers: { get: (clave) => headers[clave] ?? null },
    blob: () => Promise.resolve(new Blob([cuerpo])),
    json: () => Promise.resolve(JSON.parse(cuerpo)),
  };
}

describe('I18: adaptador del reporte', () => {
  it('pide el archivo con los filtros y lee nombre y huella de las cabeceras', async () => {
    const fetch = vi.fn(() => Promise.resolve(respuesta({
      headers: {
        'Content-Disposition': 'attachment; filename="reporte-minutos-20261001-20261031.csv"',
        'X-Reporte-Id': 'rep-1',
        'X-Reporte-Sha256': 'f'.repeat(64),
      },
      cuerpo: 'a,b',
    })));
    const api = crearReporteApi(fetch);
    const resultado = await api.descargar({ desde: '2026-10-01', hasta: '2026-10-31', formato: 'csv', tutorId: 'tutor 1' });

    expect(fetch.mock.calls[0][0]).toBe('/api/reportes/minutos?desde=2026-10-01&hasta=2026-10-31&formato=csv&tutor_id=tutor+1');
    expect(fetch.mock.calls[0][1]).toEqual({ credentials: 'same-origin' });
    expect(resultado).toMatchObject({ nombre: 'reporte-minutos-20261001-20261031.csv', id: 'rep-1', sha256: 'f'.repeat(64) });
    expect(await resultado.archivo.text()).toBe('a,b');
  });

  it('omite tutor_id vacio y convierte los errores del servidor', async () => {
    const fetch = vi.fn(() => Promise.resolve(respuesta({
      ok: false, status: 422, cuerpo: JSON.stringify({ codigo: 'periodo_invalido', mensaje: 'Periodo invalido.' }),
    })));
    const api = crearReporteApi(fetch);
    await expect(api.descargar({ desde: '2026-10-01', hasta: '2026-12-01', formato: 'pdf', tutorId: '' }))
      .rejects.toMatchObject({ estado: 422, codigo: 'periodo_invalido', message: 'Periodo invalido.' });
    expect(fetch.mock.calls[0][0]).not.toContain('tutor_id');
  });

  it('informa la falta de conexion', async () => {
    const api = crearReporteApi(() => Promise.reject(new TypeError('Failed to fetch')));
    await expect(api.descargar({ desde: 'a', hasta: 'b', formato: 'csv' })).rejects.toMatchObject({ estado: 0, codigo: 'red' });
  });
});
