import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { obtenerNotificaciones } from '../../controllers/sesion/apiNotificaciones.js';

export default function Notificaciones() {
  const [datos, setDatos] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState('');
  useEffect(() => {
    let activo = true;
    obtenerNotificaciones().then(resultado => { if (activo) setDatos(resultado); })
      .catch(() => { if (activo) setError('No se pudieron cargar las notificaciones.'); })
      .finally(() => { if (activo) setCargando(false); });
    return () => { activo = false; };
  }, []);
  return (
    <main>
      <Link to="/">Volver al inicio</Link>
      <h1>Mis notificaciones</h1>
      <p>Confirmaciones y recordatorios de tus tutorías. Se muestran las últimas 100 notificaciones.</p>
      {cargando && <p role="status">Cargando notificaciones…</p>}
      {error && <p role="alert">{error}</p>}
      {!cargando && !error && (datos.length === 0 ? <p>Aún no tienes notificaciones.</p> : (
        <ul>{datos.map(n => (
          <li key={n.id}><article>
            <h2>{n.asunto}</h2><p>{n.cuerpo}</p>
            <time dateTime={n.fecha}>{new Intl.DateTimeFormat('es-PE', { timeZone: 'America/Lima', dateStyle: 'medium', timeStyle: 'short' }).format(new Date(n.fecha))} (Lima)</time>
          </article></li>
        ))}</ul>
      ))}
      <p><Link to="/agenda">Consultar mi agenda</Link></p>
    </main>
  );
}
