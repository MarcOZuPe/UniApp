import { Link } from 'react-router-dom';
import { etiquetaRol } from '../../../domain/acceso/roles.js';
import { useSesion } from '../../hooks/useSesion.js';
import { accesosDeInicio } from '../../navegacion.js';
import Icono from '../../components/Icono.jsx';

export default function Inicio() {
  const { usuario } = useSesion();
  const accesos = accesosDeInicio(usuario.rol);

  return (
    <main className="pagina-inicio">
      <section aria-labelledby="saludo"
        className="relative overflow-hidden rounded-2xl bg-guinda-700 p-6 text-white shadow-sm md:p-10">
        <div aria-hidden="true" className="absolute -top-16 -right-16 size-64 rounded-full bg-white/10" />
        <div aria-hidden="true" className="absolute -bottom-24 right-24 size-48 rounded-full bg-white/5" />
        <p className="relative m-0 inline-block rounded-full bg-white/15 px-3 py-1 text-sm font-semibold">
          {etiquetaRol(usuario.rol)}
        </p>
        <h1 id="saludo" className="relative mt-4 mb-2 text-3xl font-bold md:text-4xl">Hola, {usuario.nombre_uso}</h1>
        <p className="relative m-0 max-w-prose text-guinda-100">¿Qué quieres hacer hoy?</p>
      </section>

      <h2 className="mt-10 mb-4 text-lg">Accesos rápidos</h2>
      <ul className="m-0 grid list-none gap-4 p-0 sm:grid-cols-2 xl:grid-cols-3">
        {accesos.map(({ ruta, texto, textoInicio, descripcion, icono }, indice) => (
          <li key={ruta}>
            <Link to={ruta}
              aria-labelledby={`acceso-${indice}`}
              aria-describedby={descripcion ? `acceso-${indice}-desc` : undefined}
              className="group flex h-full items-start gap-4 rounded-xl border border-stone-200 bg-white p-5 text-stone-900 no-underline shadow-sm transition hover:-translate-y-0.5 hover:border-guinda-300 hover:shadow-md dark:border-stone-800 dark:bg-stone-900 dark:text-stone-100 dark:hover:border-guinda-600">
              <span className="grid size-11 shrink-0 place-items-center rounded-lg bg-guinda-50 text-guinda-700 dark:bg-guinda-900 dark:text-guinda-100">
                <Icono nombre={icono} className="size-6" />
              </span>
              <span className="min-w-0 flex-1">
                <span id={`acceso-${indice}`} className="block font-semibold">{textoInicio ?? texto}</span>
                {descripcion && <span id={`acceso-${indice}-desc`} className="mt-0.5 block text-sm text-stone-600 dark:text-stone-400">{descripcion}</span>}
              </span>
              <Icono nombre="flecha" className="mt-1 size-4 text-stone-400 transition group-hover:translate-x-0.5 group-hover:text-guinda-700 dark:group-hover:text-guinda-300" />
            </Link>
          </li>
        ))}
      </ul>
    </main>
  );
}
