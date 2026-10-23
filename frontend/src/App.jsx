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

import Disponibilidad from './views/oferta/Disponibilidad.jsx';
import FormularioReserva from './views/reserva/FormularioReserva.jsx';
import Agenda from './views/reserva/Agenda.jsx';
import DetalleSesion from './views/sesion/DetalleSesion.jsx';
import Notificaciones from './views/sesion/Notificaciones.jsx';
import ExpedienteTutor from './views/tutor/ExpedienteTutor.jsx';
import RevisarTutores from './views/tutor/RevisarTutores.jsx';

export default function App() {
    useEffect(() => {
        cargarSesion();
    }, []);

    return (
        <BrowserRouter>
            <Routes>
                <Route path="/notificaciones" element={<RutaProtegida roles={['tutor', 'escolar']}><Notificaciones /></RutaProtegida>} />
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
                    path="/coordinacion/tutores"
                    element={
                        <RutaProtegida roles={['coordinador']}>
                            <RevisarTutores />
                        </RutaProtegida>
                    }
                />

                <Route
                    path="/expediente"
                    element={
                        <RutaProtegida roles={['tutor']}>
                            <ExpedienteTutor />
                        </RutaProtegida>
                    }
                />

                <Route
                    path="/tutor/disponibilidad"
                    element={
                        <RutaProtegida roles={['tutor']}>
                            <Disponibilidad />
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

                <Route
                    path="/agenda"
                    element={
                        <RutaProtegida roles={['escolar', 'tutor']}>
                            <Agenda />
                        </RutaProtegida>
                    }
                 />

                <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
        </BrowserRouter>
    );
}
