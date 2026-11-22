import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom';
import { salir } from '../../application/acceso/sesion.js';
import { etiquetaRol } from '../../domain/acceso/roles.js';
import { useSesion } from '../hooks/useSesion.js';
import { enlacesDelMenu } from '../navegacion.js';
import Icono from './Icono.jsx';

const claseEnlace = ({ isActive }) =>
  'flex items-center gap-3 whitespace-nowrap rounded-lg px-3 py-2 text-[0.95rem] no-underline transition-colors ' +
  (isActive
    ? 'bg-guinda-700 font-semibold text-white shadow-sm dark:bg-guinda-300 dark:text-guinda-900'
    : 'text-stone-700 hover:bg-stone-100 hover:text-stone-950 dark:text-stone-300 dark:hover:bg-stone-800 dark:hover:text-white');

function Logo() {
  return (
    <Link to="/" className="flex items-center gap-2.5 no-underline">
      <span aria-hidden="true" className="grid size-9 place-items-center rounded-lg bg-guinda-700 text-lg font-bold text-white dark:bg-guinda-300 dark:text-guinda-900">U</span>
      <span className="leading-tight">
        <span className="block text-lg font-bold text-stone-900 dark:text-white">UniApp</span>
        <span className="block text-xs text-stone-500 dark:text-stone-400">Tutorías solidarias</span>
      </span>
    </Link>
  );
}

function iniciales(nombre = '') {
  return nombre.split(/\s+/).filter(Boolean).slice(0, 2).map(p => p[0]).join('').toUpperCase() || '?';
}

export default function DisenoApp() {
  const { usuario } = useSesion();
  const navegar = useNavigate();
  const enlaces = usuario ? enlacesDelMenu(usuario.rol) : [];

  async function cerrarSesion() {
    await salir();
    navegar('/ingreso', { replace: true });
  }

  return (
    <div className="min-h-screen md:grid md:grid-cols-[16.5rem_1fr]">
      <nav aria-label="Principal"
        className="sticky top-0 z-10 flex flex-col gap-4 border-b border-stone-200 bg-white/95 p-4 backdrop-blur md:h-screen md:border-b-0 md:border-r md:p-5 dark:border-stone-800 dark:bg-stone-900/95">
        <div className="flex items-center justify-between gap-3">
          <Logo />
          {usuario && (
            <button type="button" onClick={cerrarSesion} aria-label="Cerrar sesión"
              className="border-stone-300 bg-transparent p-2 text-stone-600 hover:bg-stone-100 hover:text-stone-900 md:hidden dark:border-stone-700 dark:text-stone-300 dark:hover:bg-stone-800">
              <Icono nombre="salir" />
            </button>
          )}
        </div>

        {usuario ? (
          <>
            <p className="m-0 hidden px-3 pt-4 text-xs font-semibold tracking-wider text-stone-500 uppercase md:block dark:text-stone-400">Menú</p>
            <ul className="-mx-1 m-0 flex list-none gap-1 overflow-x-auto p-0 px-1 md:mx-0 md:-mt-2 md:flex-col md:overflow-visible">
              {enlaces.map(({ ruta, texto, icono, fin }) => (
                <li key={ruta}>
                  <NavLink to={ruta} end={fin} className={claseEnlace}>
                    <Icono nombre={icono} />{texto}
                  </NavLink>
                </li>
              ))}
            </ul>

            <div className="mt-auto hidden rounded-xl border border-stone-200 bg-stone-50 p-3 md:block dark:border-stone-800 dark:bg-stone-950">
              <div className="flex items-center gap-3">
                <span aria-hidden="true" className="grid size-10 place-items-center rounded-full bg-guinda-100 font-bold text-guinda-700 dark:bg-guinda-900 dark:text-guinda-100">
                  {iniciales(usuario.nombre_uso)}
                </span>
                <div className="min-w-0">
                  <p className="m-0 truncate font-semibold">{usuario.nombre_uso}</p>
                  <p className="m-0 text-sm text-stone-600 dark:text-stone-400">{etiquetaRol(usuario.rol)}</p>
                </div>
              </div>
              <button type="button" onClick={cerrarSesion}
                className="mt-3 flex w-full items-center justify-center gap-2 border-stone-300 bg-white text-sm text-stone-700 hover:bg-stone-100 dark:border-stone-700 dark:bg-stone-900 dark:text-stone-200 dark:hover:bg-stone-800">
                <Icono nombre="salir" className="size-4" />Cerrar sesión
              </button>
            </div>
          </>
        ) : (
          <Link to="/ingreso" className="rounded-lg bg-guinda-700 px-4 py-2 text-center font-semibold text-white no-underline">Ingresar</Link>
        )}
      </nav>

      <div className="contenido-app min-w-0">
        <Outlet />
      </div>
    </div>
  );
}
