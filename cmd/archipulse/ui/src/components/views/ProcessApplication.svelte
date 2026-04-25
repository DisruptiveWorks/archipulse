<script>
  import { onMount } from 'svelte';
  import { api } from '../../lib/api.js';
  import AppDetailPanel from './AppDetailPanel.svelte';
  import SaveViewDialog from './SaveViewDialog.svelte';
  import SaveViewUpdateBar from './SaveViewUpdateBar.svelte';
  import ViewInfoDialog from './ViewInfoDialog.svelte';

  const { params = {}, initialFilters = null, savedViewName = null, savedViewId = null } = $props();

  const wsId = $derived(params.wsId);

  let data        = $state(null);
  let loading     = $state(true);
  let error       = $state(null);
  let selectedApp = $state(null);
  let showSaveDialog = $state(false);
  let showInfo    = $state(false);

  // ── Filter state ──────────────────────────────────────────────────────────
  let search       = $state('');
  let capFilterArr = $state([]);
  let kindFilterArr = $state(['R', 'W', 'E', 'S']);

  const capFilter  = $derived(new Set(capFilterArr));
  const kindFilter = $derived(new Set(kindFilterArr));
  const hasFilters = $derived(search || capFilterArr.length > 0 || kindFilterArr.length < 4);
  const saveFilters = $derived({ search, capabilities: capFilterArr, kinds: kindFilterArr });

  // ── Filtering ─────────────────────────────────────────────────────────────
  const visibleProcesses = $derived(
    data ? data.processes.filter(p => {
      if (search && !p.name.toLowerCase().includes(search.toLowerCase())) return false;
      if (capFilterArr.length && !capFilterArr.includes(p.capability)) return false;
      return true;
    }) : []
  );

  const visibleProcIds = $derived(new Set(visibleProcesses.map(p => p.id)));

  const visibleLinks = $derived(
    data ? data.links.filter(l =>
      visibleProcIds.has(l.process_id) && kindFilter.has(l.kind)
    ) : []
  );

  const visibleAppIds = $derived(new Set(visibleLinks.map(l => l.app_id)));

  const visibleApps = $derived(
    data ? data.apps.filter(a => visibleAppIds.has(a.id)) : []
  );

  function clearFilters() {
    search = '';
    capFilterArr = [];
    kindFilterArr = ['R', 'W', 'E', 'S'];
  }

  // ── Kind display ──────────────────────────────────────────────────────────
  function markFor(k) {
    return { R: 'R', W: 'W', E: '▶', S: '→' }[k] ?? k;
  }
  function markColor(k) {
    return { R: '#0891b2', W: '#dc2626', E: '#f59e0b', S: '#7c3aed' }[k] ?? '#94a3b8';
  }
  function markLabel(k) {
    return { R: 'Reads data', W: 'Writes data', E: 'Executes / triggers', S: 'Serving' }[k] ?? k;
  }

  // ── Cell lookup ───────────────────────────────────────────────────────────
  const cellMap = $derived((() => {
    const m = new Map();
    for (const l of visibleLinks) {
      m.set(l.process_id + '|' + l.app_id, l.kind);
    }
    return m;
  })());

  // ── Stats ─────────────────────────────────────────────────────────────────
  const statsText = $derived(
    `${visibleProcesses.length} processes × ${visibleApps.length} apps · ${visibleLinks.length} usages`
  );

  // ── Load ──────────────────────────────────────────────────────────────────
  onMount(async () => {
    if (initialFilters?.search)       search       = initialFilters.search;
    if (initialFilters?.capabilities) capFilterArr = initialFilters.capabilities;
    if (initialFilters?.kinds)        kindFilterArr = initialFilters.kinds;

    loading = true;
    try {
      data = await api.get('/workspaces/' + wsId + '/views/process-application/matrix');
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });
</script>

<div class="content">
  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">{error}</div>
  {:else if data}

    <!-- Header -->
    <div class="flex items-center justify-between gap-4 mb-5 flex-wrap">
      <div>
        <h1 class="text-[18px] font-semibold">{savedViewName ?? 'Process–Application Usage'}</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">Which applications support each business process</div>
      </div>
      <div class="flex items-center gap-2 flex-wrap">
        {#if !savedViewName}
          <button onclick={() => showSaveDialog = true}
            class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-border text-[12px] text-muted-foreground hover:text-foreground hover:border-primary transition-colors">
            ⊕ Save view
          </button>
        {/if}
        <ViewInfoDialog title="Process–Application Usage — setup guide" bind:open={showInfo}>
          <p>This view shows a matrix of <strong>business processes</strong> (rows) and the <strong>applications</strong> that support them (columns). Cells show how the application is used by each process.</p>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Required model elements</div>
            <div class="bg-muted rounded-md px-3 py-2 font-mono text-[12px]">BusinessProcess · BusinessFunction · BusinessInteraction · BusinessService</div>
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Required relationships</div>
            <p class="text-[12px] text-muted-foreground">A relationship from a business process to an application element of type: <strong>Serving, Realization, Assignment, Triggering,</strong> or <strong>Association</strong>.</p>
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Optional: usage kind (R / W / E)</div>
            <p class="text-[12px] text-muted-foreground">Add a property <code class="bg-muted rounded px-1">usage_kind</code> on the relationship itself to mark it as <strong>R</strong> (Read), <strong>W</strong> (Write), or <strong>E</strong> (Execute/Trigger). Without it, the cell shows <strong>→</strong> (Serving).</p>
            <pre class="bg-muted rounded-md px-3 py-2 text-[11px] overflow-x-auto mt-1.5">{`<relationship id="r-1" type="Serving" source="proc-id" target="app-id">
  <properties>
    <property key="usage_kind" value="W" />
  </properties>
</relationship>`}</pre>
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Optional: capability grouping</div>
            <p class="text-[12px] text-muted-foreground">Add a <code class="bg-muted rounded px-1">capability</code> property on the process element to enable the capability filter in the sidebar.</p>
          </div>
        </ViewInfoDialog>
      </div>
    </div>

    <SaveViewDialog bind:open={showSaveDialog} {wsId} viewType="process-application" filters={saveFilters} />
    <SaveViewUpdateBar {wsId} {savedViewId} {savedViewName} currentFilters={saveFilters} {initialFilters} />

    {#if !data.processes?.length}
      <div class="text-center py-20 text-muted-foreground">
        <div class="text-[40px] mb-3">📭</div>
        <p class="text-[14px]">No process–application relationships found.</p>
        <p class="text-[13px] mt-1 opacity-70">Add <strong>Serving</strong> or <strong>Realization</strong> relationships from BusinessProcess elements to ApplicationComponent elements.</p>
      </div>
    {:else}
      <div class="flex gap-4 items-start">

        <!-- Filter panel -->
        <aside class="hidden sm:flex flex-col gap-3 w-44 flex-shrink-0">
          <input type="search" bind:value={search} placeholder="Find process…"
            class="w-full bg-card border border-border rounded-md px-2.5 py-1.5 text-[12px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary" />

          {#if data.processes.some(p => p.capability)}
            <div>
              <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-1.5">Business Capability</div>
              {#each [...new Set(data.processes.filter(p => p.capability).map(p => p.capability))].sort() as cap}
                <label class="flex items-center gap-2 px-1 py-1 cursor-pointer rounded hover:bg-muted/50 text-[12px]">
                  <input type="checkbox" bind:group={capFilterArr} value={cap} />
                  <span class="size-2 rounded-sm flex-shrink-0 bg-[#fffbcc] border border-yellow-300"></span>
                  <span class="truncate text-foreground">{cap}</span>
                </label>
              {/each}
            </div>
          {/if}

          <div>
            <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-1.5">Usage type</div>
            {#each ['R','W','E','S'] as k}
              <label class="flex items-center gap-2 px-1 py-1 cursor-pointer rounded hover:bg-muted/50 text-[12px]">
                <input type="checkbox" bind:group={kindFilterArr} value={k} />
                <span class="size-[14px] rounded flex-shrink-0 grid place-items-center text-[10px] font-bold text-white" style="background:{markColor(k)}">{markFor(k)}</span>
                <span class="text-foreground">{markLabel(k)}</span>
              </label>
            {/each}
          </div>

          <div class="text-[11px] text-muted-foreground px-1">{statsText}</div>

          {#if hasFilters}
            <button onclick={clearFilters} class="text-[12px] text-muted-foreground hover:text-foreground underline text-left">
              Clear all filters
            </button>
          {/if}
        </aside>

        <!-- Matrix -->
        <div class="flex-1 min-w-0 overflow-auto">
          {#if visibleProcesses.length === 0}
            <div class="text-center py-16 text-muted-foreground text-[13px]">No matches for the current filters.</div>
          {:else}
            <table class="text-[12px] border-collapse w-max min-w-full">
              <thead>
                <tr>
                  <!-- Corner cell -->
                  <th class="sticky left-0 z-20 bg-card border border-border px-3 py-2 text-left text-[11px] font-semibold text-muted-foreground whitespace-nowrap min-w-[180px]">
                    Process ↓ · Application →
                  </th>
                  {#each visibleApps as app}
                    <th
                      class="border border-border px-2 py-1 text-left cursor-pointer select-none transition-colors hover:bg-[#cff0ff]/60 whitespace-nowrap"
                      class:bg-[#cff0ff]={selectedApp?.id === app.id}
                      onclick={() => selectedApp = selectedApp?.id === app.id ? null : { id: app.id, name: app.name, type: app.type, properties: {} }}>
                      <div class="font-medium text-[11px] text-foreground">{app.name}</div>
                      <div class="text-[10px] text-muted-foreground opacity-70">{app.type.replace('Application', '')}</div>
                    </th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each visibleProcesses as proc}
                  <tr class="hover:bg-muted/30 transition-colors">
                    <th class="sticky left-0 z-10 bg-card border border-border px-3 py-2 text-left font-normal">
                      <div class="font-medium text-foreground text-[12px]">{proc.name}</div>
                      {#if proc.capability}
                        <div class="text-[10.5px] text-muted-foreground opacity-70">{proc.capability}</div>
                      {/if}
                    </th>
                    {#each visibleApps as app}
                      {@const kind = cellMap.get(proc.id + '|' + app.id)}
                      <td class="border border-border text-center {selectedApp?.id === app.id ? 'bg-[#cff0ff]/30' : ''}" style="width:60px;min-width:60px;max-width:80px">
                        {#if kind}
                          <span class="font-bold text-[12px]" style="color:{markColor(kind)}">{markFor(kind)}</span>
                        {/if}
                      </td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>

            <!-- Legend -->
            <div class="mt-4 flex flex-wrap gap-x-5 gap-y-1.5 px-1">
              {#each ['R','W','E','S'] as k}
                <div class="flex items-center gap-1.5 text-[12px]">
                  <span class="font-bold text-[12px]" style="color:{markColor(k)}">{markFor(k)}</span>
                  <span class="text-muted-foreground">{markLabel(k)}</span>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}
  {/if}
</div>

<AppDetailPanel app={selectedApp} {wsId} on:close={() => selectedApp = null} />
