<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../lib/api.js';
  import { VIEWS } from '../lib/views.js';
  import { makeCapabilityTree } from '../lib/cytoscape.js';
  import { Button } from '$lib/components/ui/button';

  export let params = {};

  let container;
  let tooltipEl;
  let cy = null;
  let loading = true;
  let error = null;
  let empty = false;

  $: wsId = params.wsId;
  $: meta = VIEWS['capability-tree'] || { label: 'Capability Tree', desc: '' };

  onMount(async () => {
    loading = true;
    error = null;
    empty = false;
    try {
      const data = await api.get('/workspaces/' + wsId + '/views/capability-tree/tree');
      const nodes = data.nodes || [];
      if (nodes.length === 0) {
        empty = true;
        loading = false;
        return;
      }
      loading = false;
      await tick();
      cy = makeCapabilityTree(container, nodes);

      // Tooltip
      cy.on('mouseover', 'node', e => {
        const d = e.target.data();
        const apps = d.apps || [];
        const displayName = d.kind === 'app' ? d.label.split('\n')[0] : d.label;
        const typeColor = d.kind === 'capability' ? '#e0af68'
          : d.appSub === 'component' ? '#7aa2f7'
          : d.appSub === 'service'   ? '#4a9eff'
          : d.appSub === 'function'  ? '#5a80a8'
          : '#6b7280';
        tooltipEl.querySelector('.tt-name').textContent = displayName;
        const ttType = tooltipEl.querySelector('.tt-type');
        ttType.textContent = d.kind === 'app' ? (d.appType || 'Application') : 'Capability';
        ttType.style.color = typeColor;
        // Apps list
        let appsHtml = '';
        if (apps.length) {
          appsHtml = '<div class="tt-apps-label">Supported by</div>' +
            apps.map(a => `<div class="tt-app">· ${a.name}</div>`).join('');
        }
        const appsContainer = tooltipEl.querySelector('.tt-apps');
        if (appsContainer) appsContainer.innerHTML = appsHtml;
        tooltipEl.style.display = 'block';
      });
      cy.on('mousemove', e => {
        if (tooltipEl.style.display === 'none') return;
        tooltipEl.style.left = (e.originalEvent.clientX + 14) + 'px';
        tooltipEl.style.top  = (e.originalEvent.clientY - 10) + 'px';
      });
      cy.on('mouseout', 'node', () => { tooltipEl.style.display = 'none'; });

      cy.on('tap', 'node', e => {
        cy.elements().addClass('faded');
        e.target.removeClass('faded');
        e.target.neighborhood().removeClass('faded');
      });
      cy.on('tap', e => {
        if (e.target === cy) cy.elements().removeClass('faded');
      });
    } catch (e) {
      error = e.message;
      loading = false;
    }
  });

  onDestroy(() => {
    if (cy) cy.destroy();
  });

  function fit() { if (cy) cy.fit(undefined, 40); }
  function zoomIn() {
    if (cy) cy.zoom({ level: cy.zoom() * 1.2, renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 } });
  }
  function zoomOut() {
    if (cy) cy.zoom({ level: cy.zoom() / 1.2, renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 } });
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
      <div class="text-[40px] mb-3.5">◈</div>
      <p class="text-[14px] leading-relaxed">No Capability elements found.<br>Import or create elements with type <strong>Capability</strong>.</p>
    </div>
  {:else}
    <div class="cap-cy-wrap">
      <div id="cap-cy" bind:this={container}></div>
      <div class="cap-cy-controls">
        <Button variant="outline" size="icon" class="size-8" title="Fit" onclick={fit}>⊡</Button>
        <Button variant="outline" size="icon" class="size-8" title="Zoom in" onclick={zoomIn}>+</Button>
        <Button variant="outline" size="icon" class="size-8" title="Zoom out" onclick={zoomOut}>−</Button>
      </div>
      <div class="cap-cy-legend">
        <span><i style="background:#2a2010;border:2px solid #e0af68"></i> Capability</span>
        <span><i style="background:#0d1f2e;border:2px solid #7aa2f7"></i> Component</span>
        <span><i style="background:#0d1a28;border:2px solid #4a9eff;border-style:dashed"></i> Service</span>
        <span><i style="background:#0d1520;border:1px solid #3a5a80"></i> Function</span>
      </div>
    </div>
  {/if}
</div>

<div class="cap-tooltip" bind:this={tooltipEl}>
  <div class="tt-name"></div>
  <div class="tt-type"></div>
  <div class="tt-apps"></div>
</div>
