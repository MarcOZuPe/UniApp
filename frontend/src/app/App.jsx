import { useEffect } from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';

import RutaProtegida from '../presentation/components/RutaProtegida.jsx';
import RutaProtegidaConRetorno from '../presentation/components/RutaProtegidaConRetorno.jsx';
import { cargarSesion } from '../application/acceso/acciones.js';

import Inicio from '../presentation/pages/acceso/Inicio.jsx';
import Ingreso from '../presentation/pages/acceso/Ingreso.jsx';
import NoAutorizado from '../presentation/pages/acceso/NoAutorizado.jsx';
import Recuperar from '../presentation/pages/acceso/Recuperar.jsx';
import Registro from '../presentation/pages/acceso/Registro.jsx';
import Restablecer from '../presentation/pages/acceso/Restablecer.jsx';
import Confirmar from '../presentation/pages/acceso/Confirmar.jsx';
import Coordinacion from '../presentation/pages/acceso/Coordinacion.jsx';

import Disponibilidad from '../presentation/pages/oferta/Disponibilidad.jsx';
import FormularioReserva from '../presentation/pages/reserva/FormularioReserva.jsx';
import Agenda from '../presentation/pages/reserva/Agenda.jsx';
import DetalleSesion from '../presentation/pages/sesion/DetalleSesion.jsx';
import Notificaciones from '../presentation/pages/sesion/Notificaciones.jsx';

import VincularEscolar from '../presentation/pages/apoderado/VincularEscolar.jsx';
import PanelEscolares from '../presentation/pages/apoderado/PanelEscolares.jsx';
import AceptarVinculacion from '../presentation/pages/apoderado/AceptarVinculacion.jsx';
import MiApoderado from '../presentation/pages/apoderado/MiApoderado.jsx';

import ExpedienteTutor from '../presentation/pages/tutor/ExpedienteTutor.jsx';
import RevisarTutores from '../presentation/pages/tutor/RevisarTutores.jsx';

import Catalogo from '../presentation/pages/catalogo/Catalogo.jsx';
import GestionarCatalogo from '../presentation/pages/catalogo/GestionarCatalogo.jsx';

export default function App() {
    useEffect(() => {
        cargarSesion();
    }, []);

    return (
        <BrowserRouter>
            <Routes>
                <Route path="/catalogo" element={<Catalogo />} />

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
                    path="/coordinacion/catalogo"
                    element={
                        <RutaProtegida roles={['coordinador']}>
                            <GestionarCatalogo />
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
                    path="/agenda"
                    element={
                        <RutaProtegida roles={['escolar', 'tutor']}>
                            <Agenda />
                        </RutaProtegida>
                    }
                />

                <Route
                    path="/notificaciones"
                    element={
                        <RutaProtegida roles={['escolar', 'tutor']}>
                            <Notificaciones />
                        </RutaProtegida>
                    }
                />

                <Route
                    path="/apoderado/vincular"
                    element={
                        <RutaProtegida roles={['apoderado']}>
                            <VincularEscolar />
                        </RutaProtegida>
                    }
                />

                <Route
                    path="/panel_escolares"
                    element={
                        <RutaProtegida roles={['apoderado']}>
                            <PanelEscolares />
                        </RutaProtegida>
                    }
                />

                <Route
                    path="/vinculacion/aceptar"
                    element={
                        <RutaProtegidaConRetorno roles={['escolar']}>
                            <AceptarVinculacion />
                        </RutaProtegidaConRetorno>
                    }
                />

                <Route
                    path="/mi-apoderado"
                    element={
                        <RutaProtegida roles={['escolar']}>
                            <MiApoderado />
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
