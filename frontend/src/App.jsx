import { useEffect } from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';

import RutaProtegida from './components/RutaProtegida.jsx';
import { cargarSesion } from './controllers/acceso/acciones.js';

import Inicio from './views/acceso/Inicio.jsx';
import Ingreso from './views/acceso/Ingreso.jsx';
import NoAutorizado from './views/acceso/NoAutorizado.jsx';
import Recuperar from './views/acceso/Recuperar.jsx';
import Registro from './views/acceso/Registro.jsx';
import Restablecer from './views/acceso/Restablecer.jsx';
import Confirmar from './views/acceso/Confirmar.jsx';
import Coordinacion from './views/acceso/Coordinacion.jsx';

import FormularioReserva from './views/reserva/FormularioReserva.jsx';
import DetalleSesion from './views/sesion/DetalleSesion.jsx';

export default function App() {
    useEffect(() => {
        cargarSesion();
    }, []);

    return (
        <BrowserRouter>
            <Routes>
                <Route path="/ingreso" element={<Ingreso />} />
                <Route path="/registro" element={<Registro />} />
                <Route path="/recuperar" element={<Recuperar />} />
                <Route path="/restablecer" element={<Restablecer />} />
                <Route path="/confirmar" element={<Confirmar />} />
                <Route path="/no-autorizado" element={<NoAutorizado />} />

                <Route
                    path="/coordinacion"
                    element={
                        <RutaProtegida roles={['coordinador']}>
                            <Coordinacion />
                        </RutaProtegida>
                    }
                />

                <Route
                    path="/reservar"
                    element={
                        <RutaProtegida roles={['escolar']}>
                            <FormularioReserva />
                        </RutaProtegida>
                    }
                />

                <Route
                    path="/sesiones/:id"
                    element={
                        <RutaProtegida roles={['tutor', 'escolar']}>
                            <DetalleSesion />
                        </RutaProtegida>
                    }
                />

                <Route
                    path="/"
                    element={
                        <RutaProtegida>
                            <Inicio />
                        </RutaProtegida>
                    }
                />

                <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
        </BrowserRouter>
    );
}
