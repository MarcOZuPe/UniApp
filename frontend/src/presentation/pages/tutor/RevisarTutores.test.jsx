// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../adapters/http/tutorApi.js', () => ({
  obtenerPendientes: vi.fn(),
  revisarExpediente: vi.fn(),
}));

import { obtenerPendientes, revisarExpediente } from '../../../adapters/http/tutorApi.js';
import RevisarTutores from './RevisarTutores.jsx';

const EXPEDIENTE = {
  tutor_id: 't-1', presentacion: 'Presentación', cursos: 'Álgebra', experiencia: 'Experiencia',
  documentos: [{ id: 'd-1', nombre: 'a.pdf' }],
};

beforeEach(() => {
  vi.clearAllMocks();
  obtenerPendientes.mockResolvedValue({ expedientes: [EXPEDIENTE] });
  revisarExpediente.mockResolvedValue({});
});

afterEach(cleanup);

describe('Revisar tutores', () => {
  it('lista los expedientes pendientes', async () => {
    render(
      <MemoryRouter>
        <RevisarTutores />
      </MemoryRouter>
    );
    expect(await screen.findByText('Expediente t-1')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'a.pdf' })).toBeTruthy();
  });

  it('aprueba un expediente con fecha', async () => {
    render(
      <MemoryRouter>
        <RevisarTutores />
      </MemoryRouter>
    );
    await screen.findByText('Expediente t-1');
    fireEvent.change(screen.getByLabelText(/Aprobado hasta/i), { target: { value: '2027-01-01' } });
    fireEvent.click(screen.getByRole('button', { name: 'Aprobar' }));
    await waitFor(() => expect(revisarExpediente).toHaveBeenCalledWith('t-1', expect.objectContaining({ estado: 'aprobado' })));
    expect(obtenerPendientes).toHaveBeenCalledTimes(2);
  });

  it('rechaza un expediente con motivo', async () => {
    render(
      <MemoryRouter>
        <RevisarTutores />
      </MemoryRouter>
    );
    await screen.findByText('Expediente t-1');
    fireEvent.change(screen.getByLabelText(/Motivo del rechazo/i), { target: { value: 'Documento ilegible' } });
    fireEvent.click(screen.getByRole('button', { name: 'Rechazar' }));
    await waitFor(() => expect(revisarExpediente).toHaveBeenCalledWith('t-1', {
      estado: 'rechazado', motivo_rechazo: 'Documento ilegible',
    }));
  });
});
