<script>
  import { onMount } from 'svelte';
  import {
    SvelteFlow,
    Controls,
    MiniMap,
    Background,
    BackgroundVariant,
    Panel,
  } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';
  import dagre from '@dagrejs/dagre';
  import { api } from '../lib/api.js';
  import AppNode from '../components/flow/AppNode.svelte';
  import FlowControls from '../components/flow/FlowControls.svelte';

  let { params = {} } = $props();

  // ── XyFlow state ──────────────────────────────────────────────────────────
  let nodes = $state([]);
  let edges = $state([]);
  const nodeTypes = { appNode: AppNode };

  // fitView comes from useSvelteFlow() via child FlowControls
  let fitView = $state(null);
  function fit() { fitView?.({ padding: 0.12, duration: 300 }); }

  // ── Data ──────────────────────────────────────────────────────────────────
  let allNodes = $state([]);
  let allEdges = $state([]);
  let loading  = $state(true);
  let error    = $state(null);

  // ── Panel ─────────────────────────────────────────────────────────────────
  let searchQ      = $state('');
  let selectedApps = $state(new Set()); // multi-select set of app IDs

  // ── Edge tooltip ──────────────────────────────────────────────────────────
  let tooltip = $state(null);

  // ── Relationship filters ──────────────────────────────────────────────────
  const REL_TYPES = [
    { key: 'serving',     label: 'Serves',      color: '#7aa2f7' },
    { key: 'flow',        label: 'Data Flow',    color: '#9ece6a' },
    { key: 'access',      label: 'Accesses',     color: '#bb9af7' },
    { key: 'triggering',  label: 'Triggers',     color: '#E85D3A' },
    { key: 'association', label: 'Associated',   color: '#6b7280' },
  ];
  let activeRels = $state(new Set(REL_TYPES.map(r => r.key)));

  const LIFECYCLE_COLORS = {
    'Production':     '#22c55e',
    'Pilot':          '#3b82f6',
    'Planned':        '#8b5cf6',
    'Retiring':       '#f97316',
    'Decommissioned': '#ef4444',
  };

  const REL_META = {
    serving:     { label: 'Serves',         color: '#7aa2f7', animated: false },
    flow:        { label: 'Data Flow',       color: '#9ece6a', animated: true  },
    access:      { label: 'Accesses',        color: '#bb9af7', animated: true  },
    triggering:  { label: 'Triggers',        color: '#E85D3A', animated: false },
    association: { label: 'Associated with', color: '#6b7280', animated: false },
  };

  function relKey(rel) {
    const r = (rel || '').toLowerCase();
    if (r.includes('serving'))    return 'serving';
    if (r.includes('flow'))       return 'flow';
    if (r.includes('access'))     return 'access';
    if (r.includes('triggering')) return 'triggering';
    return 'association';
  }

  function nodeTier(type) {
    if (type === 'ApplicationComponent') return 'component';
    if (type === 'ApplicationService')   return 'service';
    if (type === 'ApplicationInterface') return 'interface';
    if (type === 'ApplicationFunction')  return 'function';
    return 'other';
  }

  // ── Dagre layout ──────────────────────────────────────────────────────────
  function layoutNodes(rawNodes, rawEdges, visibleIds = null) {
    const g = new dagre.graphlib.Graph();
    g.setDefaultEdgeLabel(() => ({}));
    g.setGraph({ rankdir: 'LR', nodesep: 36, ranksep: 120, marginx: 60, marginy: 60 });

    const nw = t => t === 'component' ? 190 : 158;
    const nh = t => t === 'component' ? 52  : 42;
    const nodeSet = visibleIds ? new Set(visibleIds) : null;

    rawNodes.forEach(n => {
      if (nodeSet && !nodeSet.has(n.id)) return;
      const tier = nodeTier(n.type);
      g.setNode(n.id, { width: nw(tier), height: nh(tier) });
    });

    rawEdges.forEach(e => {
      if (!activeRels.has(relKey(e.relationship))) return;
      if (nodeSet && (!nodeSet.has(e.source) || !nodeSet.has(e.target))) return;
      if (e.source === e.target) return;
      if (g.hasNode(e.source) && g.hasNode(e.target)) g.setEdge(e.source, e.target);
    });

    dagre.layout(g);

    const nameById = Object.fromEntries(rawNodes.map(n => [n.id, n.name]));

    const flowNodes = rawNodes
      .filter(n => !nodeSet || nodeSet.has(n.id))
      .map(n => {
        const gn   = g.node(n.id);
        const tier = nodeTier(n.type);
        return {
          id:       n.id,
          type:     'appNode',
          draggable: false,
          position: gn ? { x: gn.x - nw(tier) / 2, y: gn.y - nh(tier) / 2 } : { x: 0, y: 0 },
          data:     { label: n.name, badge: n.type.replace('Application', ''), tier, lifecycle: n.lifecycle_status },
        };
      });

    const flowEdges = rawEdges
      .filter(e => {
        if (!activeRels.has(relKey(e.relationship))) return false;
        if (nodeSet && (!nodeSet.has(e.source) || !nodeSet.has(e.target))) return false;
        return e.source !== e.target;
      })
      .map(e => {
        const rk   = relKey(e.relationship);
        const meta = REL_META[rk];
        return {
          id:        e.id,
          source:    e.source,
          target:    e.target,
          animated:  meta.animated,
          style:     `stroke:${meta.color}; stroke-width:1.8px;`,
          markerEnd: { type: 'arrowclosed', color: meta.color, width: 14, height: 14 },
          data:      { relLabel: meta.label, sourceName: nameById[e.source] ?? '', targetName: nameById[e.target] ?? '', rk },
        };
      });

    return { flowNodes, flowEdges };
  }

  function applyLayout(visibleIds = null) {
    const { flowNodes, flowEdges } = layoutNodes(allNodes, allEdges, visibleIds);
    nodes = flowNodes;
    edges = flowEdges;
    // Fit after a short delay to ensure SvelteFlow has rendered new positions.
    setTimeout(fit, 80);
  }

  // ── Multi-selection helpers ───────────────────────────────────────────────
  function neighboursOfMany(ids) {
    const result = new Set(ids);
    allEdges.forEach(e => {
      if (ids.has(e.source)) result.add(e.target);
      if (ids.has(e.target)) result.add(e.source);
    });
    return result;
  }

  function currentVisibleIds() {
    return selectedApps.size > 0 ? neighboursOfMany(selectedApps) : null;
  }

  function toggleApp(id) {
    const next = new Set(selectedApps);
    if (next.has(id)) next.delete(id); else next.add(id);
    selectedApps = next;
    applyLayout(next.size > 0 ? neighboursOfMany(next) : null);
  }

  function clearFocus() {
    selectedApps = new Set();
    applyLayout(null);
  }

  // ── Rel filter toggle ─────────────────────────────────────────────────────
  function toggleRel(key) {
    const next = new Set(activeRels);
    if (next.has(key)) { if (next.size === 1) return; next.delete(key); }
    else next.add(key);
    activeRels = next;
    applyLayout(currentVisibleIds());
  }

  // ── Edge events ───────────────────────────────────────────────────────────
  // @xyflow/svelte uses onedgepointerenter/leave; callback receives { event, edge }
  function onEdgePointerEnter({ event, edge }) {
    const d = edge?.data;
    if (!d) return;
    tooltip = { text: `${d.sourceName}  →  ${d.relLabel}  →  ${d.targetName}`, x: event.clientX, y: event.clientY };
  }
  function onEdgePointerLeave() { tooltip = null; }

  // ── Panel list ────────────────────────────────────────────────────────────
  const filteredNodes = $derived(
    searchQ ? allNodes.filter(n => n.name.toLowerCase().includes(searchQ.toLowerCase())) : allNodes
  );

  // ── Load ──────────────────────────────────────────────────────────────────
  onMount(async () => {
    try {
      const data = await api.get('/workspaces/' + params.wsId + '/views/application-dependency/graph');
      allNodes = data.nodes ?? [];
      allEdges = data.edges ?? [];
      applyLayout();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });
</script>

<!-- Edge tooltip -->
{#if tooltip}
  <div class="fixed z-50 pointer-events-none rounded-lg shadow-xl px-3 py-2 text-[12px] text-foreground max-w-sm"
       style="left:{Math.min(tooltip.x + 16, window.innerWidth - 320)}px; top:{tooltip.y - 40}px; background:rgba(22,27,34,0.95); border:1px solid #30363d;">
    {tooltip.text}
  </div>
{/if}

<div style="position:fixed; top:var(--nav-h); left:var(--sidebar-w); right:0; bottom:0; display:flex; flex-direction:column; overflow:hidden; background:var(--bg);">

  <!-- Header -->
  <div class="flex items-center justify-between gap-4 px-6 pt-5 pb-4 flex-shrink-0">
    <div>
      <h1 class="text-[18px] font-semibold">Dependency Graph</h1>
      <div class="text-muted-foreground text-[13px] mt-0.5">Interactive application dependency map</div>
    </div>
    <button
      class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] hover:bg-muted transition-colors"
      onclick={fit}
    >⊡ Fit</button>
  </div>

  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6 px-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="mx-6 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">{error}</div>
  {:else if allNodes.length === 0}
    <div class="text-center py-16 px-6 text-muted-foreground">
      <div class="text-[40px] mb-3.5">📭</div>
      <p class="text-[14px]">No application elements — import a model first.</p>
    </div>
  {:else}
    <div class="flex flex-1 min-h-0">

      <!-- Left panel -->
      <div class="flex flex-col border-r border-border w-52 flex-shrink-0 bg-card/50 overflow-hidden min-h-0 h-full">

        <!-- Search -->
        <div class="px-3 pt-3 pb-2 flex-shrink-0">
          <input type="search" bind:value={searchQ} placeholder="Find application…"
            class="w-full bg-background border border-border rounded-md px-2.5 py-1.5 text-[12px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
        </div>

        <!-- Selected chips -->
        {#if selectedApps.size > 0}
          <div class="px-2 pb-2 flex-shrink-0 flex flex-wrap gap-1 border-b border-border">
            {#each [...selectedApps] as id}
              {@const node = allNodes.find(n => n.id === id)}
              {@const color = LIFECYCLE_COLORS[node?.lifecycle_status] ?? '#6b7280'}
              <span class="inline-flex items-center gap-1 pl-1.5 pr-1 py-0.5 rounded-full text-[11px] font-medium"
                    style="background:{color}22; border:1px solid {color}55; color:{color};">
                <span class="max-w-[120px] truncate">{node?.name ?? id}</span>
                <button class="flex-shrink-0 rounded-full hover:bg-white/20 p-0.5 leading-none"
                        onclick={() => toggleApp(id)}>×</button>
              </span>
            {/each}
            <button class="text-[10px] text-muted-foreground hover:text-foreground px-1" onclick={clearFocus}>clear</button>
          </div>
        {/if}

        <!-- App list (scrollable) -->
        <div class="overflow-y-auto flex-1 px-2 py-1">
          {#each filteredNodes as node}
            {@const color = LIFECYCLE_COLORS[node.lifecycle_status] ?? '#6b7280'}
            {@const selected = selectedApps.has(node.id)}
            <button
              class="w-full text-left flex items-center gap-1.5 px-2 py-1.5 rounded text-[12px] transition-colors {selected ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-muted/50'}"
              onclick={() => toggleApp(node.id)}
            >
              <span class="size-2 rounded-full flex-shrink-0" style="background:{color}"></span>
              <span class="truncate">{node.name}</span>
            </button>
          {/each}
        </div>

        <!-- Relationship filters — pinned at bottom -->
        <div class="border-t border-border px-3 py-2.5 flex-shrink-0">
          <div class="text-[10px] font-bold uppercase tracking-wide text-muted-foreground mb-1.5">Relationships</div>
          {#each REL_TYPES as rt}
            <label class="flex items-center gap-2 py-0.5 cursor-pointer select-none">
              <span class="size-2.5 rounded-sm flex-shrink-0 transition-opacity {activeRels.has(rt.key) ? '' : 'opacity-20'}" style="background:{rt.color}"></span>
              <input type="checkbox" class="sr-only" checked={activeRels.has(rt.key)} onchange={() => toggleRel(rt.key)} />
              <span class="text-[11px] {activeRels.has(rt.key) ? 'text-foreground' : 'text-muted-foreground'}">{rt.label}</span>
            </label>
          {/each}
        </div>
      </div>

      <!-- Flow canvas -->
      <div class="flex-1 min-w-0" style="background:#161b22;">
        <SvelteFlow
          {nodes}
          {edges}
          {nodeTypes}
          nodesDraggable={false}
          fitView
          minZoom={0.08}
          maxZoom={3}
          proOptions={{ hideAttribution: true }}
          onedgepointerenter={onEdgePointerEnter}
          onedgepointerleave={onEdgePointerLeave}
          style="background:#161b22; width:100%; height:100%;"
        >
          <!-- Registers fitView from inside the SvelteFlow context -->
          <FlowControls onReady={(fn) => { fitView = fn; }} />

          <Controls showInteractive={false} style="background:#1c2128; border:1px solid #30363d; border-radius:8px;" />

          <MiniMap
            position="bottom-right"
            style="background:#1c2128; border:1px solid #30363d; border-radius:8px; margin-bottom:48px;"
            nodeColor={(n) => LIFECYCLE_COLORS[n.data?.lifecycle] ?? '#4a6fa5'}
            maskColor="rgba(0,0,0,0.55)"
          />

          <Background variant={BackgroundVariant.Dots} gap={22} size={1} color="#21262d" />

          <!-- Legend panel — rendered inside SvelteFlow as an overlay -->
          <Panel position="bottom-left">
            <div class="rounded-lg px-3.5 py-3 text-[11px]" style="background:rgba(22,27,34,0.92); border:1px solid #30363d; min-width:140px;">
              <div class="text-[10px] font-bold uppercase tracking-wide mb-2" style="color:#6b7280;">Node type</div>
              {#each [
                { label: 'Component', bs: 'solid',  color: '#93b4f0', bold: true  },
                { label: 'Service',   bs: 'dashed', color: '#7aabf7', bold: false },
                { label: 'Interface', bs: 'dotted', color: '#5ebbe8', bold: false },
                { label: 'Function',  bs: 'dotted', color: '#a89cf7', bold: false },
              ] as t}
                <div class="flex items-center gap-2 mb-1.5">
                  <div style="width:20px; height:12px; border-radius:3px; border:1.5px {t.bs} {t.color}; background:transparent; flex-shrink:0;"></div>
                  <span style="color:{t.bold ? '#cdd9e5' : '#8b949e'}; font-weight:{t.bold ? 600 : 400};">{t.label}</span>
                </div>
              {/each}

              <div class="text-[10px] font-bold uppercase tracking-wide mt-3 mb-2" style="color:#6b7280;">Lifecycle</div>
              {#each Object.entries(LIFECYCLE_COLORS) as [lc, color]}
                <div class="flex items-center gap-2 mb-1.5">
                  <span style="width:8px; height:8px; border-radius:50%; background:{color}; flex-shrink:0; display:inline-block;"></span>
                  <span style="color:#8b949e;">{lc}</span>
                </div>
              {/each}
            </div>
          </Panel>
        </SvelteFlow>
      </div>

    </div>
  {/if}
</div>
