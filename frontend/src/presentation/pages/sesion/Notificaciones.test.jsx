// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Notificaciones from './Notificaciones.jsx';
import { obtenerNotificaciones } from '../../../adapters/http/notificacionesApi.js';

vi.mock('../../../adapters/http/notificacionesApi.js', () => ({
  obtenerNotificaciones: vi.fn(),
}));

const notificacionBase = {
  id: 'notificacion-1',
  asunto: 'Recordatorio de tutoría',
  cuerpo: 'Tu tutoría empieza pronto.',
};

function renderizar() {
  return render(
    <MemoryRouter>
      <Notificaciones />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(cleanup);

describe('F-21: fechas inválidas sin error boundary', () => {
  it('no desmonta la pantalla cuando una notificación trae una fecha inválida', async () => {
    obtenerNotificaciones.mockResolvedValue([
      { ...notificacionBase, fecha: 'fecha-invalida' },
    ]);

    let errorRender;
    try {
      renderizar();
      expect(await screen.findByText('Recordatorio de tutoría')).toBeTruthy();
    } catch (error) {
      errorRender = error;
    }

    expect(errorRender).toBeUndefined();
  });
});

describe('errores no recuperables en notificaciones', () => {
  it('ofrece reintentar cuando falla la carga', async () => {
    obtenerNotificaciones
      .mockRejectedValueOnce(new Error('Sin conexión.'))
      .mockResolvedValueOnce([notificacionBase]);

    renderizar();

    expect(await screen.findByRole('alert')).toBeTruthy();
    expect(screen.getByRole('button', { name: /reintentar/i })).toBeTruthy();
  });
});
