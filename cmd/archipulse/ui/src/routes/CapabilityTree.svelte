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
  import CapabilityNode from '../components/flow/CapabilityNode.svelte';
  import AppNode       from '../components/flow/AppNode.svelte';
  import FlowControls  from '../components/flow/FlowControls.svelte';

  let { params = {} } = $props();

  // ── XyFlow state ──────────────────────────────────────────────────────────
  let nodes     = $state([]);
  let edges     = $state([]);
  const nodeTypes = { capNode: CapabilityNode, appNode: AppNode };

  let fitView = $state(null);
  function fit() { fitView?.({ padding: 0.12, duration: 300 }); }

  // ── Data ──────────────────────────────────────────────────────────────────
  let allCaps  = $state([]);   // CapabilityNode[] from API
  let loading  = $state(true);
  let error    = $state(null);

  // ── Selection ─────────────────────────────────────────────────────────────
  let selectedId = $state(null);  // focused root/mid capability
  let searchQ    = $state('');

  // ── Derived lists ─────────────────────────────────────────────────────────
  const rootCaps = $derived(allCaps.filter(c => !c.parent_id));

  const filteredCaps = $derived(
    searchQ
      ? allCaps.filter(c => c.name.toLowerCase().includes(searchQ.toLowerCase()))
      : rootCaps
  );

  // ── Lifecycle colours (shared with AppNode) ───────────────────────────────
  const LIFECYCLE_COLORS = {
    'Production':     '#22c55e',
    'Pilot':          '#3b82f6',
    'Planned':        '#8b5cf6',
    'Retiring':       '#f97316',
    'Decommissioned': '#ef4444',
  };

  // ── Tree helpers ──────────────────────────────────────────────────────────
  function getSubtreeIds(rootId) {
    const ids = new Set([rootId]);
    let changed = true;
    while (changed) {
      changed = false;
      allCaps.forEach(c => {
        if (c.parent_id && ids.has(c.parent_id) && !ids.has(c.id)) {
          ids.add(c.id);
          changed = true;
        }
      });
    }
    return ids;
  }

  function appTier(type) {
    if (type === 'ApplicationComponent') return 'component';
    if (type === 'ApplicationService')   return 'service';
    if (type === 'ApplicationInterface') return 'interface';
    if (type === 'ApplicationFunction')  return 'function';
    return 'other';
  }

  // ── Dagre layout ──────────────────────────────────────────────────────────
  const CAP_W = 195, CAP_H = 54;
  const APP_W = 172, APP_H = 46;

  function applyLayout(focusId = null) {
    const capSet = focusId ? getSubtreeIds(focusId) : null;

    const g = new dagre.graphlib.Graph();
    g.setDefaultEdgeLabel(() => ({}));
    g.setGraph({ rankdir: 'LR', nodesep: 32, ranksep: 110, marginx: 60, marginy: 60 });

    // Capability nodes
    allCaps.forEach(cap => {
      if (capSet && !capSet.has(cap.id)) return;
      g.setNode(cap.id, { width: CAP_W, height: CAP_H });
    });

    // Composition edges (parent → child capability)
    allCaps.forEach(cap => {
      if (!cap.parent_id) return;
      if (capSet && (!capSet.has(cap.parent_id) || !capSet.has(cap.id))) return;
      if (g.hasNode(cap.parent_id) && g.hasNode(cap.id)) {
        g.setEdge(cap.parent_id, cap.id);
      }
    });

    // App nodes + serving edges (capability → app)
    const appEntries = [];
    allCaps.forEach(cap => {
      if (capSet && !capSet.has(cap.id)) return;
      (cap.supporting_apps ?? []).forEach(app => {
        const nodeId = `app_${cap.id}_${app.id}`;
        g.setNode(nodeId, { width: APP_W, height: APP_H });
        g.setEdge(cap.id, nodeId);
        appEntries.push({ nodeId, app, capId: cap.id });
      });
    });

    dagre.layout(g);

    // Build flow nodes
    const flowNodes = [];

    allCaps.forEach(cap => {
      if (capSet && !capSet.has(cap.id)) return;
      const gn  = g.node(cap.id);
      const appCount = (cap.supporting_apps ?? []).length;
      flowNodes.push({
        id:        cap.id,
        type:      'capNode',
        draggable: false,
        position:  gn ? { x: gn.x - CAP_W / 2, y: gn.y - CAP_H / 2 } : { x: 0, y: 0 },
        data:      { label: cap.name, appCount },
      });
    });

    appEntries.forEach(({ nodeId, app }) => {
      const gn   = g.node(nodeId);
      const tier = appTier(app.type);
      flowNodes.push({
        id:        nodeId,
        type:      'appNode',
        draggable: false,
        position:  gn ? { x: gn.x - APP_W / 2, y: gn.y - APP_H / 2 } : { x: 0, y: 0 },
        data:      { label: app.name, badge: app.type.replace('Application', ''), tier, lifecycle: app.lifecycle_status },
      });
    });

    // Build flow edges
    const flowEdges = [];

    // Composition edges
    allCaps.forEach(cap => {
      if (!cap.parent_id) return;
      if (capSet && (!capSet.has(cap.parent_id) || !capSet.has(cap.id))) return;
      flowEdges.push({
        id:        `comp_${cap.parent_id}_${cap.id}`,
        source:    cap.parent_id,
        target:    cap.id,
        style:     'stroke:#c09040; stroke-width:1.8px;',
        markerEnd: { type: 'arrowclosed', color: '#c09040', width: 12, height: 12 },
      });
    });

    // Serving edges (cap → app)
    appEntries.forEach(({ nodeId, capId }) => {
      flowEdges.push({
        id:        `srv_${nodeId}`,
        source:    capId,
        target:    nodeId,
        style:     'stroke:#4a6fa555; stroke-width:1.4px; stroke-dasharray:5,4;',
        markerEnd: { type: 'arrowclosed', color: '#4a6fa5', width: 11, height: 11 },
      });
    });

    nodes = flowNodes;
    edges = flowEdges;
    setTimeout(fit, 80);
  }

  function selectCap(id) {
    selectedId = id === selectedId ? null : id;
    applyLayout(selectedId);
  }

  function clearSelection() {
    selectedId = null;
    applyLayout(null);
  }

  // ── Load ──────────────────────────────────────────────────────────────────
  onMount(async () => {
    try {
      const data = await api.get('/workspaces/' + params.wsId + '/views/capability-tree/tree');
      allCaps = data.nodes ?? [];
      applyLayout(null);
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });
</script>

<!-- full-viewport container, same pattern as DependencyGraphView -->
<div style="position:fixed; top:var(--nav-h); left:var(--sidebar-w); right:0; bottom:0; display:flex; flex-direction:column; overflow:hidden; background:var(--bg);">

  <!-- Header -->
  <div class="flex items-center justify-between gap-4 px-6 pt-5 pb-4 flex-shrink-0">
    <div>
      <h1 class="text-[18px] font-semibold">Capability Tree</h1>
      <div class="text-muted-foreground text-[13px] mt-0.5">Business capabilities and supporting applications</div>
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
  {:else if allCaps.length === 0}
    <div class="text-center py-16 px-6 text-muted-foreground">
      <div class="text-[40px] mb-3.5">◈</div>
      <p class="text-[14px]">No Capability elements found — import or create Capability elements first.</p>
    </div>
  {:else}
    <div class="flex flex-1 min-h-0">

      <!-- Left panel -->
      <div class="flex flex-col border-r border-border w-52 flex-shrink-0 bg-card/50 overflow-hidden min-h-0 h-full">

        <!-- Search -->
        <div class="px-3 pt-3 pb-2 flex-shrink-0">
          <input type="search" bind:value={searchQ} placeholder="Find capability…"
            class="w-full bg-background border border-border rounded-md px-2.5 py-1.5 text-[12px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
        </div>

        <!-- "All" reset button when something is selected -->
        {#if selectedId}
          <div class="px-3 pb-2 flex-shrink-0">
            <button
              class="w-full text-left px-2 py-1 rounded text-[11px] text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
              onclick={clearSelection}
            >← Show all</button>
          </div>
        {/if}

        <!-- Section label -->
        <div class="px-3 pb-1 flex-shrink-0">
          <span class="text-[10px] font-bold uppercase tracking-wide text-muted-foreground">
            {searchQ ? 'Results' : 'Capabilities'}
          </span>
        </div>

        <!-- Capability list (scrollable) -->
        <div class="overflow-y-auto flex-1 px-2 pb-2">
          {#each filteredCaps as cap}
            {@const isRoot = !cap.parent_id}
            {@const selected = cap.id === selectedId}
            <button
              class="w-full text-left flex items-center gap-1.5 px-2 py-1.5 rounded text-[12px] transition-colors {selected ? 'bg-amber-500/10 text-amber-300' : 'text-foreground hover:bg-muted/50'}"
              onclick={() => selectCap(cap.id)}
            >
              <span class="size-2 rounded-sm flex-shrink-0 {isRoot ? 'bg-amber-500/70' : 'bg-amber-500/30'}"></span>
              <span class="truncate {isRoot ? 'font-medium' : ''}">{cap.name}</span>
            </button>
          {/each}
        </div>

        <!-- Stats -->
        <div class="border-t border-border px-3 py-2.5 flex-shrink-0">
          <div class="text-[10px] text-muted-foreground">
            {rootCaps.length} root · {allCaps.length} total
          </div>
        </div>
      </div>

      <!-- Flow canvas -->
      <div class="flex-1 min-w-0" style="background:#0d1526;">
        <SvelteFlow
          {nodes}
          {edges}
          {nodeTypes}
          nodesDraggable={false}
          fitView
          minZoom={0.05}
          maxZoom={3}
          proOptions={{ hideAttribution: true }}
          style="background:#0d1526; width:100%; height:100%;"
        >
          <FlowControls onReady={(fn) => { fitView = fn; }} />

          <Controls showInteractive={false} style="background:#122040; border:1px solid #1e3a5f; border-radius:8px;" />

          <MiniMap
            position="bottom-right"
            style="background:#122040; border:1px solid #1e3a5f; border-radius:8px; margin-bottom:48px;"
            nodeColor={(n) => n.type === 'capNode' ? '#c09040' : (LIFECYCLE_COLORS[n.data?.lifecycle] ?? '#4a6fa5')}
            maskColor="rgba(0,0,0,0.55)"
          />

          <Background variant={BackgroundVariant.Dots} gap={22} size={1} color="#112050" />

          <!-- Legend -->
          <Panel position="bottom-left">
            <div class="rounded-lg px-3.5 py-3 text-[11px]" style="background:rgba(13,21,38,0.94); border:1px solid #1e3a5f; min-width:140px;">
              <div class="text-[10px] font-bold uppercase tracking-wide mb-2" style="color:#6b7280;">Node type</div>
              <div class="flex items-center gap-2 mb-1.5">
                <div style="width:20px; height:12px; border-radius:3px; border:2px solid #e0af68; background:#201808; flex-shrink:0;"></div>
                <span style="color:#fcd990; font-weight:600;">Capability</span>
              </div>
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

              <div class="text-[10px] font-bold uppercase tracking-wide mt-3 mb-1.5" style="color:#6b7280;">Edges</div>
              <div class="flex items-center gap-2 mb-1.5">
                <div style="width:20px; height:2px; background:#c09040; flex-shrink:0;"></div>
                <span style="color:#8b949e;">Composition</span>
              </div>
              <div class="flex items-center gap-2">
                <div style="width:20px; height:1px; border-top:1.5px dashed #4a6fa5; flex-shrink:0;"></div>
                <span style="color:#8b949e;">Supports</span>
              </div>
            </div>
          </Panel>
        </SvelteFlow>
      </div>

    </div>
  {/if}
</div>
