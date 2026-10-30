import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { obtenerPanelCoordinacion } from '../../../application/acceso/coordinacion.js';
import { useSesion } from '../../hooks/useSesion.js';

export default function Coordinacion() {
  const { usuario } = useSesion();
  const { panel, mensajeError } = usePanelCoordinacion();

  return (
    <main>
      <h1>Zona de coordinación</h1>
      <p>Hola, {usuario?.nombre_uso}.</p>
      {mensajeError && <p className="error" role="alert">{mensajeError}</p>}
      {!panel && !mensajeError && <p role="status">Cargando panel…</p>}
      {panel && <PanelCoordinacion panel={panel} />}
      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}

function usePanelCoordinacion() {
  const [panel, setPanel] = useState(null);
  const [mensajeError, setMensajeError] = useState('');

  useEffect(() => {
    let consultaVigente = true;

    obtenerPanelCoordinacion()
      .then((panelRecibido) => {
        if (consultaVigente) setPanel(panelRecibido);
      })
      .catch((errorConsulta) => {
        if (consultaVigente) setMensajeError(errorConsulta.message);
      });

    return () => {
      consultaVigente = false;
    };
  }, []);

  return { panel, mensajeError };
}

function PanelCoordinacion({ panel }) {
  return (
    <section>
      <h2>{panel.mensaje}</h2>
      <p>Permiso activo: <code>coordinacion:acceder</code></p>
      <p><Link to="/coordinacion/tutores">Revisar expedientes de tutores</Link></p>
      <p><Link to="/coordinacion/catalogo">Importar y aprobar el catálogo escolar</Link></p>
    </section>
  );
}
