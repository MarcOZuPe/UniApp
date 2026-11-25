import path from 'node:path';

const raiz = '/src/frontend';

export default class ReporterResumen {
  onInit(ctx) {
    this.ctx = ctx;
    this.vistos = new Set();
    this.total = 0;
    this.fallos = 0;
  }

  onTaskUpdate(paquetes) {
    for (const paquete of paquetes) {
      const tarea = this.ctx.state.idMap.get(paquete[0]);
      if (!tarea || tarea.type !== 'suite' || !tarea.filepath) continue;
      if (!tarea.result?.state || tarea.result.state === 'run') continue;
      if (this.vistos.has(tarea.id)) continue;

      this.vistos.add(tarea.id);
      this.total += 1;

      const fallo = tarea.result.state === 'fail';
      if (fallo) this.fallos += 1;

      const archivo = `frontend/${path.relative(raiz, tarea.filepath)}`;
      console.log(`${fallo ? 'FAIL' : 'PASS'} ${archivo}`);
    }
  }

  onFinished() {
    console.log(`Frontend: ${this.fallos}/${this.total} archivos fallaron`);
    if (this.fallos > 0) process.exitCode = 1;
  }
}
