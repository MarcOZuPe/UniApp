import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { almacenSesion } from '../../application/acceso/almacenSesion.js';
import { obtenerReserva } from './reservaApi.js';

const respuesta = (estado, cuerpo) => ({
  ok: estado >= 200 && estado < 300,
  status: estado,
  text: async () => JSON.stringify(cuerpo),
});

describe('B-11: el adaptador de reservas debe reaccionar a un 401', () => {
  beforeEach(() => {
    almacenSesion.set({ usuario: { id: 'u-1', rol: 'escolar' }, cargando: false });
  });

  afterEach(() => {
    almacenSesion.set({ usuario: null, cargando: false });
    vi.unstubAllGlobals();
  });

  it('debe cerrar la sesión local cuando la reserva responde 401', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(
      respuesta(401, { codigo: 'no_autenticado', mensaje: 'Debes iniciar sesión.' }),
    ));

    await expect(obtenerReserva('reserva-1')).rejects.toMatchObject({ estado: 401 });

    expect(almacenSesion.getState().usuario).toBeNull();
  });
});
