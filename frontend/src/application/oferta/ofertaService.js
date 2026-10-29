import { crearClienteHttp, ErrorApi } from '../../infrastructure/http/fetchClient.js';

const DIA_MINUTOS = 24 * 60;
const DURACION_SLOT = 60;
const CUPO_MINIMO = 180;
const CUPO_MAXIMO = 360;

const cliente = crearClienteHttp();
const datosInicialesDemo = {
  cupo_semanal_minutos: CUPO_MINIMO,
  slots: [
    { id: 1, dia_semana: 1, inicio: '09:00', fin: '10:00', zona_horaria: 'America/Lima' },
    { id: 2, dia_semana: 3, inicio: '09:00', fin: '10:00', zona_horaria: 'America/Lima' },
    { id: 3, dia_semana: 5, inicio: '09:00', fin: '10:00', zona_horaria: 'America/Lima' },
  ],
};
let disponibilidadDemo = structuredClone(datosInicialesDemo);
let siguienteIdDemo = 4;

function minutos(hora) {
  if (hora === '24:00') return DIA_MINUTOS;
  const coincidencia = /^([01]\d|2[0-3]):([0-5]\d)$/.exec(hora);
  if (!coincidencia) return NaN;
  return Number(coincidencia[1]) * 60 + Number(coincidencia[2]);
}

function errorValidacion(mensaje, estado = 422, codigo = 'validacion') {
  return new ErrorApi(estado, { codigo, mensaje });
}

function validarSlot(slot) {
  const inicio = minutos(slot.inicio);
  const fin = minutos(slot.fin);
  if (!Number.isInteger(slot.dia_semana) || slot.dia_semana < 0 || slot.dia_semana > 6 ||
      !Number.isFinite(inicio) || !Number.isFinite(fin) || fin - inicio !== DURACION_SLOT) {
    throw errorValidacion('Cada slot debe durar 60 minutos y usar un día y una hora válidos.');
  }
  return { ...slot, inicioMinutos: inicio, finMinutos: fin };
}

function comprobarCupo(slots, cupo) {
  if (slots.length * DURACION_SLOT > cupo) {
    throw errorValidacion('La disponibilidad excede el cupo semanal.', 409, 'cupo_excedido');
  }
}

function comprobarSolapamientos(slots) {
  for (let i = 0; i < slots.length; i += 1) {
    for (let j = i + 1; j < slots.length; j += 1) {
      const a = validarSlot(slots[i]);
      const b = validarSlot(slots[j]);
      if (a.dia_semana === b.dia_semana &&
          a.inicioMinutos < b.finMinutos && b.inicioMinutos < a.finMinutos) {
        throw errorValidacion('Los slots del mismo día no pueden solaparse.', 409, 'solapamiento');
      }
    }
  }
}

function simular(tarea) {
  return Promise.resolve().then(tarea);
}

const servicioDemo = {
  async obtenerDisponibilidad() {
    return structuredClone(disponibilidadDemo);
  },
  agregarSlots(entrada) {
    return simular(() => {
      const propuestos = entrada.slots.map(validarSlot);
      const existentes = disponibilidadDemo.slots.map(validarSlot);
      comprobarSolapamientos([...existentes, ...propuestos]);
      comprobarCupo([...existentes, ...propuestos], disponibilidadDemo.cupo_semanal_minutos);
      const creados = propuestos.map(({ inicioMinutos, finMinutos, ...slot }) => ({
        ...slot,
        id: siguienteIdDemo++,
        zona_horaria: 'America/Lima',
      }));
      disponibilidadDemo = { ...disponibilidadDemo, slots: [...disponibilidadDemo.slots, ...creados] };
      return {
        cupo_semanal_minutos: disponibilidadDemo.cupo_semanal_minutos,
        slots: creados,
      };
    });
  },
  eliminarSlot(id) {
    return simular(() => {
      const slots = disponibilidadDemo.slots.filter((slot) => slot.id !== Number(id));
      if (slots.length === disponibilidadDemo.slots.length) {
        throw errorValidacion('No se encontró el slot solicitado.', 404, 'slot_no_encontrado');
      }
      disponibilidadDemo = { ...disponibilidadDemo, slots };
      return null;
    });
  },
  cambiarCupo(minutosNuevos) {
    return simular(() => {
      if (!Number.isInteger(minutosNuevos) ||
          minutosNuevos < CUPO_MINIMO || minutosNuevos > CUPO_MAXIMO ||
          minutosNuevos % DURACION_SLOT !== 0) {
        throw errorValidacion('El cupo debe estar entre 180 y 360 minutos, en múltiplos de 60.');
      }
      if (minutosNuevos < disponibilidadDemo.cupo_semanal_minutos) {
        throw errorValidacion(
          'El backend todavía no permite verificar compromisos activos para reducir el cupo.',
          503,
          'reduccion_no_verificable',
        );
      }
      comprobarCupo(disponibilidadDemo.slots, minutosNuevos);
      disponibilidadDemo = { ...disponibilidadDemo, cupo_semanal_minutos: minutosNuevos };
      return { cupo_semanal_minutos: minutosNuevos };
    });
  },
};

export const modoDemoOferta = import.meta.env.VITE_OFERTA_MOCK === 'true';

export const ofertaService = modoDemoOferta
  ? servicioDemo
  : {
      obtenerDisponibilidad: () => cliente.get('/api/oferta/slots'),
      agregarSlots: (entrada) => cliente.post('/api/oferta/slots', entrada),
      eliminarSlot: (id) => cliente.delete(`/api/oferta/slots/${encodeURIComponent(id)}`),
      cambiarCupo: (minutosNuevos) => cliente.put('/api/oferta/cupo', { minutos: minutosNuevos }),
    };

export const diasSemana = [
  { valor: 1, nombre: 'Lunes' },
  { valor: 2, nombre: 'Martes' },
  { valor: 3, nombre: 'Miércoles' },
  { valor: 4, nombre: 'Jueves' },
  { valor: 5, nombre: 'Viernes' },
  { valor: 6, nombre: 'Sábado' },
  { valor: 0, nombre: 'Domingo' },
];

export function validarNuevoSlot(nuevo, slots, cupoSemanal) {
  let slot;
  try {
    slot = validarSlot(nuevo);
  } catch (error) {
    return error.message;
  }
  if (slots.some((existente) => {
    const actual = validarSlot(existente);
    return actual.dia_semana === slot.dia_semana &&
      slot.inicioMinutos < actual.finMinutos &&
      actual.inicioMinutos < slot.finMinutos;
  })) {
    return 'Ya existe un horario que se solapa ese día.';
  }
  if ((slots.length + 1) * DURACION_SLOT > cupoSemanal) {
    return 'Este horario excedería tu cupo semanal.';
  }
  return '';
}
