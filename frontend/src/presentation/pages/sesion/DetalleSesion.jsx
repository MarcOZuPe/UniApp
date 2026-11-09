import { useEffect, useRef, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

import {
  obtenerSesion,
  declararAsistencia,
  reportarAusencia,
  registrarResultado,
} from '../../../adapters/http/sesionApi.js';

import { useSesion } from '../../hooks/useSesion.js';

import '../../styles/sesion.css';

function formatearFecha(valor) {
  return new Intl.DateTimeFormat('es-PE', {
    timeZone: 'America/Lima',
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(valor));
}

export default function DetalleSesion() {
  const { id } = useParams();
  const { usuario } = useSesion();

  const declarando = useRef(false);
  const [sesion, setSesion] = useState(null);
  const [cargando, setCargando] = useState(true);
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState('');
  const [mensaje, setMensaje] = useState('');
  const [resumen, setResumen] = useState('');
  const [guardandoResultado, setGuardandoResultado] = useState(false);

  useEffect(() => {
    let activo = true;

    async function cargar() {
      setCargando(true);
      setError('');
      setSesion(null);
      setResumen('');
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
    if (declarando.current || !sesion ||
      (usuario?.id !== sesion.tutor_id && usuario?.id !== sesion.escolar_id) ||
      sesion.asistencias?.some(a => a.usuario_id === usuario?.id) ||
      sesion.ausencias?.some(a => a.usuario_id === usuario?.id)) return;
    declarando.current = true;
    setEnviando(true);
    setError('');
    setMensaje('');

    try {
      const asistencia = await declararAsistencia(id);
      // La respuesta del POST confirma la persistencia incluso si falla la recarga.
      setSesion(anterior => ({ ...anterior, asistencias: [...(anterior.asistencias ?? []), asistencia] }));
      setMensaje('Tu asistencia fue declarada correctamente.');
      try {
        setSesion(await obtenerSesion(id));
      } catch {
        setError('Tu asistencia se guardó, pero no pudimos actualizar los demás estados. Vuelve a entrar para consultarlos.');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      declarando.current = false;
      setEnviando(false);
    }
  }

  async function registrarAusencia() {
    if (declarando.current || !sesion ||
      (usuario?.id !== sesion.tutor_id && usuario?.id !== sesion.escolar_id) ||
      sesion.asistencias?.some(a => a.usuario_id === usuario?.id) ||
      sesion.ausencias?.some(a => a.usuario_id === usuario?.id)) return;
    declarando.current = true;
    setEnviando(true);
    setError('');
    setMensaje('');
    try {
      const ausencia = await reportarAusencia(id);
      setSesion(anterior => ({ ...anterior, ausencias: [...(anterior.ausencias ?? []), ausencia] }));
      setMensaje('Tu ausencia fue reportada correctamente.');
      try {
        setSesion(await obtenerSesion(id));
      } catch {
        setError('Tu ausencia se guardó, pero no pudimos actualizar los demás estados. Vuelve a entrar para consultarlos.');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      declarando.current = false;
      setEnviando(false);
    }
  }

  async function guardarResultado(evento) {
    evento.preventDefault();
    if (declarando.current || !sesion || usuario?.id !== sesion.tutor_id || sesion.resultado) return;
    if (!resumen.trim() || Array.from(resumen.trim()).length > 2000) {
      setError('Escribe un resultado de entre 1 y 2.000 caracteres.');
      return;
    }
    declarando.current = true;
    setEnviando(true);
    setGuardandoResultado(true);
    setError('');
    setMensaje('');
    try {
      const resultado = await registrarResultado(id, resumen.trim());
      setSesion(anterior => ({ ...anterior, resultado }));
      setMensaje('El resultado de la tutoría fue registrado correctamente.');
      setResumen('');
    } catch (err) {
      setError(err.message);
    } finally {
      declarando.current = false;
      setEnviando(false);
      setGuardandoResultado(false);
    }
  }

  if (cargando) {
    return (
      <main className="pagina-sesion">
        <p role="status">Cargando sesión...</p>
      </main>
    );
  }

  if (!sesion) {
    return (
      <main className="pagina-sesion">
        <h1>Sesión no disponible</h1>
        <p role="alert" className="error">
          {error || 'No se pudo recuperar la sesión.'}
        </p>
        <Link to="/">Volver al inicio</Link>
      </main>
    );
  }

  const asistencias = sesion.asistencias ?? [];
  const ausencias = sesion.ausencias ?? [];
  const ausenciaPropia = ausencias.find(a => a.usuario_id === usuario?.id);
  const estadoParticipante = participanteID => asistencias.some(a => a.usuario_id === participanteID)
    ? 'Declarada' : ausencias.some(a => a.usuario_id === participanteID) ? 'Ausencia reportada' : 'Pendiente';

  const asistenciaPropia = asistencias.find(
    (asistencia) => asistencia.usuario_id === usuario?.id
  );

  const esParticipante =
    usuario?.id === sesion.tutor_id ||
    usuario?.id === sesion.escolar_id;

  return (
    <main className="pagina-sesion">
      <Link to="/">Volver al inicio</Link>

      <header className="sesion-encabezado">
        <p className="sesion-sobrelinea">FLUJO E · SESIÓN Y ASISTENCIA</p>
        <h1>Detalle de tutoría</h1>
        <p>Consulta tu sesión y declara tu propia asistencia o ausencia.</p>
      </header>
      {sesion.id === 'sesion-i05-demo' && (
        <p className="sesion-demo">Demostración I05 · Sesión sintética de prueba</p>
      )}
      <div className="sesion-paneles">

      <section className="sesion-tarjeta" aria-labelledby="datos-sesion">
        <h2 id="datos-sesion">Información de la tutoría</h2>

        <p><strong>Sesión:</strong> {sesion.id}</p>
        <p><strong>Reserva:</strong> {sesion.reserva_id}</p>
        <p className="sesion-zona">Horario de Lima (UTC−05:00)</p>
        <p><strong>Inicio:</strong> {formatearFecha(sesion.inicio)}</p>
        <p><strong>Fin:</strong> {formatearFecha(sesion.fin)}</p>
        <p><strong>Duración:</strong> 60 minutos</p>
      </section>

      <section className="sesion-tarjeta" aria-labelledby="asistencias-sesion">
        <h2 id="asistencias-sesion">Asistencias</h2>

        <p>
          <strong>Tutor:</strong>{' '}
          <span className={asistencias.some(a => a.usuario_id === sesion.tutor_id) ? 'sesion-estado declarada' : 'sesion-estado'}>{estadoParticipante(sesion.tutor_id)}</span>
        </p>

        <p>
          <strong>Escolar:</strong>{' '}
          <span className={asistencias.some(a => a.usuario_id === sesion.escolar_id) ? 'sesion-estado declarada' : 'sesion-estado'}>{estadoParticipante(sesion.escolar_id)}</span>
        </p>

        {esParticipante && (
          <button
            className="sesion-declarar"
            type="button"
            onClick={registrarAsistencia}
            disabled={enviando || Boolean(asistenciaPropia) || Boolean(ausenciaPropia)}
          >
            {enviando
              ? 'Registrando...'
              : asistenciaPropia
                ? 'Asistencia ya declarada'
                : 'Declarar mi asistencia'}
          </button>
        )}

        {esParticipante && (
          <button
            className="sesion-declarar"
            type="button"
            onClick={registrarAusencia}
            disabled={enviando || Boolean(asistenciaPropia) || Boolean(ausenciaPropia)}
          >
            {ausenciaPropia ? 'Ausencia ya reportada' : 'Reportar mi ausencia'}
          </button>
        )}
        <p><small>Solo puedes reportar tu propia ausencia. No puedes declarar asistencia y ausencia para la misma sesión.</small></p>

        <p>
          <small>
            Declarar asistencia no acredita automáticamente
            60 minutos de participación verificada.
          </small>
        </p>
      </section>

      </div>
      <section className="sesion-tarjeta sesion-resultado" aria-labelledby="resultado-sesion">
        <h2 id="resultado-sesion">Resultado de la tutoría</h2>
        {sesion.resultado ? (
          <>
            <p className="sesion-resumen">{sesion.resultado.resumen}</p>
            <p><small>Registrado por el tutor el {formatearFecha(sesion.resultado.registrado_en)}.</small></p>
          </>
        ) : usuario?.id === sesion.tutor_id ? (
          <form onSubmit={guardarResultado}>
            <label htmlFor="resumen-resultado">Resumen del resultado</label>
            <textarea
              id="resumen-resultado"
              rows={4}
              required
              value={resumen}
              onChange={evento => setResumen(evento.target.value)}
              disabled={enviando}
              aria-describedby="ayuda-resultado"
            />
            <p id="ayuda-resultado"><small>Describe lo ocurrido en la tutoría, hasta 2.000 caracteres. El resultado se registra una sola vez.</small></p>
            <button className="sesion-declarar" type="submit" disabled={enviando || !resumen.trim() || Array.from(resumen.trim()).length > 2000}>
              {guardandoResultado ? 'Guardando resultado...' : 'Registrar resultado'}
            </button>
          </form>
        ) : <p>El tutor aún no ha registrado el resultado.</p>}
        <p><small>El resultado no sustituye las declaraciones de asistencia ni acredita minutos verificados.</small></p>
      </section>
      {mensaje && <p className="sesion-exito" role="status">{mensaje}</p>}
      {error && <p role="alert" className="error">{error}</p>}
    </main>
  );
}
