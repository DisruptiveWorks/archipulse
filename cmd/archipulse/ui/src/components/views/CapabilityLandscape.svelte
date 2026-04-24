<script>
  import { onMount } from 'svelte';
  import { api } from '../../lib/api.js';
  import SaveViewDialog from './SaveViewDialog.svelte';
  import AppDetailPanel from './AppDetailPanel.svelte';
  import ViewInfoDialog from './ViewInfoDialog.svelte';

  const { params = {}, initialFilters = null, savedViewName = null } = $props();

  const wsId = $derived(params.wsId);

  let data        = $state(null);
  let loading     = $state(true);
  let error       = $state(null);
  let showSaveDialog = $state(false);
  let selectedApp = $state(null);
  let showInfo    = $state(false);

  // ── Filter / display state ────────────────────────────────────────────────
  let overlay        = $state('lifecycle_status');
  let heatmap        = $state('appCount');
  let showRetired    = $state(true);
  let search         = $state('');
  let groupFilterArr = $state([]);
  let collapsed      = $state(new Set());

  const groupFilter  = $derived(new Set(groupFilterArr));
  const saveFilters  = $derived({ overlay, heatmap, groups: groupFilterArr });

  // ── Overlay colour logic ──────────────────────────────────────────────────
  const OVERLAY_LC = {
    'Production':     { bg: '#d1fae5', border: '#10b981' },
    'Pilot':          { bg: '#ede9fe', border: '#8b5cf6' },
    'Planned':        { bg: '#e0f2fe', border: '#0ea5e9' },
    'Retiring':       { bg: '#fef3c7', border: '#f59e0b' },
    'Decommissioned': { bg: '#f1f5f9', border: '#94a3b8' },
  };
  const OVERLAY_CRIT = {
    'Critical': { bg: '#fee2e2', border: '#dc2626' },
    'High':     { bg: '#fef3c7', border: '#f59e0b' },
    'Medium':   { bg: '#e0f2fe', border: '#0ea5e9' },
    'Low':      { bg: '#f1f5f9', border: '#64748b' },
  };
  const OVERLAY_DEPLOY = {
    'On-Premise':    { bg: '#f1f5f9', border: '#64748b' },
    'Public Cloud':  { bg: '#ede9fe', border: '#6366f1' },
    'AWS':           { bg: '#fef3c7', border: '#f59e0b' },
    'Azure':         { bg: '#e0f2fe', border: '#0ea5e9' },
    'SaaS':          { bg: '#d1fae5', border: '#10b981' },
    'Hybrid':        { bg: '#fdf4ff', border: '#a855f7' },
    'Private Cloud': { bg: '#ede9fe', border: '#6366f1' },
  };

  const overlayColor = $derived((app) => {
    const p = app.properties ?? {};
    if (overlay === 'none') return { bg: '#cff0ff', border: '#7cc8ec' };
    if (overlay === 'lifecycle_status') return OVERLAY_LC[p.lifecycle_status] ?? { bg: '#f8fafc', border: '#94a3b8' };
    if (overlay === 'criticality')      return OVERLAY_CRIT[p.criticality]    ?? { bg: '#f8fafc', border: '#94a3b8' };
    if (overlay === 'deployment_model') return OVERLAY_DEPLOY[p.deployment_model] ?? { bg: '#f8fafc', border: '#94a3b8' };
    if (overlay === 'fit') {
      const critScore = { 'Critical': 3, 'High': 2, 'Medium': 1, 'Low': 0 }[p.criticality] ?? 0;
      const lcScore   = { 'Production': 2, 'Pilot': 1, 'Planned': 0, 'Retiring': -1, 'Decommissioned': -2 }[p.lifecycle_status] ?? 0;
      const score = critScore + lcScore;
      if (score >= 4) return { bg: '#d1fae5', border: '#10b981' };
      if (score >= 2) return { bg: '#e0f2fe', border: '#0ea5e9' };
      if (score >= 0) return { bg: '#fef3c7', border: '#f59e0b' };
      return { bg: '#fee2e2', border: '#dc2626' };
    }
    return { bg: '#cff0ff', border: '#7cc8ec' };
  });

  const legendEntries = $derived((() => {
    if (overlay === 'lifecycle_status') return [
      { l: 'Production',     c: '#10b981' }, { l: 'Pilot',          c: '#8b5cf6' },
      { l: 'Planned',        c: '#0ea5e9' }, { l: 'Retiring',       c: '#f59e0b' },
      { l: 'Decommissioned', c: '#94a3b8' },
    ];
    if (overlay === 'criticality') return [
      { l: 'Critical', c: '#dc2626' }, { l: 'High',   c: '#f59e0b' },
      { l: 'Medium',   c: '#0ea5e9' }, { l: 'Low',    c: '#64748b' },
    ];
    if (overlay === 'deployment_model') return [
      { l: 'On-Premise',    c: '#64748b' }, { l: 'Public Cloud', c: '#6366f1' },
      { l: 'SaaS',          c: '#10b981' }, { l: 'Hybrid',       c: '#a855f7' },
      { l: 'Private Cloud', c: '#6366f1' }, { l: 'AWS',          c: '#f59e0b' },
      { l: 'Azure',         c: '#0ea5e9' },
    ];
    if (overlay === 'fit') return [
      { l: 'Excellent fit', c: '#10b981' }, { l: 'Good fit',     c: '#0ea5e9' },
      { l: 'Review needed', c: '#f59e0b' }, { l: 'Poor fit',     c: '#dc2626' },
    ];
    return [];
  })());

  // ── Heatmap logic ─────────────────────────────────────────────────────────
  function groupStats(l1) {
    let totalApps = new Set();
    let gaps = 0;
    for (const l2 of l1.l2) {
      if (l2.apps.length === 0) gaps++;
      for (const a of l2.apps) totalApps.add(a.id);
    }
    return { total: totalApps.size, gaps, subs: l1.l2.length };
  }

  const heatmapBg = $derived((l1) => {
    if (heatmap === 'none' || !l1) return null;
    const s = groupStats(l1);
    if (heatmap === 'appCount') {
      const pct = Math.min(1, s.total / 8);
      return `rgba(16,185,129,${(0.12 + pct * 0.45).toFixed(2)})`;
    }
    if (heatmap === 'gap') {
      const pct = s.subs ? s.gaps / s.subs : 0;
      return `rgba(220,38,38,${(0.08 + pct * 0.55).toFixed(2)})`;
    }
    if (heatmap === 'avgCrit') {
      let total = 0, n = 0;
      for (const l2 of l1.l2) {
        for (const a of l2.apps) {
          total += { 'Critical': 3, 'High': 2, 'Medium': 1, 'Low': 0 }[a.properties?.criticality] ?? 0;
          n++;
        }
      }
      const pct = n ? (total / n) / 3 : 0;
      return `rgba(249,115,22,${(0.10 + pct * 0.50).toFixed(2)})`;
    }
    return null;
  });

  // ── Helpers ───────────────────────────────────────────────────────────────
  function toggleCollapsed(id) {
    const next = new Set(collapsed);
    next.has(id) ? next.delete(id) : next.add(id);
    collapsed = next;
  }

  function isRetired(app) {
    return app.properties?.lifecycle_status === 'Decommissioned' || app.properties?.lifecycle_status === 'Retiring';
  }

  // ── Load ──────────────────────────────────────────────────────────────────
  onMount(async () => {
    overlay        = initialFilters?.overlay  ?? 'lifecycle_status';
    heatmap        = initialFilters?.heatmap  ?? 'appCount';
    groupFilterArr = initialFilters?.groups   ?? [];
    loading = true;
    try {
      data = await api.get('/workspaces/' + wsId + '/views/capability-landscape/map');
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
        <h1 class="text-[18px] font-semibold">{savedViewName ?? 'Capability Landscape'}</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">Business capabilities mapped to realizing applications</div>
      </div>
      <div class="flex items-center gap-2 flex-wrap">
        <span class="text-[12px] text-muted-foreground">Overlay</span>
        <select bind:value={overlay}
          class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary">
          <option value="none">None</option>
          <option value="lifecycle_status">Lifecycle Status</option>
          <option value="criticality">Business Criticality</option>
          <option value="deployment_model">Deployment Model</option>
          <option value="fit">Application Fit</option>
        </select>
        <span class="text-[12px] text-muted-foreground">Heatmap</span>
        <select bind:value={heatmap}
          class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary">
          <option value="none">None</option>
          <option value="appCount">App Coverage</option>
          <option value="gap">Gap Analysis</option>
          <option value="avgCrit">Avg. Criticality</option>
        </select>
        {#if !savedViewName}
          <button onclick={() => showSaveDialog = true}
            class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-border text-[12px] text-muted-foreground hover:text-foreground hover:border-primary transition-colors">
            ⊕ Save view
          </button>
        {/if}
        <ViewInfoDialog title="Capability Landscape — setup guide" bind:open={showInfo}>
          <p>This view maps <strong>business capabilities</strong> (L1 groups → L2 capabilities) to the applications that realize them, showing coverage, gaps, and overlaps across your capability model.</p>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Required model elements</div>
            <table class="w-full text-[12px] border border-border rounded-md overflow-hidden">
              <thead>
                <tr class="bg-muted/60">
                  <th class="text-left px-3 py-1.5 font-semibold">Element type</th>
                  <th class="text-left px-3 py-1.5 font-semibold">Role</th>
                </tr>
              </thead>
              <tbody>
                {#each [
                  ['Capability', 'L1 groups and L2 capabilities (Strategy layer)'],
                  ['ApplicationComponent / ApplicationService', 'Applications that realize capabilities'],
                ] as [t, r]}
                  <tr class="border-t border-border">
                    <td class="px-3 py-1.5 font-mono text-[11.5px]">{t}</td>
                    <td class="px-3 py-1.5 text-muted-foreground">{r}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Required relationships</div>
            <table class="w-full text-[12px] border border-border rounded-md overflow-hidden">
              <thead>
                <tr class="bg-muted/60">
                  <th class="text-left px-3 py-1.5 font-semibold">Relationship</th>
                  <th class="text-left px-3 py-1.5 font-semibold">From → To</th>
                </tr>
              </thead>
              <tbody>
                <tr class="border-t border-border">
                  <td class="px-3 py-1.5 font-mono">Composition</td>
                  <td class="px-3 py-1.5 text-muted-foreground">L1 Capability → L2 Capability (creates the hierarchy)</td>
                </tr>
                <tr class="border-t border-border">
                  <td class="px-3 py-1.5 font-mono">Realization</td>
                  <td class="px-3 py-1.5 text-muted-foreground">ApplicationComponent → Capability (links apps to capabilities)</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="bg-amber-50 border border-amber-200 rounded-md px-3 py-2 text-[12px] text-amber-800">
            <strong>Note:</strong> This view uses ArchiMate <span class="font-mono">Capability</span> elements (Strategy layer). If your model uses <span class="font-mono">BusinessFunction</span> elements instead, use the <strong>Capability Tree</strong> view — it supports both types.
          </div>

          <div>
            <div class="font-semibold text-[12px] uppercase tracking-wide text-muted-foreground mb-1.5">Optional app properties (for overlays &amp; heatmaps)</div>
            <p class="text-[12px] text-muted-foreground">Set these on your ApplicationComponent / ApplicationService elements:</p>
            <table class="w-full text-[12px] border border-border rounded-md overflow-hidden mt-1.5">
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
                ] as [k, v]}
                  <tr class="border-t border-border">
                    <td class="px-3 py-1.5 font-mono">{k}</td>
                    <td class="px-3 py-1.5 text-muted-foreground">{v}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </ViewInfoDialog>
      </div>
    </div>

    <SaveViewDialog bind:open={showSaveDialog} {wsId} viewType="capability-landscape" filters={saveFilters} />

    <div class="flex gap-4 items-start">

      <!-- Left filter panel -->
      <aside class="hidden sm:flex flex-col gap-3 w-44 flex-shrink-0">
        <input type="search" bind:value={search} placeholder="Find capability…"
          class="w-full bg-card border border-border rounded-md px-2.5 py-1.5 text-[12px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary" />

        {#if data.l1.length > 0}
          <div>
            <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-1.5">Capability groups</div>
            {#each data.l1 as l1}
              <label class="flex items-center gap-2 px-1 py-1 cursor-pointer rounded hover:bg-muted/50 text-[12px]">
                <input type="checkbox" class="rounded" bind:group={groupFilterArr} value={l1.id} />
                <span class="truncate text-foreground">{l1.name}</span>
              </label>
            {/each}
          </div>
        {/if}

        <div>
          <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-1.5">Display</div>
          <label class="flex items-center gap-2 px-1 py-1 cursor-pointer rounded hover:bg-muted/50 text-[12px]">
            <input type="checkbox" bind:checked={showRetired} />
            <span class="text-foreground">Show retiring/decommissioned</span>
          </label>
          <button class="mt-1 text-[11px] text-muted-foreground hover:text-foreground underline px-1"
            onclick={() => collapsed = collapsed.size ? new Set() : new Set(data.l1.map(l => l.id))}>
            {collapsed.size ? 'Expand all' : 'Collapse all'}
          </button>
        </div>

        {#if data.l1.length > 0}
          <div class="text-[11px] text-muted-foreground border border-dashed border-border rounded-md px-2 py-1.5">
            {data.l1.reduce((s, l) => s + l.l2.length, 0)} capabilities ·
            {data.l1.reduce((s, l) => s + l.l2.reduce((ss, l2) => ss + l2.apps.length, 0), 0)} realizations
          </div>
        {/if}
      </aside>

      <!-- Main canvas -->
      <div class="flex-1 min-w-0">
        {#if data.l1.length === 0}
          <div class="text-center py-16 text-muted-foreground">
            <div class="text-[40px] mb-3">◈</div>
            <p class="text-[14px]">No capability hierarchy found.</p>
            <p class="text-[13px] mt-1 opacity-70">Import a model with Capability or BusinessFunction elements.</p>
          </div>
        {:else}
          {@const visibleL1 = data.l1.filter(l1 =>
            (groupFilter.size === 0 || groupFilter.has(l1.id)) &&
            (!search || l1.name.toLowerCase().includes(search.toLowerCase()) ||
             l1.l2.some(l2 => l2.name.toLowerCase().includes(search.toLowerCase())))
          )}
          <div class="space-y-3">
            {#each visibleL1 as l1}
              {@const stats = groupStats(l1)}
              {@const hmBg = heatmapBg(l1)}
              {@const isCollapsed = collapsed.has(l1.id)}

              <div class="border border-border rounded-xl overflow-hidden shadow-sm"
                style={hmBg ? `background:${hmBg}` : 'background:var(--card)'}>

                <div class="flex items-center gap-2.5 px-4 py-2.5 border-b border-border cursor-pointer select-none"
                  style="background:rgba(255,255,255,0.6)"
                  onclick={() => toggleCollapsed(l1.id)}
                  role="button" tabindex="0"
                  onkeydown={e => e.key === 'Enter' && toggleCollapsed(l1.id)}>
                  <span class="text-muted-foreground text-[12px] transition-transform {isCollapsed ? '' : 'rotate-90'}" style="display:inline-block">▶</span>
                  <span class="text-[12px] font-bold text-foreground tracking-[0.6px] uppercase">{l1.name}</span>
                  {#if heatmap === 'gap' && stats.gaps > 0}
                    <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-red-50 text-red-700 border border-dashed border-red-200">
                      {stats.gaps} gap{stats.gaps !== 1 ? 's' : ''}
                    </span>
                  {/if}
                  <span class="ml-auto text-[12px] text-muted-foreground">
                    {stats.total} app{stats.total !== 1 ? 's' : ''} · {l1.l2.length} capabilit{l1.l2.length !== 1 ? 'ies' : 'y'}
                  </span>
                </div>

                {#if !isCollapsed}
                  <div class="divide-y divide-border">
                    {#each l1.l2.filter(l2 => !search || l2.name.toLowerCase().includes(search.toLowerCase())) as l2}
                      {@const visibleApps = showRetired ? l2.apps : l2.apps.filter(a => !isRetired(a))}
                      {@const isGap = l2.apps.length === 0}
                      {@const isDup = l2.apps.length >= 3}

                      <div class="flex items-start gap-3 px-4 py-2 hover:bg-white/40 transition-colors">
                        <div class="w-52 flex-shrink-0 pt-0.5">
                          <span class="text-[12.5px] font-medium text-foreground">{l2.name}</span>
                          <span class="ml-1.5 text-[11px] {isGap ? 'text-red-600 font-semibold' : 'text-muted-foreground'}">{l2.apps.length}</span>
                        </div>

                        <div class="flex flex-wrap gap-1.5 flex-1 items-center">
                          {#if isGap && heatmap === 'gap'}
                            <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] bg-red-50 text-red-700 border border-dashed border-red-200">
                              Gap — no realization
                            </span>
                          {:else if visibleApps.length === 0}
                            <span class="text-[11px] text-muted-foreground italic">No applications</span>
                          {/if}
                          {#if isDup && heatmap === 'gap' && !isGap}
                            <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] bg-amber-50 text-amber-700 border border-dashed border-amber-200">
                              Possible overlap
                            </span>
                          {/if}
                          {#each visibleApps as app}
                            {@const col = overlayColor(app)}
                            {@const retired = isRetired(app)}
                            <button
                              class="inline-flex items-center relative px-2.5 py-1 pl-[14px] rounded text-[11.5px] font-medium transition-shadow hover:shadow-sm cursor-pointer {retired ? 'opacity-45 line-through' : ''}"
                              style="background:{col.bg}; border:1px solid {col.border}; color:#0b2936;"
                              title="{app.name} · {app.properties?.lifecycle_status ?? ''} · {app.properties?.criticality ?? ''}"
                              onclick={() => selectedApp = app}>
                              <span class="absolute left-0 top-0 bottom-0 w-[3px] rounded-l"
                                style="background:{col.border}"></span>
                              {app.name}
                            </button>
                          {/each}
                        </div>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {/each}
          </div>

          {#if overlay !== 'none' && legendEntries.length > 0}
            <div class="mt-5 flex flex-wrap gap-x-4 gap-y-1.5 px-1">
              {#each legendEntries as entry}
                <div class="flex items-center gap-1.5 text-[12px]">
                  <span class="size-3 rounded flex-shrink-0" style="background:{entry.c}"></span>
                  <span class="text-muted-foreground">{entry.l}</span>
                </div>
              {/each}
            </div>
          {/if}
        {/if}
      </div>
    </div>
  {/if}
</div>

<AppDetailPanel app={selectedApp} {wsId} on:close={() => selectedApp = null} />
