<script>
  import { push } from 'svelte-spa-router';
  import { SvelteFlow, Controls, Background, MiniMap } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';

  import { api } from '../../lib/api.js';
  import BackButton from '../BackButton.svelte';
  import ArchiMateNode from '../diagram/ArchiMateNode.svelte';

  export let params = {};

  $: wsId = params.wsId;
  $: diagId = params.diagId;

  let data = null;
  let loading = true;
  let error = null;

  let nodes = [];
  let edges = [];

  const nodeTypes = { archimate: ArchiMateNode };

  $: if (diagId) load();

  async function load() {
    loading = true;
    error = null;
    data = null;
    nodes = [];
    edges = [];
    try {
      data = await api.get('/workspaces/' + wsId + '/diagrams/' + diagId + '/render');
      nodes = (data.nodes || []).map(n => ({
        id: n.element_id,
        type: 'archimate',
        position: { x: n.x, y: n.y },
        data: { label: n.element_name, elementType: n.element_type },
        style: `width:${n.w}px;height:${n.h}px;`,
        draggable: false,
        selectable: false,
        connectable: false,
      }));
      edges = (data.connections || []).map(c => ({
        id: c.relationship_id,
        source: c.source_element_id,
        target: c.target_element_id,
        style: 'stroke:#565f89;stroke-width:1.5px;',
        markerEnd: { type: 'arrowClosed', color: '#565f89', width: 14, height: 10 },
        selectable: false,
      }));
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="content h-full flex flex-col">
  <BackButton onclick={() => push('/ws/' + wsId + '/diagrams')} label="Diagrams" />

  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading diagram…
    </div>
  {:else if error}
    <div class="mt-4 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">
      {error}
    </div>
  {:else if data}
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-[15px] font-semibold">{data.name || 'Diagram'}</h2>
      <span class="text-[11px] text-muted-foreground">{data.nodes?.length ?? 0} elements · {data.connections?.length ?? 0} connections</span>
    </div>

    <div class="flex-1 border border-border rounded-lg overflow-hidden bg-[#0d0e14]">
      <SvelteFlow
        bind:nodes
        bind:edges
        {nodeTypes}
        fitView
        fitViewOptions={{ padding: 0.12 }}
        nodesDraggable={false}
        nodesConnectable={false}
        panOnDrag={true}
        zoomOnScroll={true}
        style="background:#0d0e14;"
      >
        <Background color="#1e2030" gap={20} />
        <Controls showInteractive={false} />
        <MiniMap
          nodeColor={n => {
            if (!n.data) return '#565f89';
            const t = n.data.elementType || '';
            if (t.startsWith('Application')) return '#7aa2f7';
            if (t.startsWith('Business') || t === 'Capability') return '#e0af68';
            if (t.startsWith('Technology') || t === 'Node' || t === 'SystemSoftware') return '#9ece6a';
            return '#565f89';
          }}
          style="background:#1a1b26;"
          maskColor="rgba(0,0,0,0.3)"
        />
      </SvelteFlow>
    </div>
  {/if}
</div>
