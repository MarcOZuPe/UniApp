import { politicaDeRol } from './roles.js';

const MINIMO_BASICA = 10;
const MINIMO_ESTRICTA = 12;

export const crearPoliticaBasica = (minimo = MINIMO_BASICA) =>
  ({ clave = '', email = '' }) => {
    const observaciones = [];
    if ([...clave].length < minimo) observaciones.push(`debe tener al menos ${minimo} caracteres`);
    const nombreLocal = email.split('@')[0].toLowerCase();
    if (nombreLocal && clave.toLowerCase() === nombreLocal) observaciones.push('no puede ser igual a tu correo');
    return observaciones;
  };

export const crearPoliticaEstricta = (minimo = MINIMO_ESTRICTA) =>
  (datosClave = {}) => {
    const { clave = '' } = datosClave;
    const observaciones = crearPoliticaBasica(minimo)(datosClave);
    const combinaClases = /[A-Z]/.test(clave) &&
      /[a-z]/.test(clave) &&
      /\d/.test(clave) &&
      /[^A-Za-z0-9]/.test(clave);
    if (!combinaClases) observaciones.push('debe combinar mayúscula, minúscula, número y símbolo');
    return observaciones;
  };

const FABRICAS_POR_POLITICA = {
  basica: crearPoliticaBasica,
  estricta: crearPoliticaEstricta,
};

export const crearPolitica = (rol) =>
  (FABRICAS_POR_POLITICA[politicaDeRol(rol)] ?? crearPoliticaBasica)();
