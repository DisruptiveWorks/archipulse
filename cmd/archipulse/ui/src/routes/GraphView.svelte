<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../lib/api.js';
  import { VIEWS } from '../lib/views.js';
  import { makeGraph } from '../lib/cytoscape.js';
  import { Button } from '$lib/components/ui/button';

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
  <div class="flex items-start justify-between mb-6 gap-4">
    <div>
      <h1 class="text-[18px] font-semibold">{meta.label}</h1>
      <div class="text-muted-foreground text-[13px] mt-0.5">{meta.desc}</div>
    </div>
    <div class="flex gap-2">
      <Button variant="outline" size="sm" onclick={fit}>⊡ Fit</Button>
      <Button variant="outline" size="sm" onclick={relayout}>↺ Re-layout</Button>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">Error: {error}</div>
  {:else if empty}
    <div class="text-center py-16 px-6 text-muted-foreground">
      <div class="text-[40px] mb-3.5">📭</div>
      <p class="text-[14px] leading-relaxed">No application elements — import a model first.</p>
    </div>
  {:else}
    <div class="cy-container" bind:this={container}></div>
    <div class="mt-2.5 text-muted-foreground text-[11px]">
      Scroll to zoom · Drag to pan · Click a node to highlight connections
    </div>
  {/if}
</div>
