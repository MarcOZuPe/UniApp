// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Catalogo from './Catalogo.jsx';
import { obtenerCatalogo } from '../../../adapters/http/catalogoApi.js';
import { obtenerPerfilesPublicos } from '../../../adapters/http/tutorApi.js';

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => ({ usuario: null }),
}));

vi.mock('../../../adapters/http/catalogoApi.js', () => ({
  obtenerCatalogo: vi.fn(),
  solicitarColegio: vi.fn(),
}));

vi.mock('../../../adapters/http/tutorApi.js', () => ({
  obtenerPerfilesPublicos: vi.fn(),
}));

const catalogo = {
  fuente: 'SINTÉTICO: fixture de prueba',
  importado_en: '2026-10-23T03:23:27.031195Z',
  cursos: [{ id: 'curso-1', grado: '1.° de secundaria', nombre: 'Álgebra' }],
  colegios: [{ nombre: 'Colegio Central', distrito: 'Lima' }],
};

function renderizar() {
  return render(
    <MemoryRouter>
      <Catalogo />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  obtenerCatalogo.mockResolvedValue(catalogo);
  obtenerPerfilesPublicos.mockResolvedValue({ perfiles: [] });
});

afterEach(cleanup);

describe('F-20: fallo parcial del catálogo', () => {
  it('mantiene el catálogo visible si fallan los perfiles públicos', async () => {
    obtenerPerfilesPublicos.mockRejectedValue(new Error('Perfiles no disponibles.'));

    renderizar();

    expect((await screen.findAllByText(/Álgebra/)).length).toBeGreaterThan(0);
    expect(screen.getByText(/Colegio Central/)).toBeTruthy();
  });
});

describe('B-22: reintento del catálogo', () => {
  it('ofrece reintentar la carga después de un error inicial', async () => {
    obtenerCatalogo.mockRejectedValueOnce(new Error('Sin conexión.'));
    obtenerCatalogo.mockResolvedValueOnce(catalogo);

    renderizar();

    expect(await screen.findByRole('alert')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /reintentar/i }));

    expect((await screen.findAllByText(/Álgebra/)).length).toBeGreaterThan(0);
    expect(obtenerCatalogo).toHaveBeenCalledTimes(2);
  });
});
