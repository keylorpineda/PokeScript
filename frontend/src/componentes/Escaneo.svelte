<script>
  // Mientras compila: la Pokédex escanea el código y se encienden las fases.
  import { onMount } from 'svelte';

  const FASES = ['LEXER', 'PARSER', 'ANALIZADOR'];
  let encendidas = $state(0);

  onMount(() => {
    const id = setInterval(() => {
      if (encendidas < FASES.length) {
        encendidas += 1;
      }
    }, 380);
    return () => clearInterval(id);
  });
</script>

<div class="escaneo">
  <div class="rayo"></div>
  <div class="fases marco oscuro">
    <span class="lupa">ESCANEANDO</span>
    {#each FASES as f, i (f)}
      {#if i}<span class="flecha"></span>{/if}
      <span class="fase" class:on={i < encendidas}>{f}</span>
    {/each}
  </div>
</div>

<style>
  .escaneo {
    position: absolute;
    inset: 0;
    z-index: 5;
    pointer-events: none;
    background: repeating-linear-gradient(
      0deg,
      transparent 0 3px,
      color-mix(in srgb, var(--acento) 7%, transparent) 3px 4px
    );
    animation: entrar 0.2s steps(2) both;
  }
  @keyframes entrar {
    from {
      opacity: 0;
    }
  }
  .rayo {
    position: absolute;
    left: 0;
    right: 0;
    height: 6px;
    background: #7fffb0;
    box-shadow:
      0 0 18px 6px rgba(127, 255, 176, 0.55),
      0 -40px 40px rgba(127, 255, 176, 0.15);
    animation: bajar 0.9s steps(18) infinite;
  }
  @keyframes bajar {
    from {
      top: 0;
    }
    to {
      top: 100%;
    }
  }
  .fases {
    position: absolute;
    left: 50%;
    top: 16px;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 2px 8px;
    font-family: var(--titulo);
    font-size: 16px;
    letter-spacing: 1px;
    white-space: nowrap;
  }
  .lupa {
    color: #7fffb0;
    animation: parpadeo 0.6s steps(1) infinite;
  }
  .fase {
    color: #58587a;
  }
  .fase.on {
    color: #ffcb05;
    text-shadow: 0 0 8px rgba(255, 203, 5, 0.6);
  }
  .flecha {
    border-top: 5px solid transparent;
    border-bottom: 5px solid transparent;
    border-left: 8px solid #58587a;
  }
</style>
