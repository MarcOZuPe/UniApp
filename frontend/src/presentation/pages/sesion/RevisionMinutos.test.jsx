// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import RevisionMinutos from './RevisionMinutos.jsx';

vi.mock('../../../adapters/http/minutosApi.js', () => ({
  minutosApi: {},
}));

const pendiente = {
  sesion_id: 'sesion-1',
  estado: 'propuesto',
  minutos_propuestos: 45,
  minutos_verificados: null,
};

function crearApi(extra = {}) {
  return {
    pendientes: vi.fn().mockResolvedValue({ pendientes: [pendiente] }),
    aprobar: vi.fn().mockResolvedValue({}),
    ajustar: vi.fn().mockResolvedValue({}),
    resolver: vi.fn().mockResolvedValue({}),
    auditoria: vi.fn().mockResolvedValue({ auditoria: [] }),
    ...extra,
  };
}

function renderizar(api) {
  return render(
    <MemoryRouter>
      <RevisionMinutos api={api} />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(cleanup);

describe('B-05: aprobación de minutos sin confirmación', () => {
  it('no aprueba minutos con un solo clic sin pedir confirmación', async () => {
    const api = crearApi();
    renderizar(api);

    fireEvent.click(await screen.findByRole('button', { name: 'Aprobar 45 min' }));

    expect(api.aprobar).not.toHaveBeenCalled();
  });
});

describe('errores no recuperables en la bandeja de minutos', () => {
  it('ofrece reintentar cuando falla la carga inicial', async () => {
    const api = crearApi();
    api.pendientes
      .mockRejectedValueOnce(new Error('Sin conexión.'))
      .mockResolvedValueOnce({ pendientes: [] });
    renderizar(api);

    expect((await screen.findByRole('alert')).textContent).toContain('Sin conexión.');

    fireEvent.click(screen.getByRole('button', { name: /reintentar/i }));

    expect(await screen.findByText('No hay minutos pendientes de revisión.')).toBeTruthy();
  });
});
