// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import SeccionMinutos from './SeccionMinutos.jsx';
import RevisionMinutos from './RevisionMinutos.jsx';

afterEach(cleanup);

const sinMinutos = () => Promise.reject(Object.assign(new Error('sin minutos'), { estado: 404 }));
const sesionTerminada = { id: 'sesion-1', tutor_id: 't', escolar_id: 'e', fin: '2026-11-10T16:00:00Z' };
const registro = (cambios = {}) => ({
  sesion_id: 'sesion-1', tutor_id: 't', escolar_id: 'e', estado: 'propuesto',
  minutos_propuestos: 45, minutos_verificados: null, minutos_computables: 0, ...cambios,
});

describe('I17: sección de minutos en el detalle de sesión', () => {
  it('el tutor propone los minutos dictados y no suman hasta verificarse', async () => {
    const api = { obtener: vi.fn(sinMinutos), proponer: vi.fn(() => Promise.resolve(registro())) };
    render(<SeccionMinutos sesion={sesionTerminada} usuario={{ id: 't' }} api={api} />);

    const campo = await screen.findByLabelText('Minutos realmente dictados');
    const boton = screen.getByRole('button', { name: 'Proponer minutos' });
    fireEvent.change(campo, { target: { value: '61' } });
    expect(boton.disabled).toBe(true);
    fireEvent.change(campo, { target: { value: '45' } });
    fireEvent.click(boton);

    await screen.findByText(/Coordinación los revisará/);
    expect(api.proponer).toHaveBeenCalledWith('sesion-1', 45);
    expect(screen.getByText(/Suman al total:/).parentElement.textContent).toContain('0 min');
  });

  it('el escolar no propone y puede disputar con motivo', async () => {
    const api = {
      obtener: vi.fn(() => Promise.resolve(registro({ estado: 'verificado', minutos_verificados: 45, minutos_computables: 45 }))),
      disputar: vi.fn(() => Promise.resolve(registro({ estado: 'en_disputa', minutos_verificados: 45 }))),
    };
    render(<SeccionMinutos sesion={sesionTerminada} usuario={{ id: 'e' }} api={api} />);

    const boton = await screen.findByRole('button', { name: 'Disputar minutos' });
    expect(screen.queryByLabelText('Minutos realmente dictados')).toBeNull();
    expect(boton.disabled).toBe(true);
    fireEvent.change(screen.getByLabelText('Motivo de la disputa'), { target: { value: ' Duró 30 minutos ' } });
    fireEvent.click(boton);

    await screen.findByText(/congelados/);
    expect(api.disputar).toHaveBeenCalledWith('sesion-1', 'Duró 30 minutos');
    expect(screen.getByText(/En disputa/)).toBeTruthy();
  });

  it('el tutor no ve disputa si no hubo ajuste a la baja', async () => {
    const api = { obtener: vi.fn(() => Promise.resolve(registro({ estado: 'verificado', minutos_verificados: 45, minutos_computables: 45 }))) };
    render(<SeccionMinutos sesion={sesionTerminada} usuario={{ id: 't' }} api={api} />);
    await screen.findByText(/Verificados por coordinación/);
    expect(screen.queryByRole('button', { name: 'Disputar minutos' })).toBeNull();
  });

  it('muestra el rechazo del servidor sin anunciar éxito', async () => {
    const api = {
      obtener: vi.fn(sinMinutos),
      proponer: vi.fn(() => Promise.reject(new Error('tutor y escolar deben haber declarado asistencia'))),
    };
    render(<SeccionMinutos sesion={sesionTerminada} usuario={{ id: 't' }} api={api} />);
    fireEvent.change(await screen.findByLabelText('Minutos realmente dictados'), { target: { value: '30' } });
    fireEvent.click(screen.getByRole('button', { name: 'Proponer minutos' }));
    expect((await screen.findByRole('alert')).textContent).toContain('declarado asistencia');
    expect(screen.queryByRole('status')).toBeNull();
  });
});

describe('I17: bandeja de revisión de coordinación', () => {
  function abrir(api) {
    return render(<MemoryRouter><RevisionMinutos api={api} /></MemoryRouter>);
  }

  it('aprueba de forma explícita y recarga la bandeja', async () => {
    const api = {
      pendientes: vi.fn()
        .mockResolvedValueOnce({ pendientes: [registro()] })
        .mockResolvedValueOnce({ pendientes: [] }),
      aprobar: vi.fn(() => Promise.resolve(registro({ estado: 'verificado' }))),
    };
    abrir(api);
    fireEvent.click(await screen.findByRole('button', { name: 'Aprobar 45 min' }));
    await screen.findByText('No hay minutos pendientes de revisión.');
    expect(api.aprobar).toHaveBeenCalledWith('sesion-1');
  });

  it('un ajuste exige minutos y motivo', async () => {
    const api = {
      pendientes: vi.fn(() => Promise.resolve({ pendientes: [registro()] })),
      ajustar: vi.fn(() => Promise.resolve(registro({ estado: 'verificado' }))),
    };
    abrir(api);
    const boton = await screen.findByRole('button', { name: 'Ajustar minutos' });
    fireEvent.change(screen.getByLabelText('Ajustar a'), { target: { value: '40' } });
    expect(boton.disabled).toBe(true);
    fireEvent.change(screen.getByLabelText('Motivo (queda en la auditoría)'), { target: { value: 'Retraso' } });
    fireEvent.click(boton);
    await waitFor(() => expect(api.ajustar).toHaveBeenCalledWith('sesion-1', 40, 'Retraso'));
  });

  it('una disputa se resuelve (no se aprueba) y muestra la auditoría', async () => {
    const api = {
      pendientes: vi.fn(() => Promise.resolve({ pendientes: [registro({ estado: 'en_disputa', minutos_verificados: 45 })] })),
      resolver: vi.fn(() => Promise.resolve(registro({ estado: 'verificado' }))),
      auditoria: vi.fn(() => Promise.resolve({ auditoria: [
        { accion: 'proponer', actor_id: 't', valor_nuevo: 45, motivo: '' },
        { accion: 'disputar', actor_id: 'e', valor_nuevo: null, motivo: 'Duró 30' },
      ] })),
    };
    abrir(api);
    await screen.findByRole('button', { name: 'Resolver disputa' });
    expect(screen.queryByRole('button', { name: /Aprobar/ })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Ver auditoría' }));
    const lista = await screen.findByRole('list', { name: 'Auditoría de sesion-1' });
    expect(lista.textContent).toContain('disputar por e (Duró 30)');
    fireEvent.change(screen.getByLabelText('Minutos finales'), { target: { value: '30' } });
    fireEvent.change(screen.getByLabelText('Motivo (queda en la auditoría)'), { target: { value: 'Acuerdo' } });
    fireEvent.click(screen.getByRole('button', { name: 'Resolver disputa' }));
    await waitFor(() => expect(api.resolver).toHaveBeenCalledWith('sesion-1', 30, 'Acuerdo'));
  });
});
