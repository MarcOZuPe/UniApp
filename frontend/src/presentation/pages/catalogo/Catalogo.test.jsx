// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Catalogo from './Catalogo.jsx';
import GestionarCatalogo from './GestionarCatalogo.jsx';
import {
  importarCatalogo,
  obtenerCatalogo,
  obtenerCursosPendientes,
  obtenerSolicitudesColegio,
  revisarCurso,
} from '../../../adapters/http/catalogoApi.js';
import { obtenerPerfilesPublicos } from '../../../adapters/http/tutorApi.js';

vi.mock('../../../adapters/http/catalogoApi.js', () => ({
  importarCatalogo: vi.fn(),
  obtenerCatalogo: vi.fn(),
  obtenerCursosPendientes: vi.fn(),
  obtenerSolicitudesColegio: vi.fn(),
  revisarCurso: vi.fn(),
}));
vi.mock('../../../adapters/http/tutorApi.js', () => ({
  obtenerPerfilesPublicos: vi.fn(),
}));
vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => ({ usuario: null }),
}));

beforeEach(() => {
  vi.resetAllMocks();
  obtenerCursosPendientes.mockResolvedValue({ cursos: [] });
  obtenerSolicitudesColegio.mockResolvedValue({ solicitudes: [] });
});

afterEach(cleanup);

describe('interfaz del catálogo', () => {
  it('permite gestionar cursos y colegios con formularios, no expone JSON ni IDs de issue', async () => {
    importarCatalogo.mockResolvedValue({ id: 4, fuente: 'Carga local' });
    render(<MemoryRouter><GestionarCatalogo /></MemoryRouter>);

    expect(await screen.findByRole('heading', { name: 'Administración del catálogo escolar' })).toBeTruthy();
    expect(screen.queryByText(/I07|Datos JSON|JSON válido/)).toBeNull();

    fireEvent.change(screen.getByLabelText('Origen de los datos'), { target: { value: 'Carga local' } });
    fireEvent.change(screen.getByLabelText('Nombre del curso'), { target: { value: 'Álgebra' } });
    fireEvent.change(screen.getByLabelText('Grado'), { target: { value: '1.° de secundaria' } });
    fireEvent.change(screen.getByLabelText('Nombre del colegio'), { target: { value: 'Colegio Central' } });
    fireEvent.change(screen.getByLabelText('Distrito'), { target: { value: 'Lima' } });
    fireEvent.click(screen.getByRole('button', { name: 'Agregar otro curso' }));
    fireEvent.change(screen.getAllByLabelText('Nombre del curso')[1], { target: { value: 'Comunicación' } });
    fireEvent.change(screen.getAllByLabelText('Grado')[1], { target: { value: '2.° de secundaria' } });

    fireEvent.click(screen.getByRole('button', { name: 'Guardar catálogo' }));
    await waitFor(() => expect(importarCatalogo).toHaveBeenCalledWith({
      fuente: 'Carga local',
      cursos: [
        { nombre: 'Álgebra', grado: '1.° de secundaria' },
        { nombre: 'Comunicación', grado: '2.° de secundaria' },
      ],
      colegios: [{ nombre: 'Colegio Central', distrito: 'Lima' }],
    }));
    expect((await screen.findByRole('status')).textContent).toContain(
      'El catálogo se actualizó. Los cursos nuevos quedaron pendientes de revisión.',
    );
    expect(screen.queryByText(/Importación 4/)).toBeNull();
  });

  it('reconoce datos sintéticos con tilde y mayúsculas al importar', async () => {
    importarCatalogo.mockResolvedValue({ id: 5, fuente: 'SINTÉTICO: prueba' });
    render(<MemoryRouter><GestionarCatalogo /></MemoryRouter>);

    fireEvent.change(await screen.findByLabelText('Origen de los datos'), {
      target: { value: 'SINTÉTICO: prueba' },
    });
    fireEvent.change(screen.getByLabelText('Nombre del curso'), { target: { value: 'Álgebra' } });
    fireEvent.change(screen.getByLabelText('Grado'), { target: { value: '1.° de secundaria' } });
    fireEvent.change(screen.getByLabelText('Nombre del colegio'), { target: { value: 'Colegio Central' } });
    fireEvent.change(screen.getByLabelText('Distrito'), { target: { value: 'Lima' } });
    fireEvent.click(screen.getByRole('button', { name: 'Guardar catálogo' }));

    expect((await screen.findByRole('status')).textContent).toContain(
      'Los datos de demostración se guardaron.',
    );
  });

  it('aprueba y rechaza cursos pendientes desde la interfaz', async () => {
    obtenerCursosPendientes.mockResolvedValue({
      cursos: [{ id: 7, nombre: 'Álgebra', grado: '1.° de secundaria', estado: 'pendiente' }],
    });
    revisarCurso.mockResolvedValue({});
    render(<MemoryRouter><GestionarCatalogo /></MemoryRouter>);

    expect(await screen.findByRole('heading', { name: '1.° de secundaria · Álgebra' })).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Aprobar' }));
    await waitFor(() => expect(revisarCurso).toHaveBeenCalledWith(7, { estado: 'aprobado' }));

    const rechazar = screen.getByRole('button', { name: 'Rechazar' });
    expect(rechazar.disabled).toBe(true);
    fireEvent.change(screen.getByLabelText('Motivo del rechazo'), {
      target: { value: 'No corresponde al grado' },
    });
    expect(rechazar.disabled).toBe(false);
    fireEvent.click(rechazar);
    await waitFor(() => expect(revisarCurso).toHaveBeenCalledWith(7, {
      estado: 'rechazado',
      motivo_rechazo: 'No corresponde al grado',
    }));
  });

  it('muestra las solicitudes pendientes sin identidad del solicitante', async () => {
    obtenerSolicitudesColegio.mockResolvedValue({
      solicitudes: [{
        id: 3,
        nombre: 'Colegio Faltante',
        distrito: 'Lima',
        solicitante_id: 'no-debe-mostrarse',
        solicitado_en: '2026-10-23T03:23:27.031195Z',
      }],
    });
    render(<MemoryRouter><GestionarCatalogo /></MemoryRouter>);

    expect(await screen.findByText(/Colegio Faltante · Lima/)).toBeTruthy();
    expect(screen.queryByText(/solicitante/i)).toBeNull();
    expect(screen.queryByText('no-debe-mostrarse')).toBeNull();
  });

  it('explica los datos sintéticos sin mostrar etiquetas internas de fixtures', async () => {
    obtenerCatalogo.mockResolvedValue({
      id: 1,
      fuente: 'SINTÉTICO: fixture inicial I07',
      importado_en: '2026-10-23T03:23:27.031195Z',
      cursos: [],
      colegios: [],
    });
    obtenerPerfilesPublicos.mockResolvedValue({ perfiles: [] });
    render(<MemoryRouter><Catalogo /></MemoryRouter>);

    expect(await screen.findByText(/Datos de demostración; no representan un padrón institucional\./)).toBeTruthy();
    expect(screen.queryByText(/I07|fixture inicial|sintetico:/i)).toBeNull();
  });
});
