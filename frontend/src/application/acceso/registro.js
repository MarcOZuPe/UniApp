import { api } from '../../adapters/http/apiAcceso.js';

export const registrar = (formularioRegistro) =>
  api.post('/api/auth/registro', cuerpoDeRegistro(formularioRegistro));

function cuerpoDeRegistro({ email, nombre_uso, clave, rol }) {
  return { email, nombre_uso, clave, rol };
}
