// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import TutoresAprobados from './TutoresAprobados.jsx';
import { obtenerAprobados } from '../../../adapters/http/tutorApi.js';

vi.mock('../../../adapters/http/tutorApi.js', () => ({
  obtenerAprobados: vi.fn(),
}));

const aprobadoVigente = {
  tutor_id: 'tutor-vigente',
  presentacion: 'Presentación vigente',
  cursos: 'Álgebra',
  experiencia: 'Experiencia vigente',
  estado: 'aprobado',
  presentado_en: '2026-01-01T00:00:00Z',
  vence_en: '2099-01-01T00:00:00Z',
  documentos: [{ id: 'doc-vigente', nombre: 'vigente.pdf' }],
};

const aprobadoVencido = {
  ...aprobadoVigente,
  tutor_id: 'tutor-vencido',
  presentacion: 'Presentación vencida',
  vence_en: '2020-01-01T00:00:00Z',
  documentos: [{ id: 'doc-vencido', nombre: 'vencido.pdf' }],
};

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(cleanup);

function renderizar() {
  return render(
    <MemoryRouter>
      <TutoresAprobados />
    </MemoryRouter>
  );
}

describe('expedientes aprobados para coordinación', () => {
  it('muestra los aprobados vigentes y vencidos con sus documentos', async () => {
    obtenerAprobados.mockResolvedValue({ expedientes: [aprobadoVigente, aprobadoVencido] });
    renderizar();

    expect(await screen.findByRole('heading', { name: 'Expediente tutor-vigente' })).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Expediente tutor-vencido' })).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Expedientes aprobados' })).toBeTruthy();
    expect(screen.getByText('Aprobado, vigente')).toBeTruthy();
    expect(screen.getByText('Aprobado, vencido')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'vigente.pdf' }).getAttribute('href')).toBe(
      '/api/tutor/documentos/doc-vigente'
    );
    expect(screen.getByRole('link', { name: 'vencido.pdf' }).getAttribute('href')).toBe(
      '/api/tutor/documentos/doc-vencido'
    );
  });

  it('informa cuando todavía no hay expedientes aprobados', async () => {
    obtenerAprobados.mockResolvedValue({ expedientes: [] });
    renderizar();

    expect(await screen.findByText('No hay expedientes aprobados.')).toBeTruthy();
  });
});
