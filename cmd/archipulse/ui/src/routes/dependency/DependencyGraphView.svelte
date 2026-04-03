<script>
  import { onMount } from 'svelte';
  import {
    SvelteFlow,
    Controls,
    MiniMap,
    Background,
    BackgroundVariant,
  } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';
  import dagre from '@dagrejs/dagre';
  import { api } from '../../lib/api.js';
  import AppNode from './AppNode.svelte';

  let { params = {} } = $props();
  $effect(() => { wsId = params.wsId; });
  let wsId = $state('');

  // ── XyFlow state ─────────────────────────────────────────────────────────
  let nodes = $state([]);
  let edges = $state([]);
  const nodeTypes = { appNode: AppNode };

  // fitView called via the flow instance obtained from oninit
  let flowInstance = $state(null);
  function fit() { flowInstance?.fitView({ padding: 0.12, duration: 300 }); }

  // ── Data ──────────────────────────────────────────────────────────────────
  let allNodes   = $state([]);
  let allEdges   = $state([]);
  let loading    = $state(true);
  let error      = $state(null);

  // ── Panel state ───────────────────────────────────────────────────────────
  let searchQ    = $state('');
  let selectedId = $state(null);

  // ── Tooltip ───────────────────────────────────────────────────────────────
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
    'Production':     '#4ade80',
    'Pilot':          '#60a5fa',
    'Planned':        '#a78bfa',
    'Retiring':       '#fb923c',
    'Decommissioned': '#f87171',
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
    g.setGraph({ rankdir: 'LR', nodesep: 32, ranksep: 110, marginx: 48, marginy: 48 });

    const nw = (tier) => tier === 'component' ? 180 : 150;
    const nh = (tier) => tier === 'component' ? 48  : 38;
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
          style:     `stroke:${meta.color}; stroke-width:1.5px;`,
          markerEnd: { type: 'arrowclosed', color: meta.color, width: 16, height: 16 },
          data:      { relLabel: meta.label, sourceName: nameById[e.source] ?? '', targetName: nameById[e.target] ?? '', rk },
        };
      });

    return { flowNodes, flowEdges };
  }

  function applyLayout(visibleIds = null) {
    const { flowNodes, flowEdges } = layoutNodes(allNodes, allEdges, visibleIds);
    nodes = flowNodes;
    edges = flowEdges;
    setTimeout(fit, 60);
  }

  // ── Focus / clear ─────────────────────────────────────────────────────────
  function neighbourIds(id) {
    const ids = new Set([id]);
    allEdges.forEach(e => {
      if (e.source === id) ids.add(e.target);
      if (e.target === id) ids.add(e.source);
    });
    return ids;
  }

  function focusNode(id) {
    selectedId = id;
    applyLayout(neighbourIds(id));
  }

  function clearFocus() {
    selectedId = null;
    applyLayout();
  }

  // ── Rel filter toggle ─────────────────────────────────────────────────────
  function toggleRel(key) {
    const next = new Set(activeRels);
    if (next.has(key)) { if (next.size === 1) return; next.delete(key); }
    else next.add(key);
    activeRels = next;
    applyLayout(selectedId ? neighbourIds(selectedId) : null);
  }

  // ── Edge events ────────────────────────────────────────────────────────────
  function onEdgeMouseEnter(e) {
    const d = e.detail?.edge?.data;
    if (!d) return;
    tooltip = { text: `${d.sourceName} → ${d.relLabel} → ${d.targetName}`, x: e.detail?.event?.clientX ?? 0, y: e.detail?.event?.clientY ?? 0 };
  }
  function onEdgeMouseMove(e) {
    if (tooltip) tooltip = { ...tooltip, x: e.detail?.event?.clientX ?? tooltip.x, y: e.detail?.event?.clientY ?? tooltip.y };
  }
  function onEdgeMouseLeave() { tooltip = null; }

  // ── Panel filter ──────────────────────────────────────────────────────────
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

{#if tooltip}
  <div class="fixed z-50 pointer-events-none bg-popover border border-border rounded-lg shadow-lg px-3 py-2 text-[12px] text-foreground max-w-xs"
       style="left:{Math.min(tooltip.x + 14, window.innerWidth - 280)}px; top:{tooltip.y - 36}px">
    {tooltip.text}
  </div>
{/if}

<div class="content" style="padding:0; display:flex; flex-direction:column; height:100%;">

  <!-- Header -->
  <div class="flex items-center justify-between gap-4 px-6 pt-5 pb-4 flex-shrink-0">
    <div>
      <h1 class="text-[18px] font-semibold">Dependency Graph</h1>
      <div class="text-muted-foreground text-[13px] mt-0.5">Interactive application dependency map</div>
    </div>
    <button class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] hover:bg-muted transition-colors" onclick={fit}>⊡ Fit</button>
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
      <div class="flex flex-col border-r border-border w-52 flex-shrink-0 bg-card/50 overflow-hidden">
        <div class="px-3 pt-3 pb-2 flex-shrink-0">
          <input type="search" bind:value={searchQ} placeholder="Find application…"
            class="w-full bg-background border border-border rounded-md px-2.5 py-1.5 text-[12px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
        </div>

        <div class="overflow-y-auto flex-1 px-2 pb-2">
          {#if selectedId}
            <button class="w-full text-left px-2 py-1 rounded text-[11px] text-primary hover:bg-primary/10 mb-1.5" onclick={clearFocus}>← Show all</button>
          {/if}
          {#each filteredNodes as node}
            {@const color = LIFECYCLE_COLORS[node.lifecycle_status] ?? '#6b7280'}
            <button
              class="w-full text-left flex items-center gap-1.5 px-2 py-1.5 rounded text-[12px] transition-colors {selectedId === node.id ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-muted/50'}"
              onclick={() => selectedId === node.id ? clearFocus() : focusNode(node.id)}
            >
              <span class="size-2 rounded-full flex-shrink-0" style="background:{color}"></span>
              <span class="truncate">{node.name}</span>
            </button>
          {/each}
        </div>

        <!-- Rel filters -->
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
      <div class="flex-1 relative" style="background:#0d1117;">
        <SvelteFlow
          {nodes}
          {edges}
          {nodeTypes}
          fitView
          minZoom={0.1}
          maxZoom={3}
          proOptions={{ hideAttribution: true }}
          oninit={(instance) => { flowInstance = instance; }}
          onedgemouseenter={onEdgeMouseEnter}
          onedgemousemove={onEdgeMouseMove}
          onedgemouseleave={onEdgeMouseLeave}
          style="background:#0d1117;"
        >
          <Controls position="bottom-right" style="background:#1a1f2e; border-color:#2a2d3e;" />
          <MiniMap
            position="bottom-right"
            style="background:#1a1f2e; border:1px solid #2a2d3e; margin-bottom:44px;"
            nodeColor={(n) => LIFECYCLE_COLORS[n.data?.lifecycle] ?? '#3d59a1'}
            maskColor="rgba(0,0,0,0.6)"
          />
          <Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#1e2130" />
        </SvelteFlow>

        <!-- Legend: bottom-left corner over the canvas -->
        <div class="absolute bottom-4 left-4 z-10 pointer-events-none rounded-lg px-3.5 py-3 text-[11px]"
             style="background:rgba(13,17,23,0.88); border:1px solid #2a2d3e;">
          <div class="text-[10px] font-bold uppercase tracking-wide text-muted-foreground mb-2">Node type</div>
          {#each [
            { label: 'Component', bStyle: 'solid',  color: '#7aa2f7', bold: true  },
            { label: 'Service',   bStyle: 'dashed', color: '#60a5fa', bold: false },
            { label: 'Interface', bStyle: 'dotted', color: '#38bdf8', bold: false },
            { label: 'Function',  bStyle: 'dotted', color: '#818cf8', bold: false },
          ] as t}
            <div class="flex items-center gap-2 mb-1">
              <div class="flex-shrink-0 rounded-sm" style="width:18px; height:11px; border:1.5px {t.bStyle} {t.color}; background:transparent;"></div>
              <span style="color:{t.bold ? '#c9d1d9' : '#8b8fa8'}; font-weight:{t.bold ? 600 : 400};">{t.label}</span>
            </div>
          {/each}
          <div class="text-[10px] font-bold uppercase tracking-wide text-muted-foreground mt-3 mb-2">Lifecycle</div>
          {#each Object.entries(LIFECYCLE_COLORS) as [lc, color]}
            <div class="flex items-center gap-2 mb-1">
              <span class="size-2 rounded-full flex-shrink-0" style="background:{color}"></span>
              <span class="text-muted-foreground">{lc}</span>
            </div>
          {/each}
        </div>
      </div>

    </div>
  {/if}
</div>
