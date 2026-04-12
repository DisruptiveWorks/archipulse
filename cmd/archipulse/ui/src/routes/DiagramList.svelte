<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import DiagramView from '../components/views/DiagramView.svelte';
  import FolderTree from '../components/diagram/FolderTree.svelte';

  export let params = {};
  $: wsId = params.wsId;

  let tree = null;   // { folders: [], diagrams: [] }
  let loading = true;
  let error = null;

  let selectedDiagram = null; // { id, name }
  let collapsed = {};         // folder.id → bool

  $: if (wsId) loadTree();

  async function loadTree() {
    loading = true;
    error = null;
    tree = null;
    selectedDiagram = null;
    collapsed = {};
    try {
      tree = await api.get('/workspaces/' + wsId + '/diagram-tree');
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function selectDiagram(d) {
    selectedDiagram = d;
  }

  function toggleFolder(id) {
    collapsed[id] = !collapsed[id];
    collapsed = { ...collapsed };
  }

  function totalDiagrams(node) {
    let count = (node.diagrams || []).length;
    for (const child of (node.children || [])) count += totalDiagrams(child);
    return count;
  }

  $: total = tree ? countAll(tree) : 0;
  function countAll(t) {
    let n = (t.diagrams || []).length;
    for (const f of (t.folders || [])) n += totalDiagrams(f);
    return n;
  }
</script>

<div class="content-fill">

  <!-- ── Left: folder tree ── -->
  <div class="flex-shrink-0 w-72 border-r border-border flex flex-col overflow-hidden bg-card">
    <div class="px-4 py-3 border-b border-border flex-shrink-0">
      <h2 class="text-[13px] font-semibold text-foreground">Diagrams</h2>
      {#if !loading && !error && tree}
        <p class="text-[11px] text-muted-foreground mt-0.5">{total} views</p>
      {/if}
    </div>

    <div class="flex-1 overflow-y-auto py-1">
      {#if loading}
        <div class="flex items-center gap-2 text-muted-foreground text-[12px] px-4 py-4">
          <div class="size-3 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
          Loading…
        </div>
      {:else if error}
        <div class="text-[12px] text-destructive px-4 py-3">{error}</div>
      {:else if tree}
        <!-- Root-level loose diagrams (no folder) -->
        {#each (tree.diagrams || []) as d}
          <button
            class="w-full text-left flex items-center gap-2 px-4 py-1.5 text-[12px] hover:bg-muted/50 transition-colors
              {selectedDiagram?.id === d.id ? 'bg-primary/10 text-primary font-medium' : 'text-foreground'}"
            onclick={() => selectDiagram(d)}
          >
            <span class="text-[10px] opacity-40 flex-shrink-0">▪</span>
            <span class="truncate">{d.name || '(unnamed)'}</span>
          </button>
        {/each}

        <!-- Root-level folders -->
        {#each (tree.folders || []) as folder}
          <FolderTree
            {folder}
            {collapsed}
            {selectedDiagram}
            depth={0}
            on:toggle={e => toggleFolder(e.detail)}
            on:select={e => selectDiagram(e.detail)}
          />
        {/each}

        {#if total === 0}
          <div class="text-center py-8 text-muted-foreground text-[12px] px-4">
            No diagrams found.<br>Import a model to see them here.
          </div>
        {/if}
      {/if}
    </div>
  </div>

  <!-- ── Right: diagram viewer ── -->
  <div class="flex-1 flex flex-col overflow-hidden">
    {#if selectedDiagram}
      <DiagramView params={{ wsId, diagId: selectedDiagram.id }} embedded={true} />
    {:else}
      <div class="flex-1 flex flex-col items-center justify-center text-muted-foreground gap-3">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" opacity="0.3">
          <rect x="3" y="3" width="18" height="18" rx="2"/>
          <path d="M3 9h18M9 21V9"/>
        </svg>
        <p class="text-[13px]">Select a diagram to view it</p>
      </div>
    {/if}
  </div>
</div>
