import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import { obtenerAgenda } from '../../../adapters/http/agendaApi.js';
import { useSesion } from '../../hooks/useSesion.js';
import '../../styles/i10.css';

function esProxima(reserva, ahora) {
  return new Date(reserva.fin).getTime() > ahora;
}

function ListaReservas({ reservas }) {
  if (reservas.length === 0) {
    return <p>No tienes tutorías en esta sección.</p>;
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
          </article>
        </li>
      ))}
    </ul>
  );
}

export default function Agenda() {
  const { usuario } = useSesion();

  const [reservas, setReservas] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState('');
  const [periodo, setPeriodo] = useState('proximas');

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
  }, []);

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

          <ListaReservas reservas={visibles} />
        </>
      )}

      {usuario?.rol === 'escolar' && (
        <p>
          <Link to="/reservar">Reservar otra tutoría</Link>
        </p>
      )}

      <p>
        <Link to="/">Volver al inicio</Link>
      </p>
    </main>
  );
}
