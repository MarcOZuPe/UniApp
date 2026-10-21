import { Navigate } from 'react-router-dom';
import { useSesion } from '../controllers/acceso/useSesion.js';
import { puedeAcceder } from '../models/acceso/roles.js';

export default function RutaProtegida({ roles, children }) {
  const { usuario, cargando } = useSesion();
  if (cargando) return <p role="status">Cargando…</p>;
  if (!usuario) return <Navigate to="/ingreso" replace />;
  if (!puedeAcceder(usuario.rol, roles)) return <Navigate to="/no-autorizado" replace />;
  return children;
}
