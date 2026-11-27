// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import GestionarCatalogo from './GestionarCatalogo.jsx';
import {
  importarCatalogo,
  obtenerCursosPendientes,
  obtenerSolicitudesColegio,
  revisarCurso,
} from '../../../adapters/http/catalogoApi.js';

vi.mock('../../../adapters/http/catalogoApi.js', () => ({
  importarCatalogo: vi.fn(),
  obtenerCursosPendientes: vi.fn(),
  obtenerSolicitudesColegio: vi.fn(),
  revisarCurso: vi.fn(),
}));

const cursoPendiente = {
  id: 'curso-7',
  nombre: 'Álgebra',
  grado: '1.° de secundaria',
  estado: 'pendiente',
};

beforeEach(() => {
  vi.resetAllMocks();
  obtenerCursosPendientes.mockResolvedValue({ cursos: [cursoPendiente] });
  obtenerSolicitudesColegio.mockResolvedValue({ solicitudes: [] });
  revisarCurso.mockReturnValue(new Promise(() => {}));
  importarCatalogo.mockResolvedValue({});
});

afterEach(cleanup);

describe('F-25: doble decisión en catálogo', () => {
  it('no envía dos decisiones si se hace doble clic en Aprobar', async () => {
    render(
      <MemoryRouter>
        <GestionarCatalogo />
      </MemoryRouter>
    );

    const aprobar = await screen.findByRole('button', { name: 'Aprobar' });
    fireEvent.click(aprobar);
    fireEvent.click(aprobar);

    expect(revisarCurso).toHaveBeenCalledTimes(1);
  });
});
