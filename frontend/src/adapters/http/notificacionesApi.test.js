import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { almacenSesion } from '../../application/acceso/almacenSesion.js';
import { obtenerNotificaciones } from './notificacionesApi.js';

const respuesta = (estado, cuerpo) => ({
  ok: estado >= 200 && estado < 300,
  status: estado,
  text: async () => (cuerpo === undefined ? '' : JSON.stringify(cuerpo)),
});

describe('B-11: el adaptador de notificaciones debe reaccionar a un 401', () => {
  beforeEach(() => {
    almacenSesion.set({ usuario: { id: 'u-1', rol: 'escolar' }, cargando: false });
  });

  afterEach(() => {
    almacenSesion.set({ usuario: null, cargando: false });
    vi.unstubAllGlobals();
  });

  it('debe cerrar la sesión local cuando el backend responde 401', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(
      respuesta(401, { codigo: 'no_autenticado', mensaje: 'Debes iniciar sesión.' }),
    ));

    await expect(obtenerNotificaciones()).rejects.toMatchObject({ estado: 401 });

    expect(almacenSesion.getState().usuario).toBeNull();
  });

  it('debe devolver una lista vacía cuando el cuerpo de la respuesta está vacío', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(respuesta(200, undefined)));

    await expect(obtenerNotificaciones()).resolves.toEqual([]);
  });
});
