import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline';

const raiz = '/src/backend';

function recorrer(directorio, acumulado = []) {
  if (!fs.existsSync(directorio)) return acumulado;
  for (const entrada of fs.readdirSync(directorio, { withFileTypes: true })) {
    if (entrada.name === 'node_modules' || entrada.name === '.git') continue;
    const ruta = path.join(directorio, entrada.name);
    if (entrada.isDirectory()) {
      recorrer(ruta, acumulado);
    } else if (entrada.name.endsWith('_test.go')) {
      acumulado.push(ruta);
    }
  }
  return acumulado;
}

const mapa = new Map();
for (const archivo of recorrer(raiz)) {
  const contenido = fs.readFileSync(archivo, 'utf8');
  const relativa = path.relative(raiz, archivo).split(path.sep).join('/');
  const paquete = `uniapp/${path.posix.dirname(relativa)}`;
  for (const coincidencia of contenido.matchAll(/^func\s+(Test\w+)\s*\(/gm)) {
    mapa.set(`${paquete}::${coincidencia[1]}`, `backend/${relativa}`);
  }
}

const paquetes = new Map();
const paquetesImpresos = new Set();
let totalArchivos = 0;
let archivosFallidos = 0;

function datosPaquete(paquete) {
  if (!paquetes.has(paquete)) {
    paquetes.set(paquete, new Map());
  }
  return paquetes.get(paquete);
}

function imprimirPaquete(paquete, estadoPaquete) {
  if (paquetesImpresos.has(paquete)) return;
  paquetesImpresos.add(paquete);

  const archivos = datosPaquete(paquete);
  if (archivos.size === 0) {
    if (estadoPaquete === 'fail') {
      totalArchivos += 1;
      archivosFallidos += 1;
      console.log(`FAIL backend/${paquete}`);
    }
    return;
  }

  for (const [archivo, fallo] of [...archivos.entries()].sort(([a], [b]) => a.localeCompare(b))) {
    totalArchivos += 1;
    if (fallo) archivosFallidos += 1;
    console.log(`${fallo ? 'FAIL' : 'PASS'} ${archivo}`);
  }
}

const entrada = readline.createInterface({ input: process.stdin, crlfDelay: Infinity });

entrada.on('line', (linea) => {
  if (!linea.trim()) return;

  let evento;
  try {
    evento = JSON.parse(linea);
  } catch {
    return;
  }

  const accion = evento.Action;
  if (accion !== 'pass' && accion !== 'fail' && accion !== 'skip') return;

  if (!evento.Test) {
    if (evento.Package) imprimirPaquete(evento.Package, accion);
    return;
  }

  const test = evento.Test.split('/')[0];
  const clave = `${evento.Package}::${test}`;
  const archivo = mapa.get(clave) || `backend/${evento.Package || 'paquete'}`;
  const archivos = datosPaquete(evento.Package);
  const falloPrevio = archivos.get(archivo) || false;

  if (accion === 'fail') {
    archivos.set(archivo, true);
  } else if (!archivos.has(archivo)) {
    archivos.set(archivo, falloPrevio);
  }
});

entrada.on('close', () => {
  for (const [paquete, archivos] of paquetes) {
    if (!paquetesImpresos.has(paquete)) {
      imprimirPaquete(paquete, archivos.size > 0 ? 'pass' : 'fail');
    }
  }

  console.log(`Backend: ${archivosFallidos}/${totalArchivos} archivos fallaron`);
  process.exitCode = archivosFallidos > 0 ? 1 : 0;
});
