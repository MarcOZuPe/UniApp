import { describe, expect, it } from 'vitest';

import {
  agruparPorTutor,
  contarIncidentes,
  contarPendientes,
} from './agrupacion.js';

describe('agrupación de incidentes por tutor', () => {
  it('ordena primero a los tutores con más incidentes pendientes', () => {
    const grupos = agruparPorTutor([
      { tutor_id: 't-1', tutor_nombre: 'Tutor Uno', estado: 'revisado' },
      { tutor_id: 't-1', tutor_nombre: 'Tutor Uno', estado: 'pendiente' },
      { tutor_id: 't-2', tutor_nombre: 'Tutor Dos', estado: 'pendiente' },
      { tutor_id: 't-2', tutor_nombre: 'Tutor Dos', estado: 'pendiente' },
    ]);

    expect(grupos.map((grupo) => grupo.tutorNombre)).toEqual(['Tutor Dos', 'Tutor Uno']);
    expect(grupos[0].pendientes).toBe(2);
    expect(grupos[1].pendientes).toBe(1);
  });

  it('usa el nombre del tutor cuando no hay identificador', () => {
    const grupos = agruparPorTutor([
      { tutor_nombre: 'Tutor Sin ID', estado: 'pendiente' },
      { tutor_nombre: 'Tutor Sin ID', estado: 'revisado' },
    ]);

    expect(grupos).toHaveLength(1);
    expect(grupos[0].clave).toBe('Tutor Sin ID');
    expect(grupos[0].incidentes).toHaveLength(2);
  });
});

describe('textos de conteo', () => {
  it('concuerda en singular y plural', () => {
    expect(contarIncidentes(1)).toBe('1 reporte');
    expect(contarIncidentes(2)).toBe('2 reportes');
    expect(contarPendientes(1)).toBe('1 pendiente');
    expect(contarPendientes(0)).toBe('0 pendientes');
  });
});
