// Impide instalar dependencias con npm o yarn: el proyecto usa solo pnpm.
// Se ejecuta en el script "preinstall" de package.json.
const agente = process.env.npm_config_user_agent ?? '';

if (!agente.startsWith('pnpm/')) {
  console.error('\n✖ Este proyecto usa pnpm. Instala las dependencias con:\n\n    pnpm install\n');
  process.exit(1);
}
