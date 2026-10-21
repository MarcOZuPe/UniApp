import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

import {
  obtenerSesion,
  declararAsistencia,
} from '../../controllers/sesion/apiSesion.js';

import { useSesion } from '../../controllers/acceso/useSesion.js';

function formatearFecha(valor) {
  return new Intl.DateTimeFormat('es-PE', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(valor));
}

export default function DetalleSesion() {
  const { id } = useParams();
  const { usuario } = useSesion();

  const [sesion, setSesion] = useState(null);
  const [cargando, setCargando] = useState(true);
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState('');
  const [mensaje, setMensaje] = useState('');

  useEffect(() => {
    let activo = true;

    async function cargar() {
      setCargando(true);
      setError('');
      setSesion(null);
      setMensaje('');

      try {
        const datos = await obtenerSesion(id);
        if (activo) {
          setSesion(datos);
        }
      } catch (err) {
        if (activo) {
          setError(err.message);
        }
      } finally {
        if (activo) {
          setCargando(false);
        }
      }
    }

    cargar();

    return () => {
      activo = false;
    };
  }, [id]);

  async function registrarAsistencia() {
    setEnviando(true);
    setError('');
    setMensaje('');

    try {
      await declararAsistencia(id);
      const actualizada = await obtenerSesion(id);

      setSesion(actualizada);
      setMensaje('Tu asistencia fue declarada correctamente.');
    } catch (err) {
      setError(err.message);
    } finally {
      setEnviando(false);
    }
  }

  if (cargando) {
    return (
      <main>
        <p role="status">Cargando sesión...</p>
      </main>
    );
  }

  if (!sesion) {
    return (
      <main>
        <h1>Sesión no disponible</h1>
        <p role="alert" className="error">
          {error || 'No se pudo recuperar la sesión.'}
        </p>
        <Link to="/">Volver al inicio</Link>
      </main>
    );
  }

  const asistencias = sesion.asistencias ?? [];

  const asistenciaPropia = asistencias.find(
    (asistencia) => asistencia.usuario_id === usuario?.id
  );

  const esParticipante =
    usuario?.id === sesion.tutor_id ||
    usuario?.id === sesion.escolar_id;

  return (
    <main>
      <Link to="/">Volver al inicio</Link>

      <h1>Detalle de sesión</h1>

      <section aria-labelledby="datos-sesion">
        <h2 id="datos-sesion">Información de la tutoría</h2>

        <p><strong>Sesión:</strong> {sesion.id}</p>
        <p><strong>Reserva:</strong> {sesion.reserva_id}</p>
        <p><strong>Inicio:</strong> {formatearFecha(sesion.inicio)}</p>
        <p><strong>Fin:</strong> {formatearFecha(sesion.fin)}</p>
        <p><strong>Duración:</strong> 60 minutos</p>
      </section>

      <section aria-labelledby="asistencias-sesion">
        <h2 id="asistencias-sesion">Asistencias</h2>

        <p>
          <strong>Tutor:</strong>{' '}
          {asistencias.some((a) => a.rol === 'tutor')
            ? 'Declarada'
            : 'Pendiente'}
        </p>

        <p>
          <strong>Escolar:</strong>{' '}
          {asistencias.some((a) => a.rol === 'escolar')
            ? 'Declarada'
            : 'Pendiente'}
        </p>

        {esParticipante && (
          <button
            type="button"
            onClick={registrarAsistencia}
            disabled={enviando || Boolean(asistenciaPropia)}
          >
            {enviando
              ? 'Registrando...'
              : asistenciaPropia
                ? 'Asistencia ya declarada'
                : 'Declarar mi asistencia'}
          </button>
        )}

        <p>
          <small>
            Declarar asistencia no acredita automáticamente
            60 minutos de participación verificada.
          </small>
        </p>
      </section>

      {mensaje && <p role="status">{mensaje}</p>}
      {error && <p role="alert" className="error">{error}</p>}
    </main>
  );
}
