import { useEffect, useState } from 'react';

import {
  cambiarAutorizacion,
  historialAutorizaciones,
  listarVinculaciones,
} from '../../adapters/http/apoderadoApi.js';

// Estado y acciones del panel del apoderado. Mantener este hook fuera de la
// vista permite que la página se centre en presentar la información.
export function usePanelEscolares() {
  const [vinculaciones, setVinculaciones] = useState([]);
  const [escolarSeleccionadoID, setEscolarSeleccionadoID] = useState('');
  const [historialesPorEscolar, setHistorialesPorEscolar] = useState({});
  const [escolarCargandoHistorial, setEscolarCargandoHistorial] = useState('');
  const [mensajeEstado, setMensajeEstado] = useState('');
  const [mensajeEsError, setMensajeEsError] = useState(false);
  const [cargandoVinculaciones, setCargandoVinculaciones] = useState(true);
  const [escolarCambiandoAutorizacion, setEscolarCambiandoAutorizacion] = useState('');

  async function cargarHistorial(escolarID) {
    setEscolarCargandoHistorial(escolarID);
    try {
      const respuestaHistorial = await historialAutorizaciones(escolarID);
      setHistorialesPorEscolar((historialesActuales) => ({
        ...historialesActuales,
        [escolarID]: respuestaHistorial.autorizaciones || [],
      }));
    } catch {
      setHistorialesPorEscolar((historialesActuales) => ({
        ...historialesActuales,
        [escolarID]: null,
      }));
    } finally {
      setEscolarCargandoHistorial('');
    }
  }

  async function cargarVinculaciones() {
    try {
      const respuestaVinculaciones = await listarVinculaciones();
      const vinculacionesRecibidas = respuestaVinculaciones.vinculaciones || [];
      setVinculaciones(vinculacionesRecibidas);
      setEscolarSeleccionadoID((seleccionActual) =>
        vinculacionesRecibidas.some(
          (vinculacion) => vinculacion.escolar.id === seleccionActual
        )
          ? seleccionActual
          : ''
      );
    } catch (errorConsulta) {
      setMensajeEsError(true);
      setMensajeEstado(errorConsulta.message);
    } finally {
      setCargandoVinculaciones(false);
    }
  }

  useEffect(() => {
    cargarVinculaciones();
  }, []);

  function seleccionarEscolar(evento) {
    const escolarID = evento.target.value;
    setEscolarSeleccionadoID(escolarID);
    setMensajeEstado('');
    setMensajeEsError(false);

    const vinculacion = vinculaciones.find(
      (vinculacionCandidata) => vinculacionCandidata.escolar.id === escolarID
    );
    if (
      vinculacion?.estado === 'aceptada' &&
      historialesPorEscolar[escolarID] === undefined
    ) {
      cargarHistorial(escolarID);
    }
  }

  async function alternarAutorizacion(vinculacion) {
    const autorizacionVigente = vinculacion.autorizacion?.autorizado === true;
    setEscolarCambiandoAutorizacion(vinculacion.escolar.id);
    setMensajeEstado('');
    setMensajeEsError(false);
    try {
      const respuestaCambio = await cambiarAutorizacion(
        vinculacion.escolar.id,
        !autorizacionVigente
      );
      setMensajeEstado(respuestaCambio.mensaje);
      await cargarVinculaciones();
      await cargarHistorial(vinculacion.escolar.id);
    } catch (errorCambio) {
      setMensajeEsError(true);
      setMensajeEstado(errorCambio.message);
    } finally {
      setEscolarCambiandoAutorizacion('');
    }
  }

  return {
    vinculaciones,
    escolarSeleccionadoID,
    historialesPorEscolar,
    escolarCargandoHistorial,
    mensajeEstado,
    mensajeEsError,
    cargandoVinculaciones,
    escolarCambiandoAutorizacion,
    seleccionarEscolar,
    alternarAutorizacion,
  };
}
