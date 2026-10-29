import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import {
  importarCatalogo,
  obtenerCursosPendientes,
  obtenerSolicitudesColegio,
  revisarCurso,
} from '../../../adapters/http/catalogoApi.js';

const cursoVacio = () => ({ nombre: '', grado: '' });
const colegioVacio = () => ({ nombre: '', distrito: '' });

export default function GestionarCatalogo() {
  const [cursos, setCursos] = useState([]);
  const [solicitudes, setSolicitudes] = useState([]);
  const [fuente, setFuente] = useState('');
  const [cursosNuevos, setCursosNuevos] = useState([cursoVacio()]);
  const [colegiosNuevos, setColegiosNuevos] = useState([colegioVacio()]);
  const [motivos, setMotivos] = useState({});
  const [cargando, setCargando] = useState(true);
  const [procesando, setProcesando] = useState(false);
  const [error, setError] = useState('');
  const [confirmacion, setConfirmacion] = useState('');

  const cargar = useCallback(async () => {
    setError('');
    try {
      const [resultadoCursos, resultadoSolicitudes] = await Promise.all([
        obtenerCursosPendientes(),
        obtenerSolicitudesColegio(),
      ]);
      setCursos(resultadoCursos.cursos);
      setSolicitudes(resultadoSolicitudes.solicitudes);
    } catch (fallo) {
      setError(fallo.message);
    } finally {
      setCargando(false);
    }
  }, []);

  useEffect(() => {
    cargar();
  }, [cargar]);

  function actualizarFila(setFilas, filas, indice, campo, valor) {
    setFilas(filas.map((fila, posicion) => (
      posicion === indice ? { ...fila, [campo]: valor } : fila
    )));
  }

  async function importar(evento) {
    evento.preventDefault();
    setError('');
    setConfirmacion('');
    setProcesando(true);
    try {
      const resultado = await importarCatalogo({
        fuente: fuente.trim(),
        cursos: cursosNuevos.map(({ nombre, grado }) => ({
          nombre: nombre.trim(),
          grado: grado.trim(),
        })),
        colegios: colegiosNuevos.map(({ nombre, distrito }) => ({
          nombre: nombre.trim(),
          distrito: distrito.trim(),
        })),
      });
      const sintetica = resultado.fuente.toLocaleLowerCase().startsWith('sintetico:');
      setConfirmacion(sintetica
        ? 'Los datos de demostración se guardaron. Los cursos nuevos quedaron pendientes de revisión.'
        : 'El catálogo se actualizó. Los cursos nuevos quedaron pendientes de revisión.');
      setFuente('');
      setCursosNuevos([cursoVacio()]);
      setColegiosNuevos([colegioVacio()]);
      await cargar();
    } catch (fallo) {
      setError(fallo.message);
    } finally {
      setProcesando(false);
    }
  }

  async function decidir(curso, estado) {
    setError('');
    setConfirmacion('');
    const motivo = motivos[curso.id] || '';
    try {
      await revisarCurso(curso.id, {
        estado,
        ...(estado === 'rechazado' ? { motivo_rechazo: motivo } : {}),
      });
      await cargar();
    } catch (fallo) {
      setError(fallo.message);
    }
  }

  return (
    <main className="pagina-catalogo">
      <h1>Administración del catálogo escolar</h1>
      {error && <p className="error" role="alert">{error}</p>}
      {confirmacion && <p role="status">{confirmacion}</p>}
      <section>
        <h2>Agregar cursos y colegios</h2>
        <p>Indica el origen de los datos. Los cursos nuevos requieren aprobación antes de publicarse. Si son de prueba, identifícalos como datos de demostración.</p>
        <form onSubmit={importar}>
          <label className="campo">
            Origen de los datos
            <input
              value={fuente}
              onChange={(evento) => setFuente(evento.target.value)}
              maxLength={120}
              required
              placeholder="Por ejemplo, datos de demostración"
            />
          </label>

          <fieldset className="grupo-catalogo">
            <legend>Cursos y grados</legend>
            {cursosNuevos.map((curso, indice) => (
              <div className="fila-catalogo" key={indice}>
                <label className="campo">
                  Nombre del curso
                  <input
                    value={curso.nombre}
                    onChange={(evento) => actualizarFila(setCursosNuevos, cursosNuevos, indice, 'nombre', evento.target.value)}
                    maxLength={120}
                    required
                  />
                </label>
                <label className="campo">
                  Grado
                  <input
                    value={curso.grado}
                    onChange={(evento) => actualizarFila(setCursosNuevos, cursosNuevos, indice, 'grado', evento.target.value)}
                    maxLength={80}
                    required
                  />
                </label>
                {cursosNuevos.length > 1 && (
                  <button
                    className="boton-secundario"
                    type="button"
                    aria-label={`Quitar curso ${indice + 1}`}
                    onClick={() => setCursosNuevos(cursosNuevos.filter((_, posicion) => posicion !== indice))}
                  >
                    Quitar
                  </button>
                )}
              </div>
            ))}
            <button className="boton-secundario" type="button" onClick={() => setCursosNuevos([...cursosNuevos, cursoVacio()])}>
              Agregar otro curso
            </button>
          </fieldset>

          <fieldset className="grupo-catalogo">
            <legend>Colegios del padrón</legend>
            {colegiosNuevos.map((colegio, indice) => (
              <div className="fila-catalogo" key={indice}>
                <label className="campo">
                  Nombre del colegio
                  <input
                    value={colegio.nombre}
                    onChange={(evento) => actualizarFila(setColegiosNuevos, colegiosNuevos, indice, 'nombre', evento.target.value)}
                    maxLength={160}
                    required
                  />
                </label>
                <label className="campo">
                  Distrito
                  <input
                    value={colegio.distrito}
                    onChange={(evento) => actualizarFila(setColegiosNuevos, colegiosNuevos, indice, 'distrito', evento.target.value)}
                    maxLength={100}
                    required
                  />
                </label>
                {colegiosNuevos.length > 1 && (
                  <button
                    className="boton-secundario"
                    type="button"
                    aria-label={`Quitar colegio ${indice + 1}`}
                    onClick={() => setColegiosNuevos(colegiosNuevos.filter((_, posicion) => posicion !== indice))}
                  >
                    Quitar
                  </button>
                )}
              </div>
            ))}
            <button className="boton-secundario" type="button" onClick={() => setColegiosNuevos([...colegiosNuevos, colegioVacio()])}>
              Agregar otro colegio
            </button>
          </fieldset>

          <button type="submit" disabled={procesando}>
            {procesando ? 'Guardando…' : 'Guardar catálogo'}
          </button>
        </form>
      </section>

      <section>
        <h2>Cursos pendientes de aprobación</h2>
        {cargando && <p role="status">Cargando cursos…</p>}
        {!cargando && cursos.length === 0 && <p>No hay cursos pendientes de revisión.</p>}
        {cursos.map((curso) => (
          <article className="perfil-publico" key={curso.id}>
            <h3>{curso.grado} · {curso.nombre}</h3>
            <label className="campo">
              Motivo del rechazo
              <input
                value={motivos[curso.id] || ''}
                maxLength={500}
                onChange={(evento) => setMotivos({ ...motivos, [curso.id]: evento.target.value })}
              />
            </label>
            <button type="button" onClick={() => decidir(curso, 'aprobado')}>Aprobar</button>
            {' '}
            <button
              type="button"
              disabled={!motivos[curso.id]?.trim()}
              onClick={() => decidir(curso, 'rechazado')}
            >
              Rechazar
            </button>
          </article>
        ))}
      </section>

      <section>
        <h2>Solicitudes de colegios faltantes</h2>
        {solicitudes.length === 0
          ? <p>No hay solicitudes pendientes.</p>
          : (
            <ul>
              {solicitudes.map((solicitud) => (
                <li key={solicitud.id}>
                  {solicitud.nombre} · {solicitud.distrito} ·{' '}
                  {new Date(solicitud.solicitado_en).toLocaleString()}
                </li>
              ))}
            </ul>
          )}
      </section>
      <p><Link to="/coordinacion">Volver a coordinación</Link></p>
    </main>
  );
}
