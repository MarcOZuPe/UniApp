import { requiereCorreoInstitucional } from './roles.js';
import { crearPolitica, crearPoliticaBasica } from './politicas.js';

export { crearPolitica, crearPoliticaBasica, crearPoliticaEstricta } from './politicas.js';

const CORREO_VALIDO = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const DOMINIOS_INSTITUCIONALES = 'uni.edu.pe,uni.pe';

export function validarRegistro(formularioRegistro, dominiosTutor = DOMINIOS_INSTITUCIONALES) {
  const erroresPorCampo = {};
  const email = normalizarEmail(formularioRegistro.email);
  const errorCorreo = validarCorreo(email);

  if (errorCorreo) {
    erroresPorCampo.email = errorCorreo;
  } else if (requiereCorreoInstitucional(formularioRegistro.rol) &&
    !esCorreoInstitucional(email, dominiosTutor)) {
    erroresPorCampo.email = `Los tutores usan su correo institucional (${correosInstitucionales(dominiosTutor)})`;
  }

  validarNombreUso(erroresPorCampo, formularioRegistro.nombre_uso);
  validarClave(erroresPorCampo, {
    clave: formularioRegistro.clave || '',
    email,
    rol: formularioRegistro.rol,
  });
  validarConfirmacionClave(erroresPorCampo, formularioRegistro);
  return erroresPorCampo;
}

export function validarIngreso(formularioIngreso) {
  const erroresPorCampo = {};
  registrarErrorCorreo(erroresPorCampo, normalizarEmail(formularioIngreso.email));
  if (!formularioIngreso.clave) {
    erroresPorCampo.clave = 'Ingresa tu contraseña';
  }
  return erroresPorCampo;
}

export function validarRecuperar(formularioRecuperacion) {
  const erroresPorCampo = {};
  registrarErrorCorreo(erroresPorCampo, normalizarEmail(formularioRecuperacion.email));
  return erroresPorCampo;
}

export function validarRestablecer(formularioRestablecimiento) {
  const erroresPorCampo = {};
  validarClaveBasica(erroresPorCampo, formularioRestablecimiento.clave || '');
  validarConfirmacionClave(erroresPorCampo, formularioRestablecimiento);
  return erroresPorCampo;
}

export function dominiosDe(dominios) {
  return String(dominios ?? '')
    .split(',')
    .map((dominio) => dominio.trim().toLowerCase())
    .filter(Boolean);
}

export function correosInstitucionales(dominios = DOMINIOS_INSTITUCIONALES) {
  const correos = dominiosDe(dominios).map((dominio) => '@' + dominio);
  return correos.length ? correos.join(' o ') : 'correo institucional';
}

function normalizarEmail(email) {
  return String(email ?? '').trim().toLowerCase();
}

function registrarErrorCorreo(erroresPorCampo, email) {
  const errorCorreo = validarCorreo(email);
  if (errorCorreo) {
    erroresPorCampo.email = errorCorreo;
  }
}

function validarNombreUso(erroresPorCampo, nombreUso) {
  const longitudNombre = [...String(nombreUso ?? '').trim()].length;
  if (longitudNombre < 2 || longitudNombre > 60) {
    erroresPorCampo.nombre_uso = 'Entre 2 y 60 caracteres';
  }
}

function validarClave(erroresPorCampo, datosClave) {
  const observacionesClave = crearPolitica(datosClave.rol)(datosClave);
  if (observacionesClave.length) {
    erroresPorCampo.clave = 'La contraseña ' + observacionesClave.join('; ');
  }
}

function validarClaveBasica(erroresPorCampo, clave) {
  const observacionesClave = crearPoliticaBasica()({ clave });
  if (observacionesClave.length) {
    erroresPorCampo.clave = 'La contraseña ' + observacionesClave.join('; ');
  }
}

function validarConfirmacionClave(erroresPorCampo, datosConClave) {
  if (datosConClave.clave !== datosConClave.confirmar) {
    erroresPorCampo.confirmar = 'Las contraseñas no coinciden';
  }
}

function esCorreoInstitucional(email, dominios) {
  return dominiosDe(dominios).some((dominio) => email.endsWith('@' + dominio));
}

function validarCorreo(email) {
  return CORREO_VALIDO.test(email) ? '' : 'Correo no válido';
}
