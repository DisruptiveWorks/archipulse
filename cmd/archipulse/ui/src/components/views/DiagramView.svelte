<script>
  import { push } from 'svelte-spa-router';
  import { SvelteFlow, Controls, Background, MiniMap, Panel, ConnectionMode } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';

  import { api } from '../../lib/api.js';
  import BackButton from '../BackButton.svelte';
  import ArchiMateNode from '../diagram/ArchiMateNode.svelte';
  import ValueStreamNode from '../diagram/ValueStreamNode.svelte';
  import ArchiMateEdge from '../diagram/ArchiMateEdge.svelte';
  import { getColor } from '../diagram/archimate-icons.js';

  export let params = {};
  export let embedded = false; // true when used inside DiagramList split layout

  $: wsId = params.wsId;
  $: diagId = params.diagId;

  let data = null;
  let loading = true;
  let error = null;

  let nodes = [];
  let edges = [];

  const nodeTypes = { archimate: ArchiMateNode, valuestream: ValueStreamNode };
  const edgeTypes = { archimate: ArchiMateEdge };

  $: if (diagId) load();

  // ── MiniMap color ──────────────────────────────────────────────────────────

  function minimapColor(n) {
    if (!n.data) return '#D1D5DB';
    return getColor(n.data.elementType).stroke;
  }

  // ── Load ──────────────────────────────────────────────────────────────────

  async function load() {
    loading = true;
    error = null;
    data = null;
    nodes = [];
    edges = [];
    try {
      data = await api.get('/workspaces/' + wsId + '/diagrams/' + diagId + '/render');

      const rawNodes = data.nodes || [];

      // Assign a unique instance key to each node.
      // The same element_id can appear multiple times (e.g. DBMS in Partition 1, 2, and Distributed Servers).
      // XY Flow requires unique IDs, so we suffix duplicates with _1, _2, etc.
      const elemIdCount = {};
      const nodeInstanceId = rawNodes.map(n => {
        const c = elemIdCount[n.element_id] || 0;
        elemIdCount[n.element_id] = c + 1;
        return c === 0 ? n.element_id : `${n.element_id}_${c}`;
      });

      // Build a lookup of raw node data by element_id (first occurrence wins — used for edge bounds).
      const nodeById = {};
      for (const n of rawNodes) {
        if (!nodeById[n.element_id]) nodeById[n.element_id] = n;
      }

      // Build a lookup of raw node data by node_id (OEF diagram identifier — always unique).
      const nodeByNodeId = {};
      for (const n of rawNodes) {
        if (n.node_id) nodeByNodeId[n.node_id] = n;
      }

      // Build a map of element_id → all instances [{iid, idx, node}] for parent resolution.
      // When the same element appears multiple times, spatial containment picks the right parent.
      const instancesByElemId = {};
      rawNodes.forEach((n, i) => {
        const iid = nodeInstanceId[i];
        if (!instancesByElemId[n.element_id]) instancesByElemId[n.element_id] = [];
        instancesByElemId[n.element_id].push({ iid, idx: i, node: n });
      });

      // Given a child node, return the correct parent instance ({iid, node}) using spatial containment.
      function resolveParentInstance(child) {
        if (!child.parent_element_id) return null;
        const candidates = instancesByElemId[child.parent_element_id];
        if (!candidates || candidates.length === 0) return null;
        if (candidates.length === 1) return candidates[0];
        // Multiple parent instances: find the one whose bounds contain the child.
        for (const c of candidates) {
          const p = c.node;
          if (child.x >= p.x && child.y >= p.y &&
              child.x + child.w <= p.x + p.w &&
              child.y + child.h <= p.y + p.h) {
            return c;
          }
        }
        return candidates[0]; // fallback
      }

      // Nodes that appear as a parent_element_id of another node are containers.
      const containerIds = new Set(
        rawNodes.filter(n => n.parent_element_id).map(n => n.parent_element_id)
      );

      // XY Flow requires parent nodes to appear before their children in the array.
      // Sort by topological order using instance IDs.
      const sortedIndices = [];
      const seenInstances = new Set();

      function visit(idx) {
        const iid = nodeInstanceId[idx];
        if (seenInstances.has(iid)) return;
        const n = rawNodes[idx];
        if (n.parent_element_id) {
          const parentInst = resolveParentInstance(n);
          if (parentInst && !seenInstances.has(parentInst.iid)) {
            visit(parentInst.idx);
          }
        }
        seenInstances.add(iid);
        sortedIndices.push(idx);
      }
      rawNodes.forEach((_, i) => visit(i));

      nodes = sortedIndices.map(i => {
        const n = rawNodes[i];
        const iid = nodeInstanceId[i];
        const parentInst = resolveParentInstance(n);
        const parentIid = parentInst?.iid || null;
        const parent = parentInst?.node || null;
        const isContainer = containerIds.has(n.element_id);
        const isVS = n.element_type === 'ValueStream';

        // XY Flow child positions must be relative to their direct parent.
        const position = parent
          ? { x: n.x - parent.x, y: n.y - parent.y }
          : { x: n.x, y: n.y };

        return {
          id: iid,
          type: isVS ? 'valuestream' : 'archimate',
          position,
          ...(parentIid ? { parentId: parentIid, extent: 'parent' } : {}),
          data: { label: n.element_name, elementType: n.element_type, isContainer },
          style: `width:${n.w}px;height:${n.h}px;`,
          draggable: false,
          selectable: true,
          connectable: false,
        };
      });

      // Build an exact map from OEF node_id → XY Flow instance id.
      // This lets edges resolve the correct instance without any heuristic.
      const nodeIdToIid = {};
      rawNodes.forEach((n, i) => {
        if (n.node_id) nodeIdToIid[n.node_id] = nodeInstanceId[i];
      });

      // Edges: pass raw bounds (absolute coords) and bendpoints so ArchiMateEdge
      // can compute the path without relying on XY Flow's handle positions.
      edges = (data.connections || []).map(c => {
        // Use OEF node_id for exact instance resolution; fall back to element_id for old layouts.
        const srcIid = (c.source_node_id && nodeIdToIid[c.source_node_id]) || c.source_element_id;
        const tgtIid = (c.target_node_id && nodeIdToIid[c.target_node_id]) || c.target_element_id;
        const src = (c.source_node_id && nodeByNodeId[c.source_node_id]) || nodeById[c.source_element_id];
        const tgt = (c.target_node_id && nodeByNodeId[c.target_node_id]) || nodeById[c.target_element_id];
        return {
          id: c.relationship_id,
          source: srcIid,
          target: tgtIid,
          type: 'archimate',
          data: {
            relationshipType: c.relationship_type,
            accessType:       c.access_type || null,
            isDirected:       c.is_directed  || false,
            reversed:         c.reversed     || false,
            label:            c.label        || '',
            modifier:         c.modifier     || '',
            bendpoints: c.bendpoints || [],
            sourceBounds: src ? { x: src.x, y: src.y, w: src.w, h: src.h } : null,
            targetBounds: tgt ? { x: tgt.x, y: tgt.y, w: tgt.w, h: tgt.h } : null,
          },
          selectable: false,
        };
      });
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
</script>

<!--
  Global ArchiMate SVG marker definitions.
  Colors are hardcoded per variant (dark/mid/light) instead of using context-stroke,
  which is not reliably supported in Chrome when markers are defined outside the
  referencing SVG element.
    dark = #374151  (Triggering, Flow, Assignment, Serving, Composition, Aggregation, Specialization)
    mid  = #6B7280  (Realization)
    light= #9CA3AF  (Association, Access, Influence)
-->
<svg width="0" height="0" style="position:absolute;overflow:hidden;pointer-events:none;">
  <defs>
    <!-- Filled arrowhead — dark only (Triggering, Flow, Assignment end) -->
    <marker id="am-filled-arrow-dark" viewBox="0 0 8 6" markerWidth="8" markerHeight="6"
      refX="8" refY="3" orient="auto">
      <polygon points="0,0 8,3 0,6" fill="#374151" />
    </marker>

    <!-- Open V arrowhead — dark (Serving) -->
    <marker id="am-open-arrow-dark" viewBox="0 0 8 8" markerWidth="8" markerHeight="8"
      refX="7" refY="4" orient="auto">
      <path d="M 0,0 L 7,4 L 0,8" fill="none" stroke="#374151" stroke-width="1.4" />
    </marker>

    <!-- Open V arrowhead — light (Association directed, Access Write, Influence) -->
    <marker id="am-open-arrow-light" viewBox="0 0 8 8" markerWidth="8" markerHeight="8"
      refX="7" refY="4" orient="auto">
      <path d="M 0,0 L 7,4 L 0,8" fill="none" stroke="#9CA3AF" stroke-width="1.4" />
    </marker>

    <!-- Open V arrowhead reversed — light only (Access Read/ReadWrite marker-start) -->
    <marker id="am-open-arrow-rev-light" viewBox="0 0 8 8" markerWidth="8" markerHeight="8"
      refX="7" refY="4" orient="auto-start-reverse">
      <path d="M 0,0 L 7,4 L 0,8" fill="none" stroke="#9CA3AF" stroke-width="1.4" />
    </marker>

    <!-- Hollow closed triangle — mid (Realization) -->
    <marker id="am-open-triangle-mid" viewBox="0 0 10 8" markerWidth="10" markerHeight="8"
      refX="10" refY="4" orient="auto">
      <polygon points="0,0 10,4 0,8" fill="#F8FAFC" stroke="#6B7280" stroke-width="1.4" />
    </marker>

    <!-- Hollow closed triangle — dark (Specialization) -->
    <marker id="am-open-triangle-dark" viewBox="0 0 10 8" markerWidth="10" markerHeight="8"
      refX="10" refY="4" orient="auto">
      <polygon points="0,0 10,4 0,8" fill="#F8FAFC" stroke="#374151" stroke-width="1.4" />
    </marker>

    <!-- Filled diamond at source — dark only (Composition) -->
    <marker id="am-filled-diamond-dark" viewBox="-1 -1 14 10" markerWidth="14" markerHeight="10"
      refX="12" refY="4" orient="auto-start-reverse">
      <polygon points="0,4 6,0 12,4 6,8" fill="#374151" />
    </marker>

    <!-- Hollow diamond at source — dark only (Aggregation) -->
    <marker id="am-open-diamond-dark" viewBox="-1 -1 14 10" markerWidth="14" markerHeight="10"
      refX="12" refY="4" orient="auto-start-reverse">
      <polygon points="0,4 6,0 12,4 6,8" fill="#F8FAFC" stroke="#374151" stroke-width="1.4" />
    </marker>

    <!-- Filled circle at source — dark only (Assignment) -->
    <marker id="am-filled-circle-dark" viewBox="-1 -1 10 10" markerWidth="8" markerHeight="8"
      refX="8" refY="4" orient="auto-start-reverse">
      <circle cx="4" cy="4" r="4" fill="#374151" />
    </marker>
  </defs>
</svg>

<div class="{embedded ? 'h-full flex flex-col' : 'content h-full flex flex-col'}">
  {#if !embedded}
    <BackButton onclick={() => push('/ws/' + wsId + '/diagrams')} label="Diagrams" />
  {/if}

  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6 {embedded ? 'px-4' : ''}">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading diagram…
    </div>
  {:else if error}
    <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2 {embedded ? 'm-4' : 'mt-4'}">
      {error}
    </div>
  {:else if data}
    {#if !embedded}
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-[15px] font-semibold">{data.name || 'Diagram'}</h2>
        <span class="text-[11px] text-muted-foreground">{data.nodes?.length ?? 0} elements · {data.connections?.length ?? 0} connections</span>
      </div>
    {/if}

    <div class="flex-1 {embedded ? '' : 'border border-border rounded-lg'} overflow-hidden" style="background:#F8FAFC;">
      <SvelteFlow
        bind:nodes
        bind:edges
        {nodeTypes}
        {edgeTypes}
        fitView
        fitViewOptions={{ padding: 0.1 }}
        nodesDraggable={false}
        nodesConnectable={false}
        panOnDrag={true}
        zoomOnScroll={true}
        colorMode="light"
        connectionMode={ConnectionMode.Loose}
        style="background:#F8FAFC;"
      >
        <Background color="#E5E7EB" gap={20} size={1} />
        <Controls showInteractive={false} />
        <MiniMap
          nodeColor={minimapColor}
          style="background:#F1F5F9;border:1px solid #E2E8F0;border-radius:8px;"
          maskColor="rgba(100,116,139,0.15)"
        />
        <Panel position="top-right">
          <div style="
            background:white;
            border:1px solid #E5E7EB;
            border-radius:8px;
            padding:8px 10px;
            font-size:11px;
            font-family:ui-sans-serif,system-ui,sans-serif;
            color:#374151;
            display:flex;
            flex-direction:column;
            gap:4px;
            box-shadow:0 1px 4px rgba(0,0,0,0.08);
          ">
            <div style="display:flex;align-items:center;gap:6px;"><span style="width:12px;height:12px;border-radius:2px;background:#FFFBEB;border:1.5px solid #D97706;flex-shrink:0;"></span>Business</div>
            <div style="display:flex;align-items:center;gap:6px;"><span style="width:12px;height:12px;border-radius:2px;background:#EFF6FF;border:1.5px solid #2563EB;flex-shrink:0;"></span>Application</div>
            <div style="display:flex;align-items:center;gap:6px;"><span style="width:12px;height:12px;border-radius:2px;background:#F0FDF4;border:1.5px solid #16A34A;flex-shrink:0;"></span>Technology</div>
            <div style="display:flex;align-items:center;gap:6px;"><span style="width:12px;height:12px;border-radius:2px;background:#FAF5FF;border:1.5px solid #7C3AED;flex-shrink:0;"></span>Motivation</div>
            <div style="display:flex;align-items:center;gap:6px;"><span style="width:12px;height:12px;border-radius:2px;background:#FEFCE8;border:1.5px solid #B45309;flex-shrink:0;"></span>Strategy</div>
            <div style="display:flex;align-items:center;gap:6px;"><span style="width:12px;height:12px;border-radius:2px;background:#FFF1F2;border:1.5px solid #BE123C;flex-shrink:0;"></span>Implementation</div>
            <div style="display:flex;align-items:center;gap:6px;"><span style="width:12px;height:12px;border-radius:2px;background:#FDF4FF;border:1.5px solid #A21CAF;flex-shrink:0;"></span>Physical</div>
          </div>
        </Panel>
      </SvelteFlow>
    </div>
  {/if}
</div>

<style>
  :global(.svelte-flow .svelte-flow__edges) {
    z-index: 10;
    pointer-events: none;
  }
  :global(.svelte-flow__edge) {
    pointer-events: all;
  }
</style>
