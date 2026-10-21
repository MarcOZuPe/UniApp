export class ErrorApi extends Error {
  constructor(estado, cuerpo) {
    super(cuerpo?.mensaje || 'No se pudo completar la solicitud.');
    this.estado = estado;
    this.codigo = cuerpo?.codigo;
    this.campos = cuerpo?.campos || {};
  }
}

const CABECERAS_JSON = { 'Content-Type': 'application/json' };
const ERROR_DE_RED = { codigo: 'red', mensaje: 'No hay conexión con el servidor.' };


function construirPeticion(metodo, cuerpo, formulario = false) {
  return {
    method: metodo,
    credentials: 'same-origin',
    headers: cuerpo && !formulario ? CABECERAS_JSON : {},
    body: cuerpo ? (formulario ? cuerpo : JSON.stringify(cuerpo)) : undefined,
  };
}

async function enviar(implementacionFetch, url, opciones) {
  try {
    return await implementacionFetch(url, opciones);
  } catch {
    throw new ErrorApi(0, ERROR_DE_RED);
  }
}

async function interpretarCuerpo(respuesta) {
  const texto = await respuesta.text();
  if (!texto) return null;
  try {
    return JSON.parse(texto);
  } catch {
    return null;
  }
}

export function crearClienteHttp(implementacionFetch = (...argumentos) => fetch(...argumentos), base = '') {
  async function pedir(metodo, ruta, cuerpo, formulario = false) {
    const respuesta = await enviar(implementacionFetch, base + ruta, construirPeticion(metodo, cuerpo, formulario));
    const datos = await interpretarCuerpo(respuesta);
    if (!respuesta.ok) throw new ErrorApi(respuesta.status, datos);
    return datos;
  }
  return {
    get: (ruta) => pedir('GET', ruta),
    post: (ruta, cuerpo) => pedir('POST', ruta, cuerpo),
    postFormulario: (ruta, formulario) => pedir('POST', ruta, formulario, true),
  };
}
