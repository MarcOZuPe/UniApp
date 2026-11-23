import { Outlet } from 'react-router-dom';
import Icono from './Icono.jsx';

const PUNTOS = [
  { icono: 'grupo', texto: 'Estudiantes UNI verificados acompañan a escolares.' },
  { icono: 'agenda', texto: 'Sesiones virtuales con horarios claros y recordatorios.' },
  { icono: 'verificado', texto: 'Cada minuto dictado se revisa antes de contar.' },
];

// Pantallas de acceso (ingreso, registro, recuperación…): panel institucional a la
// izquierda y el formulario de la página a la derecha. Cada página trae su propio <main>.
export default function DisenoAcceso() {
  return (
    <div className="grid min-h-screen md:grid-cols-[1.1fr_1fr]">
      <aside className="relative hidden flex-col justify-between overflow-hidden bg-guinda-700 p-12 text-white md:flex">
        <div aria-hidden="true" className="absolute -top-24 -right-24 size-96 rounded-full bg-white/10" />
        <div aria-hidden="true" className="absolute -bottom-32 -left-20 size-80 rounded-full bg-white/5" />

        <p className="relative m-0 flex items-center gap-2.5 text-xl font-bold">
          <span aria-hidden="true" className="grid size-9 place-items-center rounded-lg bg-white text-guinda-700">U</span>
          UniApp
        </p>

        <div className="relative">
          <p className="m-0 max-w-md text-4xl leading-tight font-bold">
            Tutorías solidarias entre estudiantes UNI y escolares.
          </p>
          <ul className="mt-8 list-none space-y-4 p-0">
            {PUNTOS.map(({ icono, texto }) => (
              <li key={icono} className="flex max-w-md items-center gap-3 text-guinda-50">
                <span className="grid size-9 shrink-0 place-items-center rounded-lg bg-white/15">
                  <Icono nombre={icono} />
                </span>
                {texto}
              </li>
            ))}
          </ul>
        </div>

        <p className="relative m-0 text-sm text-guinda-100">Piloto independiente con datos sintéticos · Proyecto académico</p>
      </aside>

      <div className="acceso-contenido flex items-center justify-center p-4">
        <div className="w-full">
          <p className="mb-0 flex items-center justify-center gap-2 text-xl font-bold text-guinda-700 md:hidden dark:text-guinda-300">
            <span aria-hidden="true" className="grid size-8 place-items-center rounded-lg bg-guinda-700 text-white dark:bg-guinda-300 dark:text-guinda-900">U</span>
            UniApp
          </p>
          <Outlet />
        </div>
      </div>
    </div>
  );
}
