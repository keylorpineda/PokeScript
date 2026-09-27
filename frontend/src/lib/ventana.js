// Pantalla completa: con Wails se usa la ventana nativa; en el navegador, la
// API de pantalla completa del documento.

export async function alternarPantallaCompleta() {
  const rt = window.runtime;
  if (rt?.WindowIsFullscreen) {
    if (await rt.WindowIsFullscreen()) rt.WindowUnfullscreen();
    else rt.WindowFullscreen();
    return;
  }
  try {
    if (document.fullscreenElement) await document.exitFullscreen();
    else await document.documentElement.requestFullscreen();
  } catch {
    // El navegador no lo permite: no pasa nada.
  }
}
