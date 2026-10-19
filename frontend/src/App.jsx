import { useEffect } from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import RutaProtegida from './components/RutaProtegida.jsx';
import { cargarSesion } from './controllers/acceso/acciones.js';
import Inicio from './views/acceso/Inicio.jsx';
import Ingreso from './views/acceso/Ingreso.jsx';
import NoAutorizado from './views/acceso/NoAutorizado.jsx';
import Registro from './views/acceso/Registro.jsx';

export default function App() {
  useEffect(() => { cargarSesion(); }, []);
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/ingreso" element={<Ingreso />} />
        <Route path="/registro" element={<Registro />} />
        <Route path="/no-autorizado" element={<NoAutorizado />} />
        <Route path="/" element={<RutaProtegida><Inicio /></RutaProtegida>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
