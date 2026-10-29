import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { panelCoordinacion } from '../../../application/acceso/acciones.js';
import { useSesion } from '../../hooks/useSesion.js';

export default function Coordinacion() {
  const { usuario } = useSesion();
  const [panel, setPanel] = useState(null);
  const [error, setError] = useState('');

  useEffect(() => {
    let vigente = true;
    panelCoordinacion()
      .then((datos) => {
        if (vigente) setPanel(datos);
      })
      .catch((fallo) => {
        if (vigente) setError(fallo.message);
      });
    return () => {
      vigente = false;
    };
  }, []);

  return (
    <main>
      <h1>Zona de coordinación</h1>
      <p>Hola, {usuario?.nombre_uso}.</p>
      {error && <p className="error" role="alert">{error}</p>}
      {!panel && !error && <p role="status">Cargando panel…</p>}
      {panel && (
        <section>
          <h2>{panel.mensaje}</h2>
          <p>Permiso activo: <code>coordinacion:acceder</code></p>
          <p><Link to="/coordinacion/tutores">Revisar expedientes de tutores</Link></p>
          <p><Link to="/coordinacion/catalogo">Importar y aprobar el catálogo escolar</Link></p>
        </section>
      )}
      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}
