import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { miApoderado } from '../../../adapters/http/apoderadoApi.js';

export default function MiApoderado() {
  const { vinculacion, cargandoVinculacion, mensajeError } = useMiApoderado();

  return (
    <main>
      <h1>Mi apoderado</h1>
      {cargandoVinculacion && <p role="status">Cargando…</p>}
      {mensajeError && <p className="error" role="alert">{mensajeError}</p>}
      {!cargandoVinculacion && !mensajeError && !vinculacion && (
        <p>Todavía no tienes un apoderado vinculado.</p>
      )}
      {vinculacion && <DetalleApoderado vinculacion={vinculacion} />}
      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}

function useMiApoderado() {
  const [vinculacion, setVinculacion] = useState(null);
  const [cargandoVinculacion, setCargandoVinculacion] = useState(true);
  const [mensajeError, setMensajeError] = useState('');

  useEffect(() => {
    miApoderado()
      .then((respuestaApoderado) => setVinculacion(respuestaApoderado.vinculacion))
      .catch((errorConsulta) => setMensajeError(errorConsulta.message))
      .finally(() => setCargandoVinculacion(false));
  }, []);

  return { vinculacion, cargandoVinculacion, mensajeError };
}

function DetalleApoderado({ vinculacion }) {
  return (
    <>
      <p>
        <strong>{vinculacion.apoderado.nombre_uso}</strong>
        <br />
        {vinculacion.apoderado.email}
      </p>
      <p>
        La vinculación fue aceptada y no se puede desvincular.
      </p>
      {vinculacion.autorizacion && (
        <EstadoAutorizacion autorizacion={vinculacion.autorizacion} />
      )}
    </>
  );
}

function EstadoAutorizacion({ autorizacion }) {
  return (
    <p role="status">
      {autorizacion.autorizado
        ? 'Tu apoderado autorizó la creación de nuevas tutorías.'
        : 'Tu apoderado revocó la autorización para crear nuevas tutorías.'}
      <br />
      <small>
        Estado registrado el {formatearFechaRegistro(autorizacion.creada_en)}.
        Las tutorías ya confirmadas no se modifican.
      </small>
    </p>
  );
}

function formatearFechaRegistro(valor) {
  return new Date(valor).toLocaleString('es-PE');
}
