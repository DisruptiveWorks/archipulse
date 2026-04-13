<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import DiagramView from '../components/views/DiagramView.svelte';
  import FolderTree from '../components/diagram/FolderTree.svelte';

  export let params = {};
  $: wsId = params.wsId;

  let tree = null;
  let loading = true;
  let error = null;

  let selectedDiagram = null;
  let collapsed = {};

  // Panel open state — starts closed on mobile, open on desktop.
  let panelOpen = false;
  onMount(() => {
    panelOpen = window.innerWidth >= 768;
  });

  function togglePanel() {
    panelOpen = !panelOpen;
  }

  // Auto-close panel on mobile after selecting a diagram.
  function selectDiagram(d) {
    selectedDiagram = d;
    if (window.innerWidth < 768) panelOpen = false;
  }

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

  <!-- ── Left panel: folder tree ── -->
  <!-- On mobile: absolute overlay drawer. On desktop: fixed-width sidebar. -->
  <div
    class="
      flex-shrink-0 border-r border-border flex flex-col overflow-hidden bg-card
      transition-all duration-200 ease-in-out
      {panelOpen ? 'w-72' : 'w-0'}
      md:relative absolute z-20 h-full
    "
    style="min-width: 0;"
  >
    <!-- Panel content — hidden when width collapses -->
    <div class="w-72 h-full flex flex-col overflow-hidden">
      <div class="px-4 py-3 border-b border-border flex-shrink-0 flex items-center justify-between">
        <div>
          <h2 class="text-[13px] font-semibold text-foreground">Diagrams</h2>
          {#if !loading && !error && tree}
            <p class="text-[11px] text-muted-foreground mt-0.5">{total} views</p>
          {/if}
        </div>
        <!-- Close button (visible on mobile inside panel) -->
        <button
          class="md:hidden p-1 rounded hover:bg-muted text-muted-foreground"
          onclick={togglePanel}
          aria-label="Close panel"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8">
            <line x1="3" y1="3" x2="13" y2="13"/>
            <line x1="13" y1="3" x2="3" y2="13"/>
          </svg>
        </button>
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
  </div>

  <!-- Mobile backdrop — tap to close panel -->
  {#if panelOpen}
    <button
      class="md:hidden absolute inset-0 z-10 bg-black/30"
      onclick={togglePanel}
      aria-label="Close panel"
      tabindex="-1"
    ></button>
  {/if}

  <!-- ── Right: diagram viewer ── -->
  <div class="flex-1 flex flex-col overflow-hidden min-w-0">

    <!-- Toggle button bar — always visible on mobile, shown on desktop when panel is closed -->
    <div class="flex items-center gap-2 px-3 py-2 border-b border-border flex-shrink-0 bg-card/80 backdrop-blur-sm">
      <button
        class="flex items-center gap-1.5 px-2 py-1 rounded text-[12px] text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
        onclick={togglePanel}
        aria-label="{panelOpen ? 'Hide' : 'Show'} diagram list"
      >
        <!-- Sidebar icon -->
        <svg width="15" height="15" viewBox="0 0 15 15" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="1" y="1" width="13" height="13" rx="1.5"/>
          <line x1="5" y1="1" x2="5" y2="14"/>
        </svg>
        <span class="hidden sm:inline">{panelOpen ? 'Hide' : 'Diagrams'}</span>
      </button>

      {#if selectedDiagram}
        <span class="text-[12px] text-muted-foreground truncate">{selectedDiagram.name}</span>
      {/if}
    </div>

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
