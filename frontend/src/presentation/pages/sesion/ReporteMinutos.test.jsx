// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';

import ReporteMinutos from './ReporteMinutos.jsx';

afterEach(cleanup);

const hoy = new Date(2026, 10, 22);
const constancia = {
  id: 'rep-1', generado_por: 'coord', generado_en: '2026-11-22T15:00:00Z',
  desde: '2026-10-01T05:00:00Z', hasta: '2026-11-01T05:00:00Z', tutor_id: '',
  formato: 'csv', filas: 3, sha256: 'a'.repeat(64),
};

function apiFalsa(cambios = {}) {
  return {
    registro: vi.fn(() => Promise.resolve({ reportes: [] })),
    descargar: vi.fn(() => Promise.resolve({ archivo: new Blob(['x']), nombre: 'reporte.csv', id: 'rep-1', sha256: 'b'.repeat(64) })),
    ...cambios,
  };
}

describe('I18: pantalla de reporte de minutos', () => {
  it('propone el mes anterior y descarga el CSV con la huella', async () => {
    const api = apiFalsa();
    const guardar = vi.fn();
    render(<ReporteMinutos api={api} guardarArchivo={guardar} hoy={hoy} />);

    expect(screen.getByLabelText('Desde').value).toBe('2026-10-01');
    expect(screen.getByLabelText('Hasta').value).toBe('2026-10-31');
    expect(screen.getByRole('note').textContent).toMatch(/Sin validez oficial/);

    fireEvent.click(screen.getByRole('button', { name: 'Descargar CSV' }));

    const estado = await screen.findByText(/Se descargó/);
    expect(estado.textContent).toContain('reporte.csv');
    expect(api.descargar).toHaveBeenCalledWith({ desde: '2026-10-01', hasta: '2026-10-31', tutorId: '', formato: 'csv' });
    expect(guardar).toHaveBeenCalledWith(expect.any(Blob), 'reporte.csv');
    await waitFor(() => expect(api.registro).toHaveBeenCalledTimes(2));
  });

  it('el boton PDF pide el formato pdf y limpia el tutor', async () => {
    const api = apiFalsa();
    render(<ReporteMinutos api={api} guardarArchivo={vi.fn()} hoy={hoy} />);
    fireEvent.change(screen.getByLabelText('ID del tutor (opcional)'), { target: { value: '  tutor-a ' } });
    fireEvent.click(screen.getByRole('button', { name: 'Descargar PDF' }));
    await screen.findByText(/Se descargó/);
    expect(api.descargar).toHaveBeenCalledWith(expect.objectContaining({ formato: 'pdf', tutorId: 'tutor-a' }));
  });

  it('valida el periodo antes de pedirlo y enfoca el campo con error', async () => {
    const api = apiFalsa();
    render(<ReporteMinutos api={api} guardarArchivo={vi.fn()} hoy={hoy} />);
    const hasta = screen.getByLabelText('Hasta');
    expect(hasta.getAttribute('aria-invalid')).toBe('false');
    fireEvent.change(hasta, { target: { value: '2026-12-15' } });
    fireEvent.click(screen.getByRole('button', { name: 'Descargar CSV' }));

    expect(await screen.findByText(/como máximo 31 días/)).toBeTruthy();
    expect(hasta.getAttribute('aria-invalid')).toBe('true');
    expect(hasta.getAttribute('aria-describedby')).toMatch(/-error$/);
    expect(document.activeElement).toBe(hasta);
    expect(api.descargar).not.toHaveBeenCalled();
  });

  it('muestra el error del servidor', async () => {
    const api = apiFalsa({ descargar: vi.fn(() => Promise.reject(new Error('No se pudo generar el reporte.'))) });
    render(<ReporteMinutos api={api} guardarArchivo={vi.fn()} hoy={hoy} />);
    fireEvent.click(screen.getByRole('button', { name: 'Descargar CSV' }));
    expect((await screen.findByRole('alert')).textContent).toContain('No se pudo generar el reporte.');
  });

  it('lista las constancias en una tabla navegable con teclado', async () => {
    const api = apiFalsa({ registro: vi.fn(() => Promise.resolve({ reportes: [constancia] })) });
    render(<ReporteMinutos api={api} guardarArchivo={vi.fn()} hoy={hoy} />);
    const region = await screen.findByRole('region', { name: 'Constancias de descarga' });
    expect(region.getAttribute('tabindex')).toBe('0');
    expect(region.textContent).toContain('Todos');
    expect(region.textContent).toContain('CSV');
    expect(screen.getByTitle('a'.repeat(64)).textContent).toBe('aaaaaaaaaaaa…');
  });
});
