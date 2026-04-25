<script>
  import { onMount } from 'svelte';
  import { api } from '../../lib/api.js';
  import SaveViewDialog from './SaveViewDialog.svelte';
  import SaveViewUpdateBar from './SaveViewUpdateBar.svelte';
  import AppDetailPanel from './AppDetailPanel.svelte';
  import ViewInfoDialog from './ViewInfoDialog.svelte';

  const { params = {}, initialFilters = null, savedViewName = null, savedViewId = null } = $props();

  const wsId = $derived(params.wsId);

  let data        = $state(null);
  let loading     = $state(true);
  let error       = $state(null);
  let showSaveDialog = $state(false);
  let selectedApp = $state(null);
  let showInfo    = $state(false);

  // ── Filter state ──────────────────────────────────────────────────────────
  let colorBy       = $state('criticality');
  let search        = $state('');
  let domainFilterArr = $state([]);
  let lcFilterArr     = $state([]);

  const domainFilter    = $derived(new Set(domainFilterArr));
  const lifecycleFilter = $derived(new Set(lcFilterArr));
  const hasFilters      = $derived(search || domainFilterArr.length > 0 || lcFilterArr.length > 0);
  const saveFilters     = $derived({ colorBy, domains: domainFilterArr, lifecycle: lcFilterArr });

  // ── Domain palette ────────────────────────────────────────────────────────
  const DOMAIN_COLORS = [
    '#2563eb','#d97706','#16a34a','#7c3aed','#dc2626',
    '#0891b2','#db2777','#ca8a04','#9333ea','#059669',
  ];
  function domainColor(idx) { return DOMAIN_COLORS[idx % DOMAIN_COLORS.length]; }

  // ── Color-by logic ────────────────────────────────────────────────────────
  const CRIT_COLORS = {
    'Critical': '#dc2626', 'High': '#f59e0b', 'Medium': '#0ea5e9', 'Low': '#16a34a',
  };
  const LC_COLORS = {
    'Production': '#16a34a', 'Pilot': '#8b5cf6', 'Planned': '#0ea5e9',
    'Retiring':   '#f59e0b', 'Decommissioned': '#94a3b8',
  };
  const DEPLOY_COLORS = {
    'On-Premise': '#64748b', 'Public Cloud': '#6366f1', 'SaaS': '#10b981',
    'Hybrid':     '#a855f7', 'Private Cloud': '#6366f1', 'AWS': '#f59e0b', 'Azure': '#0ea5e9',
  };

  const colorBarFor = $derived((app) => {
    const p = app.properties ?? {};
    if (colorBy === 'criticality')      return CRIT_COLORS[p.criticality]        ?? '#94a3b8';
    if (colorBy === 'lifecycle_status') return LC_COLORS[p.lifecycle_status]      ?? '#94a3b8';
    if (colorBy === 'deployment_model') return DEPLOY_COLORS[p.deployment_model]  ?? '#94a3b8';
    return '#94a3b8';
  });

  const colorLabelFor = $derived((app) => {
    const p = app.properties ?? {};
    if (colorBy === 'criticality')      return p.criticality        ?? '';
    if (colorBy === 'lifecycle_status') return p.lifecycle_status   ?? '';
    if (colorBy === 'deployment_model') return p.deployment_model   ?? '';
    return '';
  });

  const legendEntries = $derived((() => {
    if (colorBy === 'criticality')      return Object.entries(CRIT_COLORS).map(([v,c]) => ({ v, c }));
    if (colorBy === 'lifecycle_status') return Object.entries(LC_COLORS).map(([v,c]) => ({ v, c }));
    if (colorBy === 'deployment_model') return Object.entries(DEPLOY_COLORS).map(([v,c]) => ({ v, c }));
    return [];
  })());

  // ── Filtering ─────────────────────────────────────────────────────────────
  const matchesApp = $derived((app) => {
    if (search && !app.name.toLowerCase().includes(search.toLowerCase())) return false;
    if (lcFilterArr.length && !lcFilterArr.includes(app.properties?.lifecycle_status)) return false;
    return true;
  });

  function clearFilters() {
    search = '';
    domainFilterArr = [];
    lcFilterArr = [];
  }

  // ── Load ──────────────────────────────────────────────────────────────────
  onMount(async () => {
    colorBy         = initialFilters?.colorBy   ?? 'criticality';
    domainFilterArr = initialFilters?.domains   ?? [];
    lcFilterArr     = initialFilters?.lifecycle ?? [];
    loading = true;
    try {
      data = await api.get('/workspaces/' + wsId + '/views/application-landscape/map');
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
        <h1 class="text-[18px] font-semibold">{savedViewName ?? 'Application Landscape'}</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">Applications grouped by business domain</div>
      </div>
      <div class="flex items-center gap-2 flex-wrap">
        <span class="text-[12px] text-muted-foreground">Color by</span>
        <select bind:value={colorBy}
          class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary">
          <option value="criticality">Business Criticality</option>
          <option value="lifecycle_status">Lifecycle Status</option>
          <option value="deployment_model">Deployment Model</option>
        </select>
        {#if !savedViewName}
          <button onclick={() => showSaveDialog = true}
            class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-border text-[12px] text-muted-foreground hover:text-foreground hover:border-primary transition-colors">
            ⊕ Save view
          </button>
        {/if}
        <ViewInfoDialog title="Application Landscape — setup guide" bind:open={showInfo}>
          <p>This view groups <strong>ApplicationComponent</strong> and <strong>ApplicationService</strong> elements by their <strong>business domain</strong>, letting you see which apps belong to each area of the business.</p>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Required model element</div>
            <div class="bg-muted rounded-md px-3 py-2 font-mono text-[12px]">ApplicationComponent <span class="text-muted-foreground">or</span> ApplicationService</div>
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Required property</div>
            <table class="w-full text-[12px] border border-border rounded-md overflow-hidden">
              <thead>
                <tr class="bg-muted/60">
                  <th class="text-left px-3 py-1.5 font-semibold">Property key</th>
                  <th class="text-left px-3 py-1.5 font-semibold">Example value</th>
                  <th class="text-left px-3 py-1.5 font-semibold">Notes</th>
                </tr>
              </thead>
              <tbody>
                <tr class="border-t border-border">
                  <td class="px-3 py-1.5 font-mono">domain</td>
                  <td class="px-3 py-1.5 text-muted-foreground">Sales &amp; Distribution</td>
                  <td class="px-3 py-1.5 text-muted-foreground">Primary key</td>
                </tr>
                <tr class="border-t border-border">
                  <td class="px-3 py-1.5 font-mono">business_domain</td>
                  <td class="px-3 py-1.5 text-muted-foreground">Finance</td>
                  <td class="px-3 py-1.5 text-muted-foreground">Fallback if <span class="font-mono">domain</span> is absent</td>
                </tr>
              </tbody>
            </table>
            <p class="text-[11.5px] text-muted-foreground mt-1.5">Apps without either property appear in an <em>Uncategorized</em> group.</p>
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Optional properties (enrich the detail panel)</div>
            <table class="w-full text-[12px] border border-border rounded-md overflow-hidden">
              <thead>
                <tr class="bg-muted/60">
                  <th class="text-left px-3 py-1.5 font-semibold">Property key</th>
                  <th class="text-left px-3 py-1.5 font-semibold">Expected values</th>
                </tr>
              </thead>
              <tbody>
                {#each [
                  ['criticality',       'Critical · High · Medium · Low'],
                  ['lifecycle_status',  'Production · Pilot · Planned · Retiring · Decommissioned'],
                  ['deployment_model',  'On-Premise · Public Cloud · SaaS · AWS · Azure · Hybrid'],
                  ['vendor',            'Free text'],
                  ['business_owner',    'Free text'],
                  ['user_count',        'Numeric'],
                ] as [k, v]}
                  <tr class="border-t border-border">
                    <td class="px-3 py-1.5 font-mono">{k}</td>
                    <td class="px-3 py-1.5 text-muted-foreground">{v}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">How to import</div>
            <p>Set these as <strong>properties</strong> on your elements in your AOEF XML model, then import it via the sidebar. Example:</p>
            <pre class="bg-muted rounded-md px-3 py-2 text-[11px] overflow-x-auto">{`<element id="id-1" type="ApplicationComponent">
  <name>CRM System</name>
  <properties>
    <property key="domain" value="Sales &amp; Distribution" />
    <property key="criticality" value="High" />
    <property key="lifecycle_status" value="Production" />
  </properties>
</element>`}</pre>
          </div>
        </ViewInfoDialog>
      </div>
    </div>

    <SaveViewDialog bind:open={showSaveDialog} {wsId} viewType="application-landscape" filters={saveFilters} />
    <SaveViewUpdateBar {wsId} {savedViewId} {savedViewName} currentFilters={saveFilters} {initialFilters} />

    {#if !data.domains?.length}
      <div class="text-center py-20 text-muted-foreground">
        <div class="text-[40px] mb-3">📭</div>
        <p class="text-[14px]">No domains found.</p>
        <p class="text-[13px] mt-1 opacity-70">Add a <strong>domain</strong> property to your ApplicationComponent elements to group them here.</p>
      </div>
    {:else}
      <div class="flex gap-4 items-start">

        <!-- Filter panel -->
        <aside class="hidden sm:flex flex-col gap-3 w-44 flex-shrink-0">
          <input type="search" bind:value={search} placeholder="Find application…"
            class="w-full bg-card border border-border rounded-md px-2.5 py-1.5 text-[12px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary" />

          <div>
            <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-1.5">Domain</div>
            {#each data.domains as d, i}
              <label class="flex items-center gap-2 px-1 py-1 cursor-pointer rounded hover:bg-muted/50 text-[12px]">
                <input type="checkbox" bind:group={domainFilterArr} value={d.id} />
                <span class="size-2 rounded-sm flex-shrink-0" style="background:{domainColor(i)}"></span>
                <span class="truncate text-foreground">{d.name}</span>
              </label>
            {/each}
          </div>

          <div>
            <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-1.5">Lifecycle</div>
            {#each ['Production','Pilot','Planned','Retiring','Decommissioned'] as lc}
              <label class="flex items-center gap-2 px-1 py-1 cursor-pointer rounded hover:bg-muted/50 text-[12px]">
                <input type="checkbox" bind:group={lcFilterArr} value={lc} />
                <span class="text-foreground">{lc}</span>
              </label>
            {/each}
          </div>

          {#if hasFilters}
            <button onclick={clearFilters} class="text-[12px] text-muted-foreground hover:text-foreground underline text-left">
              Clear all filters
            </button>
          {/if}
        </aside>

        <!-- Domain grid -->
        <div class="flex-1 min-w-0">
          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {#each data.domains.filter(d => domainFilter.size === 0 || domainFilter.has(d.id)) as d, i}
              {@const visibleApps = d.apps.filter(matchesApp)}
              {@const total = d.apps.length}

              <div class="bg-card border border-border rounded-xl overflow-hidden flex flex-col min-h-[140px]" style="box-shadow:0 1px 3px rgba(0,0,0,0.07)">
                <div class="flex items-center gap-2 px-3 py-2.5 border-b border-border bg-muted/30">
                  <span class="size-2.5 rounded-sm flex-shrink-0" style="background:{domainColor(i)}"></span>
                  <span class="text-[12.5px] font-semibold text-foreground truncate flex-1">{d.name}</span>
                  <span class="text-[11px] text-muted-foreground flex-shrink-0">{visibleApps.length}/{total}</span>
                </div>

                <div class="flex flex-col gap-1.5 p-2.5 flex-1">
                  {#if visibleApps.length === 0}
                    <div class="text-[11.5px] text-muted-foreground italic py-2 px-1">No matches</div>
                  {:else}
                    {#each visibleApps as app}
                      {@const barColor = colorBarFor(app)}
                      {@const label = colorLabelFor(app)}
                      <button
                        class="flex items-center gap-2 px-2.5 py-1.5 rounded bg-[#cff0ff] border border-[#7cc8ec] text-[11.5px] text-[#0b2936] hover:shadow-md transition-shadow cursor-pointer relative w-full text-left"
                        onclick={() => selectedApp = app}>
                        <span class="absolute left-0 top-0 bottom-0 w-[3px] rounded-l" style="background:{barColor}"></span>
                        <div class="pl-1 min-w-0 flex-1">
                          <div class="font-medium leading-tight truncate">{app.name}</div>
                          <div class="text-[10.5px] opacity-60 truncate">{app.type.replace('Application','')}</div>
                        </div>
                        {#if label}
                          <span class="text-[10px] opacity-70 flex-shrink-0 ml-1">{label}</span>
                        {/if}
                      </button>
                    {/each}
                  {/if}
                </div>
              </div>
            {/each}
          </div>

          {#if legendEntries.length > 0}
            <div class="mt-5 flex flex-wrap gap-x-4 gap-y-1.5 px-1">
              {#each legendEntries as entry}
                <div class="flex items-center gap-1.5 text-[12px]">
                  <span class="size-3 rounded flex-shrink-0" style="background:{entry.c}"></span>
                  <span class="text-muted-foreground">{entry.v}</span>
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
