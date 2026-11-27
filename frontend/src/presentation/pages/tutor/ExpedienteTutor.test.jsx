// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import ExpedienteTutor from './ExpedienteTutor.jsx';
import { obtenerMiExpediente, presentarExpediente } from '../../../adapters/http/tutorApi.js';

vi.mock('../../../adapters/http/tutorApi.js', () => ({
  obtenerMiExpediente: vi.fn(),
  presentarExpediente: vi.fn(),
}));

beforeEach(() => {
  vi.resetAllMocks();
  obtenerMiExpediente.mockRejectedValue({ estado: 404 });
  presentarExpediente.mockReturnValue(new Promise(() => {}));
});

afterEach(cleanup);

describe('F-24: validación del expediente del tutor', () => {
  it('rechaza un archivo que no es PDF antes de enviarlo', async () => {
    render(
      <MemoryRouter>
        <ExpedienteTutor />
      </MemoryRouter>
    );

    await screen.findByRole('button', { name: 'Enviar a revisión' });
    fireEvent.change(screen.getByLabelText('Nombre de uso y breve presentación'), {
      target: { value: 'Presentación nueva' },
    });
    fireEvent.change(screen.getByLabelText('Cursos o temas que puedes enseñar'), {
      target: { value: 'Álgebra' },
    });
    fireEvent.change(screen.getByLabelText('Experiencia o formación relevante'), {
      target: { value: 'Experiencia docente' },
    });
    fireEvent.change(screen.getByLabelText(/Documentos de sustento/), {
      target: { files: [new File(['texto'], 'notas.txt', { type: 'text/plain' })] },
    });
    fireEvent.submit(document.querySelector('form'));

    await waitFor(() => expect(presentarExpediente).not.toHaveBeenCalled());
  });
});
