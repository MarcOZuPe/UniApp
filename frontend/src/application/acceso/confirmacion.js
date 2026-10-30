import { api } from '../../adapters/http/apiAcceso.js';

export const confirmarCorreo = (token) =>
  api.post('/api/auth/confirmar', { token });

export const reenviarConfirmacion = (email) =>
  api.post('/api/auth/reenviar-confirmacion', { email });
