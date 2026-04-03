<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../lib/api.js';
  import { VIEWS } from '../lib/views.js';
  import { makeGraph, makeDependencyGraph } from '../lib/cytoscape.js';
  import { Button } from '$lib/components/ui/button';

  export let params = {};

  let container;
  let cy = null;
  let loading = true;
  let error = null;
  let empty = false;

  $: wsId     = params.wsId;
  $: viewName = params.viewName;
  $: meta     = VIEWS[viewName] || { label: viewName, desc: '' };
  $: isDependency = viewName === 'application-dependency';

  // ── Dependency-specific state ────────────────────────────────────────────
  let allNodes   = [];   // { id, name, type, lifecycle_status }
  let searchQ    = '';
  let selectedId = null; // currently focused node id

  // Relationship type filters
  const REL_TYPES = [
    { key: 'serving',     label: 'Serves',    color: '#7aa2f7' },
    { key: 'flow',        label: 'Data Flow',  color: '#9ece6a' },
    { key: 'access',      label: 'Accesses',   color: '#bb9af7' },
    { key: 'triggering',  label: 'Triggers',   color: '#E85D3A' },
    { key: 'association', label: 'Associated', color: '#6b7280' },
  ];
  let activeRels = new Set(REL_TYPES.map(r => r.key));

  // Edge tooltip
  let edgeTooltip = null; // { relLabel, sourceName, targetName, x, y }

  // Node type legend
  const TIERS = [
    { key: 'component',  label: 'Component',  style: 'solid',  color: '#7aa2f7' },
    { key: 'service',    label: 'Service',     style: 'dashed', color: '#60a5fa' },
    { key: 'interface',  label: 'Interface',   style: 'dotted', color: '#38bdf8' },
    { key: 'function',   label: 'Function',    style: 'dotted', color: '#818cf8' },
  ];

  const LIFECYCLE_COLORS = {
    'Production':     '#4ade80',
    'Pilot':          '#60a5fa',
    'Planned':        '#a78bfa',
    'Retiring':       '#fb923c',
    'Decommissioned': '#f87171',
  };

  $: filteredNodes = searchQ
    ? allNodes.filter(n => n.name.toLowerCase().includes(searchQ.toLowerCase()))
    : allNodes;

  // ── Async helpers ──────────────────────────────────────────────────────────

  async function tick() {
    return new Promise(resolve => requestAnimationFrame(resolve));
  }

  // ── Graph build ─────────────────────────────────────────────────────────────

  function applyRelFilter() {
    if (!cy) return;
    cy.edges().forEach(edge => {
      const rk = edge.data('relKey');
      if (activeRels.has(rk)) {
        edge.style({ display: 'element' });
      } else {
        edge.style({ display: 'none' });
      }
    });
  }

  function toggleRel(key) {
    const next = new Set(activeRels);
    if (next.has(key)) {
      if (next.size === 1) return; // keep at least one active
      next.delete(key);
    } else {
      next.add(key);
    }
    activeRels = next;
    applyRelFilter();
  }

  const DAGRE_OPTS = { name: 'dagre', rankDir: 'LR', nodeSep: 28, rankSep: 90, animate: false, padding: 48 };

  function focusNode(id) {
    if (!cy) return;
    selectedId = id;
    const node = cy.$id(id);
    const visible = node.union(node.neighborhood());
    // Hide everything outside the neighbourhood and re-layout the subgraph.
    cy.elements().style('display', 'none').removeClass('selected faded');
    visible.style('display', 'element');
    node.addClass('selected');
    visible.layout({ ...DAGRE_OPTS, animate: true, animationDuration: 350 }).run();
    cy.once('layoutstop', () => cy.fit(visible, 60));
  }

  function clearFocus() {
    if (!cy) return;
    selectedId = null;
    cy.elements().style('display', 'element').removeClass('selected faded');
    applyRelFilter();
    cy.layout({ ...DAGRE_OPTS, animate: true, animationDuration: 350 }).run();
    cy.once('layoutstop', () => cy.fit(undefined, 48));
  }

  onMount(async () => {
    loading = true;
    error   = null;
    empty   = false;
    try {
      const data = await api.get('/workspaces/' + wsId + '/views/' + viewName + '/graph');
      if ((data.nodes || []).length === 0) { empty = true; loading = false; return; }

      allNodes = data.nodes || [];
      loading = false;
      await tick();

      if (isDependency) {
        cy = makeDependencyGraph(container, data, {
          onEdgeHover(edgeData, mouseEvent) {
            edgeTooltip = {
              relLabel:   edgeData.relLabel,
              sourceName: edgeData.sourceName,
              targetName: edgeData.targetName,
              x: mouseEvent.clientX + 14,
              y: mouseEvent.clientY - 10,
            };
          },
          onEdgeLeave() { edgeTooltip = null; },
          onNodeClick(nodeData) {
            selectedId = nodeData ? nodeData.id : null;
          },
        });
      } else {
        cy = makeGraph(container, data);
      }
    } catch (e) {
      error   = e.message;
      loading = false;
    }
  });

  onDestroy(() => { if (cy) cy.destroy(); });

  function fit()      { if (cy) cy.fit(undefined, 48); }
  function relayout() {
    if (!cy) return;
    if (isDependency) {
      clearFocus();
    } else {
      cy.layout({ name: 'cose', animate: true, padding: 40, nodeRepulsion: 8000 }).run();
    }
  }
</script>

<!-- Edge tooltip (fixed, pointer-events: none) -->
{#if edgeTooltip}
  <div
    class="fixed z-50 pointer-events-none bg-popover border border-border rounded-lg shadow-lg px-3 py-2 text-[12px]"
    style="left:{Math.min(edgeTooltip.x, window.innerWidth - 240)}px; top:{Math.min(edgeTooltip.y, window.innerHeight - 100)}px"
  >
    <span class="font-semibold text-foreground">{edgeTooltip.sourceName}</span>
    <span class="text-muted-foreground mx-1.5">→ {edgeTooltip.relLabel} →</span>
    <span class="font-semibold text-foreground">{edgeTooltip.targetName}</span>
  </div>
{/if}

<div class="content" style="padding:0; display:flex; flex-direction:column; height:100%;">

  <!-- Header bar -->
  <div class="flex items-center justify-between gap-4 px-6 pt-5 pb-4 flex-shrink-0 flex-wrap">
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
    <div class="flex items-center gap-2 text-muted-foreground py-6 px-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="mx-6 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">{error}</div>
  {:else if empty}
    <div class="text-center py-16 px-6 text-muted-foreground">
      <div class="text-[40px] mb-3.5">📭</div>
      <p class="text-[14px] leading-relaxed">No application elements — import a model first.</p>
    </div>
  {:else if isDependency}

    <!-- Dependency layout: left panel + graph -->
    <div class="flex flex-1 min-h-0 gap-0">

      <!-- Left panel -->
      <div class="flex flex-col border-r border-border w-52 flex-shrink-0 bg-card/50">

        <!-- Search -->
        <div class="px-3 pt-3 pb-2 flex-shrink-0">
          <input
            type="search"
            bind:value={searchQ}
            placeholder="Find application…"
            class="w-full bg-background border border-border rounded-md px-2.5 py-1.5 text-[12px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
          />
        </div>

        <!-- Node list -->
        <div class="overflow-y-auto flex-1 px-2 pb-2">
          {#if selectedId}
            <button
              class="w-full text-left px-2 py-1 rounded text-[11px] text-primary hover:bg-primary/10 mb-1.5"
              on:click={clearFocus}
            >← Show all</button>
          {/if}
          {#each filteredNodes as node}
            {@const lc = node.lifecycle_status}
            {@const color = LIFECYCLE_COLORS[lc] ?? '#6b7280'}
            <button
              class="w-full text-left flex items-center gap-1.5 px-2 py-1.5 rounded text-[12px] transition-colors {selectedId === node.id ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-muted/50'}"
              on:click={() => selectedId === node.id ? clearFocus() : focusNode(node.id)}
            >
              <span class="size-2 rounded-full flex-shrink-0" style="background:{color}"></span>
              <span class="truncate">{node.name}</span>
            </button>
          {/each}
        </div>

        <!-- Relationship filter -->
        <div class="border-t border-border px-3 py-2.5 flex-shrink-0">
          <div class="text-[10px] font-bold uppercase tracking-wide text-muted-foreground mb-2">Relationships</div>
          {#each REL_TYPES as rt}
            <label class="flex items-center gap-2 py-0.5 cursor-pointer">
              <span
                class="size-2.5 rounded-sm flex-shrink-0 transition-opacity {activeRels.has(rt.key) ? '' : 'opacity-25'}"
                style="background:{rt.color}"
              ></span>
              <input
                type="checkbox"
                class="sr-only"
                checked={activeRels.has(rt.key)}
                on:change={() => toggleRel(rt.key)}
              />
              <span class="text-[11px] {activeRels.has(rt.key) ? 'text-foreground' : 'text-muted-foreground'}">{rt.label}</span>
            </label>
          {/each}
        </div>

        <!-- Node type legend -->
        <div class="border-t border-border px-3 py-2.5 flex-shrink-0">
          <div class="text-[10px] font-bold uppercase tracking-wide text-muted-foreground mb-2">Node types</div>
          {#each TIERS as tier}
            <div class="flex items-center gap-2 py-0.5">
              <span
                class="w-5 h-2.5 flex-shrink-0 rounded-sm"
                style="
                  background: transparent;
                  border: 1.5px {tier.style === 'dotted' ? 'dotted' : tier.style === 'dashed' ? 'dashed' : 'solid'} {tier.color};
                "
              ></span>
              <span class="text-[11px] text-muted-foreground">{tier.label}</span>
            </div>
          {/each}
        </div>
      </div>

      <!-- Graph canvas -->
      <div class="flex-1 flex flex-col min-w-0">
        <div class="cy-container flex-1" bind:this={container}></div>
        <div class="px-4 py-1.5 text-muted-foreground text-[11px] flex-shrink-0 border-t border-border/50">
          Scroll to zoom · Drag to pan · Click a node to highlight · Hover an edge for details
        </div>
      </div>
    </div>

  {:else}
    <!-- Integration map (full width) -->
    <div class="cy-container flex-1 mx-6 mb-4" bind:this={container}></div>
    <div class="px-6 pb-3 text-muted-foreground text-[11px]">
      Scroll to zoom · Drag to pan · Click a node to highlight connections
    </div>
  {/if}
</div>
