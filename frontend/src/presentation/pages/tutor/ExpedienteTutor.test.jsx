// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import ExpedienteTutor from './ExpedienteTutor.jsx';
import {
  obtenerMiExpediente,
  presentarExpediente,
} from '../../../adapters/http/tutorApi.js';

vi.mock('../../../adapters/http/tutorApi.js', () => ({
  obtenerMiExpediente: vi.fn(),
  presentarExpediente: vi.fn(),
}));

const documento = {
  id: 'documento-1',
  nombre: 'constancia.pdf',
  tipo: 'application/pdf',
  tamano_bytes: 100,
};

const expedienteBase = {
  tutor_id: 'tutor-1',
  presentacion: 'Presentación del tutor',
  cursos: 'Álgebra',
  experiencia: 'Experiencia del tutor',
  estado: 'aprobado',
  presentado_en: '2026-01-01T00:00:00Z',
  vence_en: '2099-01-01T00:00:00Z',
  documentos: [documento],
};

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(cleanup);

function renderizar() {
  return render(
    <MemoryRouter>
      <ExpedienteTutor />
    </MemoryRouter>
  );
}

describe('expediente del tutor', () => {
  it('muestra el formulario cuando el tutor todavía no presentó un expediente', async () => {
    obtenerMiExpediente.mockRejectedValue({ estado: 404, message: 'No existe' });
    renderizar();

    expect(await screen.findByRole('button', { name: 'Enviar a revisión' })).toBeTruthy();
    expect(screen.getByLabelText('Nombre de uso y breve presentación').value).toBe('');
  });

  it('muestra el estado pendiente y no permite volver a presentar desde la interfaz', async () => {
    obtenerMiExpediente.mockResolvedValue({ ...expedienteBase, estado: 'pendiente', vence_en: undefined });
    renderizar();

    expect(await screen.findByRole('heading', { name: 'Estado: pendiente' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Volver a presentar' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Enviar a revisión' })).toBeNull();
  });

  it('mantiene el formulario oculto mientras la aprobación está vigente', async () => {
    obtenerMiExpediente.mockResolvedValue({
      ...expedienteBase,
      vence_en: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    });
    renderizar();

    expect(await screen.findByText(/Aprobado hasta/)).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Volver a presentar' })).toBeNull();
  });

  it('marca el expediente como vencido y permite volver a presentarlo', async () => {
    obtenerMiExpediente.mockResolvedValue({
      ...expedienteBase,
      vence_en: '2020-01-01T00:00:00Z',
    });
    renderizar();

    expect(await screen.findByRole('heading', { name: 'Estado: vencido' })).toBeTruthy();
    expect(screen.getByText(/Tu expediente venció el/)).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Volver a presentar' })).toBeTruthy();
    expect(screen.getByLabelText('Nombre de uso y breve presentación').value).toBe(
      expedienteBase.presentacion
    );
    expect(screen.getByLabelText('Cursos o temas que puedes enseñar').value).toBe(
      expedienteBase.cursos
    );
    expect(screen.getByLabelText('Experiencia o formación relevante').value).toBe(
      expedienteBase.experiencia
    );
  });
});
