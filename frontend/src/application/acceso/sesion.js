import { almacenSesion } from './almacenSesion.js';
import { api } from '../../adapters/http/apiAcceso.js';

export async function cargarSesion() {
  try {
    const usuario = await api.get('/api/auth/yo');
    guardarUsuarioEnSesion(usuario);
  } catch {
    cerrarSesionLocal();
  }
}

export async function ingresar({ email, clave }) {
  const usuario = await api.post('/api/auth/ingreso', { email, clave });
  guardarUsuarioEnSesion(usuario);
  return usuario;
}

export async function salir() {
  try {
    await api.post('/api/auth/salida');
  } finally {
    cerrarSesionLocal();
  }
}

function guardarUsuarioEnSesion(usuario) {
  almacenSesion.set({ usuario, cargando: false });
}

function cerrarSesionLocal() {
  almacenSesion.set({ usuario: null, cargando: false });
}
