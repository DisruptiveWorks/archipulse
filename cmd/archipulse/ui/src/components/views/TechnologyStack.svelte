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
  let hoverApp    = $state(null);
  let showSaveDialog = $state(false);
  let showInfo    = $state(false);

  // ── Filter state ──────────────────────────────────────────────────────────
  let search      = $state('');
  let lcFilterArr = $state([]);
  let hiddenCats  = $state(new Set());

  const hasFilters  = $derived(search || lcFilterArr.length > 0 || hiddenCats.size > 0);
  const saveFilters = $derived({ search, lifecycle: lcFilterArr });

  // ── Lifecycle colors ──────────────────────────────────────────────────────
  const LC_COLORS = {
    'Production': '#16a34a',
    'Phase In':   '#0ea5e9',
    'Phase Out':  '#f59e0b',
    'Retired':    '#94a3b8',
  };
  function lcColor(lc) { return LC_COLORS[lc] ?? '#94a3b8'; }

  // ── Category colors ───────────────────────────────────────────────────────
  const CAT_COLORS = {
    'Business Application':      '#3b82f6',
    'Data and Analytics':        '#8b5cf6',
    'Integration and Middleware':'#f97316',
    'Infrastructure':            '#64748b',
    'Security and Identity':     '#ef4444',
    'OT / Industrial':           '#eab308',
  };
  function catColor(cat) { return CAT_COLORS[cat] ?? '#94a3b8'; }

  // ── Derived data ──────────────────────────────────────────────────────────
  const allCategories = $derived(
    data ? [...new Map(data.tech.map(t => [t.category, t])).keys()].filter(Boolean).sort() : []
  );

  const visibleApps = $derived(
    data ? data.apps.filter(a =>
      !search || a.name.toLowerCase().includes(search.toLowerCase())
    ) : []
  );

  const visibleTech = $derived(
    data ? data.tech.filter(t => !hiddenCats.has(t.category)) : []
  );

  const appTechSet = $derived(
    data ? new Set(data.app_tech.map(([a, t]) => a + '|' + t)) : new Set()
  );

  function hasLink(appId, techId) { return appTechSet.has(appId + '|' + techId); }

  function toggleCat(cat) {
    const s = new Set(hiddenCats);
    if (s.has(cat)) s.delete(cat); else s.add(cat);
    hiddenCats = s;
  }

  function clearFilters() {
    search      = '';
    lcFilterArr = [];
    hiddenCats  = new Set();
  }

  // ── Load ──────────────────────────────────────────────────────────────────
  onMount(async () => {
    if (initialFilters?.search)    search      = initialFilters.search;
    if (initialFilters?.lifecycle) lcFilterArr = initialFilters.lifecycle;

    loading = true;
    try {
      data = await api.get('/workspaces/' + wsId + '/views/technology-stack/matrix');
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
        <h1 class="text-[18px] font-semibold">{savedViewName ?? 'Technology Stack'}</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">Applications and the technology components they run on</div>
      </div>
      <div class="flex items-center gap-2 flex-wrap">
        {#if !savedViewName}
          <button onclick={() => showSaveDialog = true}
            class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-border text-[12px] text-muted-foreground hover:text-foreground hover:border-primary transition-colors">
            ⊕ Save view
          </button>
        {/if}
        <ViewInfoDialog title="Technology Stack — setup guide" bind:open={showInfo}>
          <p>This view shows which <strong>technology elements</strong> (infrastructure, databases, platforms) each application runs on.</p>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Required model elements</div>
            <div class="bg-muted rounded-md px-3 py-2 font-mono text-[12px]">Node · SystemSoftware · TechnologyService · Device · Artifact</div>
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Required relationships</div>
            <p class="text-[12px] text-muted-foreground">An <strong>Assignment</strong> or <strong>Realization</strong> relationship between an application element and a technology element.</p>
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Optional properties on technology elements</div>
            <table class="w-full text-[12px] border border-border rounded-md overflow-hidden">
              <thead>
                <tr class="bg-muted/60">
                  <th class="text-left px-3 py-1.5 font-semibold">Property key</th>
                  <th class="text-left px-3 py-1.5 font-semibold">Example value</th>
                </tr>
              </thead>
              <tbody>
                {#each [['category','Infrastructure'],['vendor','Oracle'],['version','19c'],['lifecycle','Production']] as [k,v]}
                  <tr class="border-t border-border">
                    <td class="px-3 py-1.5 font-mono">{k}</td>
                    <td class="px-3 py-1.5 text-muted-foreground">{v}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
            <p class="text-[11.5px] text-muted-foreground mt-1.5">Cell left-border color encodes the technology's lifecycle status.</p>
          </div>
        </ViewInfoDialog>
      </div>
    </div>

    <SaveViewDialog bind:open={showSaveDialog} {wsId} viewType="technology-stack" filters={saveFilters} />
    <SaveViewUpdateBar {wsId} {savedViewId} {savedViewName} currentFilters={saveFilters} {initialFilters} />

    {#if !data.apps?.length && !data.tech?.length}
      <div class="text-center py-20 text-muted-foreground">
        <div class="text-[40px] mb-3">📭</div>
        <p class="text-[14px]">No technology relationships found.</p>
        <p class="text-[13px] mt-1 opacity-70">Add <strong>Assignment</strong> or <strong>Realization</strong> relationships from ApplicationComponent elements to technology elements.</p>
      </div>
    {:else}
      <div class="flex gap-4 items-start">

        <!-- Filter panel -->
        <aside class="hidden sm:flex flex-col gap-3 w-48 flex-shrink-0">
          <input type="search" bind:value={search} placeholder="Find application…"
            class="w-full bg-card border border-border rounded-md px-2.5 py-1.5 text-[12px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary" />

          {#if allCategories.length > 0}
            <div>
              <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-1.5">Tech categories</div>
              {#each allCategories as cat}
                {@const hidden = hiddenCats.has(cat)}
                {@const count = data.tech.filter(t => t.category === cat).length}
                <label class="flex items-center gap-2 px-1 py-1 text-[12px] cursor-pointer select-none group">
                  <span class="flex-shrink-0 size-3.5 rounded border flex items-center justify-center transition-colors"
                    style="background: {hidden ? 'transparent' : catColor(cat)}; border-color: {catColor(cat)}">
                    {#if !hidden}
                      <svg class="size-2.5 text-white" viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.8">
                        <path d="M1.5 5l2.5 2.5 4.5-4.5"/>
                      </svg>
                    {/if}
                  </span>
                  <input type="checkbox" class="sr-only" checked={!hidden} onchange={() => toggleCat(cat)} />
                  <span class="truncate flex-1 {hidden ? 'text-muted-foreground line-through' : 'text-foreground'}">{cat}</span>
                  <span class="text-[11px] text-muted-foreground">{count}</span>
                </label>
              {/each}
            </div>
          {/if}

          <div class="text-[11px] text-muted-foreground px-1">{visibleApps.length} apps · {visibleTech.length} tech</div>

          {#if hasFilters}
            <button onclick={clearFilters} class="text-[12px] text-muted-foreground hover:text-foreground underline text-left">
              Clear all filters
            </button>
          {/if}
        </aside>

        <!-- Single matrix table -->
        <div class="flex-1 min-w-0 overflow-auto">
          {#if visibleTech.length === 0}
            <div class="text-center py-16 text-muted-foreground text-[13px]">No technology columns visible — enable categories above.</div>
          {:else}
            <div class="border border-border rounded-lg overflow-x-auto">
              <table class="text-[12px] border-collapse" style="table-layout: fixed; width: max-content; min-width: 100%">
                <colgroup>
                  <col style="width:190px" />
                  {#each visibleTech as _}<col style="width:100px" />{/each}
                </colgroup>
                <thead>
                  <tr>
                    <th class="sticky left-0 z-10 bg-muted border-r border-b border-border px-2.5 py-2 text-left font-normal">
                      <span class="text-[10px] font-semibold text-muted-foreground uppercase tracking-wide">App ↓ · Tech →</span>
                    </th>
                    {#each visibleTech as tech}
                      <th class="bg-muted/60 border-r border-b border-border px-1.5 py-2 font-normal text-center align-top" style="border-top: 3px solid {catColor(tech.category)}">
                        <div class="font-semibold text-[11px] text-foreground leading-tight break-words hyphens-auto">{tech.name}</div>
                        {#if tech.vendor || tech.version}
                          <div class="text-[10px] text-muted-foreground mt-0.5">{[tech.vendor, tech.version].filter(Boolean).join(' ')}</div>
                        {/if}
                      </th>
                    {/each}
                  </tr>
                </thead>
                <tbody>
                  {#each visibleApps as app}
                    <tr>
                      <td class="sticky left-0 z-10 border-r border-b border-border px-2.5 py-2 {selectedApp?.id === app.id ? 'bg-[#cff0ff]' : hoverApp === app.id ? 'bg-muted/20' : 'bg-card'}">
                        <button
                          class="flex items-center gap-2 text-left w-full cursor-pointer"
                          onmouseenter={() => hoverApp = app.id}
                          onmouseleave={() => hoverApp = null}
                          onclick={() => selectedApp = selectedApp?.id === app.id ? null : { id: app.id, name: app.name, type: app.type, properties: {} }}>
                          <span class="size-1.5 rounded-full flex-shrink-0 bg-[#0ea5e9]"></span>
                          <span class="truncate text-[11.5px] font-medium text-foreground">{app.name}</span>
                        </button>
                      </td>
                      {#each visibleTech as tech}
                        {@const on = hasLink(app.id, tech.id)}
                        <td
                          class="border-r border-b border-border text-center transition-colors {on ? 'bg-[#c9f0c7]/50' : ''} {(!on && hoverApp === app.id) ? 'opacity-25' : ''}"
                          style={on ? `border-left: 3px solid ${lcColor(tech.lifecycle)}` : ''}>
                          {#if on}
                            <span class="text-green-700 text-[14px]">●</span>
                          {/if}
                        </td>
                      {/each}
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>

            <!-- Legend -->
            <div class="flex flex-wrap gap-x-5 gap-y-1.5 px-1 mt-3">
              <div class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wide self-center">Tech lifecycle:</div>
              {#each Object.entries(LC_COLORS) as [lc, color]}
                <div class="flex items-center gap-1.5 text-[12px]">
                  <span class="w-3 h-3 rounded flex-shrink-0" style="background:{color}"></span>
                  <span class="text-muted-foreground">{lc}</span>
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
