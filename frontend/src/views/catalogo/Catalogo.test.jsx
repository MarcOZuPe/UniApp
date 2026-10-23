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
} from '../../controllers/catalogo/apiCatalogo.js';
import { obtenerPerfilesPublicos } from '../../controllers/tutor/apiTutor.js';

vi.mock('../../controllers/catalogo/apiCatalogo.js', () => ({
  importarCatalogo: vi.fn(),
  obtenerCatalogo: vi.fn(),
  obtenerCursosPendientes: vi.fn(),
  obtenerSolicitudesColegio: vi.fn(),
  revisarCurso: vi.fn(),
}));
vi.mock('../../controllers/tutor/apiTutor.js', () => ({
  obtenerPerfilesPublicos: vi.fn(),
}));
vi.mock('../../controllers/acceso/useSesion.js', () => ({
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

  it('explica los datos sintéticos sin mostrar etiquetas internas de fixtures', async () => {
    obtenerCatalogo.mockResolvedValue({
      id: 1,
      fuente: 'sintetico: fixture inicial I07',
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
