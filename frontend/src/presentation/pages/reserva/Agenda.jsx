import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { cancelarReserva } from '../../../adapters/http/reservaApi.js';
import { obtenerAgenda } from '../../../adapters/http/agendaApi.js';
import { obtenerSesionPorReserva } from '../../../adapters/http/sesionApi.js';
import { useSesion } from '../../hooks/useSesion.js';
import '../../styles/i10.css';

function esProxima(reserva, ahora) {
  return new Date(reserva.fin).getTime() > ahora;
}

function ListaReservas({ reservas, rol, onCambio }) {
  const navigate = useNavigate();
  const [abriendoReserva, setAbriendoReserva] = useState(null);
  const [cancelando, setCancelando] = useState(null);
  const [errorSesion, setErrorSesion] = useState('');

  async function cancelar(reserva) {
    if (!window.confirm('¿Cancelar esta tutoría? El horario volverá a estar disponible.')) {
      return;
    }
    const motivo = window.prompt('Motivo de la cancelación (opcional):') ?? '';
    setCancelando(reserva.id);
    setErrorSesion('');
    try {
      await cancelarReserva(reserva.id, motivo);
      onCambio();
    } catch (err) {
      setErrorSesion(err.message || 'No se pudo cancelar la reserva.');
    } finally {
      setCancelando(null);
    }
  }

  async function abrirSesion(reservaID) {
    setAbriendoReserva(reservaID);
    setErrorSesion('');

    try {
      const sesion = await obtenerSesionPorReserva(reservaID);
      navigate(`/sesiones/${encodeURIComponent(sesion.id)}`);
    } catch (err) {
      setErrorSesion(err.message || 'No se pudo abrir la sesión.');
    } finally {
      setAbriendoReserva(null);
    }
  }

  if (reservas.length === 0) {
    return <p className="i10-vacio">No tienes tutorías en esta sección.</p>;
  }

  return (
    <ul>
      {reservas.map((reserva) => (
        <li key={reserva.id}>
          <article>
            <h3>{reserva.curso}</h3>

            <p>
              <strong>Inicio:</strong> {reserva.inicio_lima} (hora de Lima)
            </p>

            <p>
              <strong>Fin:</strong> {reserva.fin_lima} (hora de Lima)
            </p>

            <p>
              <strong>Estado:</strong> {reserva.estado}
            </p>

            <p>
              <strong>Tema:</strong> {reserva.tema}
            </p>

            <p>
              <strong>Tutor:</strong> {reserva.tutor_nombre}
            </p>

            <p>
              <small>Reserva: {reserva.id}</small>
            </p>

            <button
              type="button"
              onClick={() => abrirSesion(reserva.id)}
              disabled={abriendoReserva !== null || cancelando !== null}
            >
              {abriendoReserva === reserva.id
                ? 'Abriendo sesión...'
                : 'Ver detalle de sesión'}
            </button>

            {rol === 'escolar' && reserva.estado === 'confirmada' && (
              <button
                type="button"
                onClick={() => cancelar(reserva)}
                disabled={cancelando !== null}
              >
                {cancelando === reserva.id ? 'Cancelando…' : 'Cancelar reserva'}
              </button>
            )}
          </article>
        </li>
      ))}
      {errorSesion && (
        <li>
          <p role="alert" className="error">{errorSesion}</p>
        </li>
      )}
    </ul>
  );
}

export default function Agenda() {
  const { usuario } = useSesion();

  const [reservas, setReservas] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState('');
  const [periodo, setPeriodo] = useState('proximas');
  const [version, setVersion] = useState(0);

  useEffect(() => {
    let activo = true;

    async function cargar() {
      try {
        const resultado = await obtenerAgenda();

        if (activo) {
          setReservas(resultado);
        }
      } catch (err) {
        if (activo) {
          setError(err.message || 'No se pudo cargar la agenda.');
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
  }, [version]);

  const ahora = Date.now();

  const proximas = reservas.filter((reserva) =>
    esProxima(reserva, ahora)
  );

  const pasadas = reservas.filter((reserva) =>
    !esProxima(reserva, ahora)
  );

  const visibles = periodo === 'proximas' ? proximas : pasadas;

  return (
    <main className="i10-agenda">
      <h1>Mi agenda</h1>

      <p>
        {usuario?.rol === 'tutor'
          ? 'Consulta las tutorías que tienes asignadas.'
          : 'Consulta tus tutorías reservadas.'}
      </p>

      <p>Todos los horarios se muestran en hora de Lima.</p>

      {cargando && <p role="status">Cargando agenda…</p>}

      {error && (
        <p role="alert" className="error">
          {error}
        </p>
      )}

      {!cargando && !error && (
        <>
          <nav aria-label="Período de la agenda">
            <button
              type="button"
              aria-pressed={periodo === 'proximas'}
              onClick={() => setPeriodo('proximas')}
            >
              Próximas ({proximas.length})
            </button>

            <button
              type="button"
              aria-pressed={periodo === 'pasadas'}
              onClick={() => setPeriodo('pasadas')}
            >
              Pasadas ({pasadas.length})
            </button>
          </nav>

          <h2>
            {periodo === 'proximas'
              ? 'Próximas tutorías'
              : 'Tutorías anteriores'}
          </h2>

          <ListaReservas
            reservas={visibles}
            rol={usuario?.rol}
            onCambio={() => setVersion((valor) => valor + 1)}
          />
        </>
      )}

      {usuario?.rol === 'escolar' && (
        <p>
          <Link to="/reservar" className="i10-principal">Reservar otra tutoría</Link>
        </p>
      )}
    </main>
  );
}
