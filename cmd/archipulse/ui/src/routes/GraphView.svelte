<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../lib/api.js';
  import { VIEWS } from '../lib/views.js';
  import { makeGraph } from '../lib/cytoscape.js';

  export let params = {};

  let container;
  let cy = null;
  let loading = true;
  let error = null;
  let empty = false;

  $: wsId = params.wsId;
  $: viewName = params.viewName;
  $: meta = VIEWS[viewName] || { label: viewName, desc: '' };

  onMount(async () => {
    loading = true;
    error = null;
    empty = false;
    try {
      const data = await api.get('/workspaces/' + wsId + '/views/' + viewName + '/graph');
      if ((data.nodes || []).length === 0) {
        empty = true;
        loading = false;
        return;
      }
      loading = false;
      // Wait for container to be in DOM
      await tick();
      cy = makeGraph(container, data);
    } catch (e) {
      error = e.message;
      loading = false;
    }
  });

  onDestroy(() => {
    if (cy) cy.destroy();
  });

  function fit() {
    if (cy) cy.fit(undefined, 40);
  }

  function relayout() {
    if (cy) cy.layout({ name: 'cose', animate: true, padding: 40, nodeRepulsion: 8000 }).run();
  }

  async function tick() {
    return new Promise(resolve => requestAnimationFrame(resolve));
  }
</script>

<div class="content">
  <div class="page-header">
    <div>
      <h1>{meta.label}</h1>
      <div class="sub">{meta.desc}</div>
    </div>
    <div style="display:flex;gap:8px">
      <button class="btn btn-ghost btn-sm" on:click={fit}>⊡ Fit</button>
      <button class="btn btn-ghost btn-sm" on:click={relayout}>↺ Re-layout</button>
    </div>
  </div>

  {#if loading}
    <div class="loading"><div class="spinner"></div> Loading…</div>
  {:else if error}
    <div class="alert alert-error">Error: {error}</div>
  {:else if empty}
    <div class="empty-state">
      <div class="es-icon">📭</div>
      <p>No application elements — import a model first.</p>
    </div>
  {:else}
    <div class="cy-container" bind:this={container}></div>
    <div style="margin-top:10px;color:var(--muted);font-size:11px">
      Scroll to zoom · Drag to pan · Click a node to highlight connections
    </div>
  {/if}
</div>
