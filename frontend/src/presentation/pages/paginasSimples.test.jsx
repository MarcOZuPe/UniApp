// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../hooks/useSesion.js', () => ({ useSesion: vi.fn() }));
vi.mock('../../adapters/http/apoderadoApi.js', () => ({
  miApoderado: vi.fn(),
  consultarVinculacion: vi.fn(),
  aceptarVinculacion: vi.fn(),
}));
vi.mock('../../adapters/http/tutorApi.js', () => ({ obtenerAprobados: vi.fn() }));
vi.mock('../../application/acceso/confirmacion.js', () => ({ confirmarCorreo: vi.fn() }));

import { useSesion } from '../hooks/useSesion.js';
import {
  aceptarVinculacion,
  consultarVinculacion,
  miApoderado,
} from '../../adapters/http/apoderadoApi.js';
import { obtenerAprobados } from '../../adapters/http/tutorApi.js';
import { confirmarCorreo } from '../../application/acceso/confirmacion.js';
import Inicio from './acceso/Inicio.jsx';
import NoAutorizado from './acceso/NoAutorizado.jsx';
import Confirmar from './acceso/Confirmar.jsx';
import MiApoderado from './apoderado/MiApoderado.jsx';
import AceptarVinculacion from './apoderado/AceptarVinculacion.jsx';
import TutoresAprobados from './tutor/TutoresAprobados.jsx';

beforeEach(() => {
  vi.clearAllMocks();
  useSesion.mockReturnValue({ usuario: { rol: 'escolar', nombre_uso: 'Ana' } });
  miApoderado.mockResolvedValue({ vinculacion: null });
  consultarVinculacion.mockResolvedValue({ vinculacion: null });
  aceptarVinculacion.mockResolvedValue({ vinculacion: null });
  obtenerAprobados.mockResolvedValue({ expedientes: [] });
  confirmarCorreo.mockResolvedValue({ mensaje: 'Cuenta confirmada.' });
});

afterEach(cleanup);

describe('Inicio y NoAutorizado', () => {
  it('muestra el saludo y los accesos del rol', () => {
    render(
      <MemoryRouter>
        <Inicio />
      </MemoryRouter>
    );
    expect(screen.getByText('Hola, Ana')).toBeTruthy();
    expect(screen.getByText('Escolar')).toBeTruthy();
    expect(screen.getByRole('link', { name: /Reservar una tutoría/i })).toBeTruthy();
  });

  it('muestra la pantalla de sin permiso', () => {
    render(
      <MemoryRouter>
        <NoAutorizado />
      </MemoryRouter>
    );
    expect(screen.getByRole('alert').textContent).toContain('no tiene acceso');
  });
});

describe('Mi apoderado', () => {
  it('muestra los datos y la autorización', async () => {
    miApoderado.mockResolvedValue({
      vinculacion: {
        apoderado: { nombre_uso: 'Apoderado Uno', email: 'apoderado@example.test' },
        autorizacion: { autorizado: true, creada_en: '2026-11-01T10:00:00Z' },
      },
    });
    render(
      <MemoryRouter>
        <MiApoderado />
      </MemoryRouter>
    );
    expect(await screen.findByText('Apoderado Uno')).toBeTruthy();
    expect(screen.getByText(/autorizó la creación/i)).toBeTruthy();
  });

  it('muestra el estado sin apoderado', async () => {
    render(
      <MemoryRouter>
        <MiApoderado />
      </MemoryRouter>
    );
    expect(await screen.findByText(/Todavía no tienes un apoderado/i)).toBeTruthy();
  });
});

describe('Aceptar vinculación', () => {
  it('rechaza un enlace sin token', async () => {
    render(
      <MemoryRouter>
        <AceptarVinculacion />
      </MemoryRouter>
    );
    expect(await screen.findByText(/no contiene un token/i)).toBeTruthy();
  });

  it('consulta y acepta la vinculación', async () => {
    consultarVinculacion.mockResolvedValue({
      vinculacion: { apoderado: { nombre_uso: 'Apoderado', email: 'a@example.test' } },
    });
    aceptarVinculacion.mockResolvedValue({
      vinculacion: { apoderado: { nombre_uso: 'Apoderado', email: 'a@example.test' } },
    });
    render(
      <MemoryRouter initialEntries={['/aceptar?token=abc']}>
        <AceptarVinculacion />
      </MemoryRouter>
    );
    fireEvent.click(await screen.findByRole('button', { name: 'Aceptar vinculación' }));
    expect(await screen.findByText(/Vinculación aceptada/i)).toBeTruthy();
    expect(aceptarVinculacion).toHaveBeenCalledWith('abc');
  });
});

describe('Tutores aprobados', () => {
  it('muestra expedientes vigentes y vencidos', async () => {
    obtenerAprobados.mockResolvedValue({
      expedientes: [
        {
          tutor_id: 't-1', presentacion: 'P', cursos: 'C', experiencia: 'E',
          vence_en: '2099-01-01T00:00:00Z', documentos: [{ id: 'd-1', nombre: 'a.pdf' }],
        },
        {
          tutor_id: 't-2', presentacion: 'P2', cursos: 'C2', experiencia: 'E2',
          vence_en: '2020-01-01T00:00:00Z', documentos: [],
        },
      ],
    });
    render(
      <MemoryRouter>
        <TutoresAprobados />
      </MemoryRouter>
    );
    expect(await screen.findByText('Aprobado, vigente')).toBeTruthy();
    expect(screen.getByText('Aprobado, vencido')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'a.pdf' })).toBeTruthy();
  });
});

describe('Confirmar correo', () => {
  it('muestra enlace inválido sin token', () => {
    render(
      <MemoryRouter>
        <Confirmar />
      </MemoryRouter>
    );
    expect(screen.getByText('Falta el token de confirmación.')).toBeTruthy();
  });

  it('confirma el correo con token', async () => {
    render(
      <MemoryRouter initialEntries={['/confirmar?token=abc']}>
        <Confirmar />
      </MemoryRouter>
    );
    fireEvent.click(screen.getByRole('button', { name: 'Confirmar mi correo' }));
    await waitFor(() => expect(confirmarCorreo).toHaveBeenCalledWith('abc'));
    expect(await screen.findByText('Cuenta confirmada.')).toBeTruthy();
  });
});
