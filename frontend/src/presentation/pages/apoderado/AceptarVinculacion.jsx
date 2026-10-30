import { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
  aceptarVinculacion as aceptarVinculacionApi,
  consultarVinculacion,
} from '../../../adapters/http/apoderadoApi.js';

export default function AceptarVinculacion() {
  const [parametros] = useSearchParams();
  const tokenVinculacion = parametros.get('token') || '';
  const {
    vinculacion,
    aceptada,
    mensajeError,
    cargandoVinculacion,
    aceptandoVinculacion,
    aceptar,
  } = useVinculacionAceptacion(tokenVinculacion);

  if (cargandoVinculacion) {
    return <main><p role="status">Validando enlace…</p></main>;
  }

  return (
    <main>
      <h1>Aceptar vinculación</h1>
      {mensajeError && <p className="error" role="alert">{mensajeError}</p>}
      {vinculacion && (
        <>
          <p>
            <strong>{vinculacion.apoderado.nombre_uso}</strong> ({vinculacion.apoderado.email})
            quiere vincularse como tu apoderado.
          </p>
          {aceptada ? (
            <p role="status">
              Vinculación aceptada. Tu apoderado ya está registrado y la relación
              no se puede deshacer.
            </p>
          ) : (
            <button onClick={aceptar} disabled={aceptandoVinculacion}>
              {aceptandoVinculacion ? 'Aceptando…' : 'Aceptar vinculación'}
            </button>
          )}
          <p><Link to="/mi-apoderado">Ver mi apoderado</Link></p>
        </>
      )}
      {!vinculacion && !mensajeError && <p>No se encontró una solicitud pendiente.</p>}
      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}

function useVinculacionAceptacion(tokenVinculacion) {
  const [vinculacion, setVinculacion] = useState(null);
  const [aceptada, setAceptada] = useState(false);
  const [mensajeError, setMensajeError] = useState('');
  const [cargandoVinculacion, setCargandoVinculacion] = useState(true);
  const [aceptandoVinculacion, setAceptandoVinculacion] = useState(false);

  useEffect(() => {
    if (!tokenVinculacion) {
      setMensajeError('El enlace no contiene un token de vinculación.');
      setCargandoVinculacion(false);
      return;
    }

    consultarVinculacion(tokenVinculacion)
      .then((respuestaConsulta) => setVinculacion(respuestaConsulta.vinculacion))
      .catch((errorConsulta) => setMensajeError(errorConsulta.message))
      .finally(() => setCargandoVinculacion(false));
  }, [tokenVinculacion]);

  async function aceptar() {
    setMensajeError('');
    setAceptandoVinculacion(true);
    try {
      const respuestaAceptacion = await aceptarVinculacionApi(tokenVinculacion);
      setVinculacion(respuestaAceptacion.vinculacion);
      setAceptada(true);
    } catch (errorAceptacion) {
      setMensajeError(errorAceptacion.message);
    } finally {
      setAceptandoVinculacion(false);
    }
  }

  return {
    vinculacion,
    aceptada,
    mensajeError,
    cargandoVinculacion,
    aceptandoVinculacion,
    aceptar,
  };
}
