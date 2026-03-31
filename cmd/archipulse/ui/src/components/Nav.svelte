<script>
  import { push } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';

  export let wsId = null;
  export let wsName = null;
  export let viewLabel = null;

  function showCreateWs() {
    // Dispatch event to App.svelte to open modal
    window.dispatchEvent(new CustomEvent('archipulse:create-ws'));
  }
</script>

<nav>
  <a class="nav-logo" href="#/" use:push={'/'}>
    <svg width="22" height="22" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg">
      <polygon points="16,2 27,8 27,22 16,28 5,22 5,8" fill="#E85D3A"/>
      <polygon points="16,9 22,13 22,21 16,25 10,21 10,13" fill="none" stroke="white" stroke-width="0.8" stroke-linejoin="round" opacity="0.4"/>
      <polygon points="16,14 19,16 19,19 16,21 13,19 13,16" fill="white" opacity="0.15"/>
    </svg>
    <div class="nav-wordmark"><span class="w-archi">Archi</span><span class="w-pulse">Pulse</span></div>
  </a>
  <div class="nav-sep"></div>
  <div class="breadcrumb">
    {#if !wsId}
      <span style="color:var(--text-muted)">Workspaces</span>
    {:else if !viewLabel}
      <a href="#/">Workspaces</a>
      <span class="sep">/</span>
      <span class="current">{wsName || wsId}</span>
    {:else}
      <a href="#/">Workspaces</a>
      <span class="sep">/</span>
      <a href="#/ws/{wsId}">{wsName || wsId}</a>
      <span class="sep">/</span>
      <span class="current">{viewLabel}</span>
    {/if}
  </div>
  <div class="nav-spacer"></div>
  <Button size="sm" onclick={showCreateWs}>+ New workspace</Button>
</nav>
