<script>
  import { onMount } from 'svelte';
  import { api } from '../../lib/api.js';
  import SaveViewDialog from './SaveViewDialog.svelte';

  export let params = {};
  export let initialFilters = null;
  export let savedViewName = null;

  $: wsId = params.wsId;

  let data = null;
  let loading = true;
  let error = null;
  let overlay = 'lifecycle_status'; // overridden in onMount from initialFilters if present
  let showSaveDialog = false;

  // Tooltip
  let tooltip = null; // { app, x, y }

  // ── Colour palette ────────────────────────────────────────────────────────

  // Well-known value → colour for common property keys
  const KNOWN_COLORS = {
    lifecycle_status: {
      'Production':    '#16a34a',
      'Pilot':         '#2563eb',
      'Planned':       '#7c3aed',
      'Retiring':      '#ea580c',
      'Decommissioned':'#dc2626',
    },
    criticality: {
      'Critical': '#dc2626',
      'High':     '#ea580c',
      'Medium':   '#ca8a04',
      'Low':      '#16a34a',
    },
    deployment_model: {
      'On-Premise':  '#16a34a',
      'Public Cloud':'#2563eb',
      'SaaS':        '#0891b2',
      'Hybrid':      '#7c3aed',
    },
  };

  const PALETTE = [
    '#16a34a','#2563eb','#7c3aed','#ea580c','#dc2626',
    '#0891b2','#db2777','#ca8a04','#0284c7','#9333ea',
  ];
  const UNSET_COLOR = '#94a3b8';

  // Per-overlay, assign consistent colours to each distinct value
  function buildColorMap(overlay, l1List) {
    const known = KNOWN_COLORS[overlay] ?? {};
    const seen = new Set(Object.keys(known));
    let idx = 0;
    const map = { ...known };

    for (const l1 of l1List) {
      for (const l2 of l1.l2) {
        for (const app of l2.apps) {
          const v = app.properties?.[overlay] ?? '';
          if (v && !map[v]) {
            // skip palette slots already used by known values
            while (PALETTE[idx] && Object.values(known).includes(PALETTE[idx])) idx++;
            map[v] = PALETTE[idx % PALETTE.length];
            idx++;
          }
        }
      }
    }
    return map;
  }

  $: colorMap = data ? buildColorMap(overlay, data.l1) : {};

  // Chip uses a colored left border (same saturated color as legend dot)
  // with a neutral background — directly connects chip to legend visually.
  function chipStyle(app, ov) {
    const v = app.properties?.[ov] ?? '';
    const color = v ? (colorMap[v] ?? '#94a3b8') : UNSET_COLOR;
    return `border-left: 3px solid ${color}; background:${color}18; color:#1e293b; padding-left:7px;`;
  }

  // ── Legend entries (distinct values present) ──────────────────────────────
  $: legendEntries = (() => {
    if (!data) return [];
    const seen = new Map();
    for (const l1 of data.l1) {
      for (const l2 of l1.l2) {
        for (const app of l2.apps) {
          const v = app.properties?.[overlay] ?? '(unset)';
          const c = v === '(unset)' ? UNSET_COLOR : (colorMap[v] ?? '#6b7280');
          if (!seen.has(v)) seen.set(v, c);
        }
      }
    }
    // Sort: known order first if lifecycle, then alpha, (unset) last
    return [...seen.entries()]
      .map(([v, c]) => ({ value: v, color: c }))
      .sort((a, b) => {
        if (a.value === '(unset)') return 1;
        if (b.value === '(unset)') return -1;
        return a.value.localeCompare(b.value);
      });
  })();

  // ── Prop label ────────────────────────────────────────────────────────────
  const PROP_LABELS = {
    lifecycle_status: 'Lifecycle Status',
    deployment_model: 'Deployment Model',
    criticality:      'Business Criticality',
    vendor:           'Vendor',
    business_owner:   'Business Owner',
    user_count:       'User Count',
  };
  function propLabel(k) {
    return PROP_LABELS[k] ?? k.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
  }

  // ── Tooltip ───────────────────────────────────────────────────────────────
  function showTooltip(e, app) {
    const rect = e.currentTarget.getBoundingClientRect();
    tooltip = { app, x: rect.right + 8, y: rect.top };
  }
  function hideTooltip() { tooltip = null; }

  $: saveFilters = { overlay };

  // ── Load ──────────────────────────────────────────────────────────────────
  onMount(async () => {
    overlay = initialFilters?.overlay ?? 'lifecycle_status';
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

<!-- App chip tooltip -->
{#if tooltip}
  <div
    class="fixed z-50 bg-popover border border-border rounded-lg shadow-lg p-3 w-56 pointer-events-none"
    style="left:{Math.min(tooltip.x, window.innerWidth - 232)}px; top:{Math.min(tooltip.y, window.innerHeight - 200)}px"
  >
    <div class="text-[13px] font-semibold text-foreground mb-1">{tooltip.app.name}</div>
    <div class="text-[11px] text-muted-foreground mb-2">{tooltip.app.type.replace('Application', '')}</div>
    {#if Object.keys(tooltip.app.properties ?? {}).length > 0}
      <div class="space-y-1">
        {#each Object.entries(tooltip.app.properties) as [k, v]}
          <div class="flex justify-between text-[11px]">
            <span class="text-muted-foreground">{propLabel(k)}</span>
            <span class="font-medium text-foreground ml-2 text-right">{v}</span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/if}

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
        <div class="text-muted-foreground text-[13px] mt-0.5">Capabilities mapped to realizing applications</div>
      </div>

      <div class="flex items-center gap-2 flex-wrap">
        <!-- Overlay selector -->
        {#if data.properties?.length > 0}
          <span class="text-[12px] text-muted-foreground">Overlay</span>
          <select
            bind:value={overlay}
            class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
          >
            {#each data.properties as p}
              <option value={p}>{propLabel(p)}</option>
            {/each}
          </select>
        {/if}
        {#if !savedViewName}
          <button
            onclick={() => showSaveDialog = true}
            class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-border text-[12px] text-muted-foreground hover:text-foreground hover:border-primary transition-colors"
          >
            ⊕ Save view
          </button>
        {/if}
      </div>
    </div>

    <SaveViewDialog
      bind:open={showSaveDialog}
      {wsId}
      viewType="application-landscape"
      filters={saveFilters}
    />

    <!-- Colour legend -->
    {#if legendEntries.length > 0}
      <div class="flex flex-wrap gap-x-4 gap-y-1.5 mb-5 px-1">
        {#each legendEntries as entry}
          <div class="flex items-center gap-1.5 text-[12px]">
            <span class="size-3 rounded flex-shrink-0" style="background:{entry.color}"></span>
            <span class="text-muted-foreground">{entry.value}</span>
          </div>
        {/each}
      </div>
    {/if}

    <!-- Landscape grid -->
    {#if data.l1.length === 0}
      <div class="text-center py-16 text-muted-foreground">
        <div class="text-[40px] mb-3">📭</div>
        <p class="text-[14px]">No capability hierarchy found. Import a model with Capability elements first.</p>
      </div>
    {:else}
      <div class="space-y-4">
        {#each data.l1 as l1}
          {@const totalApps = l1.l2.reduce((s, l2) => s + l2.apps.length, 0)}
          <div class="border border-slate-300 rounded-xl overflow-hidden" style="box-shadow: 0 1px 3px rgba(0,0,0,0.08), 0 0 0 1px rgba(0,0,0,0.04);">
            <!-- L1 header -->
            <div class="bg-slate-100 border-b border-slate-200 px-4 py-2.5 flex items-center gap-3">
              <span class="text-[12px] font-bold text-slate-600 tracking-[0.8px] uppercase">{l1.name}</span>
              <span class="text-[11px] text-slate-400 ml-auto">{totalApps} app{totalApps !== 1 ? 's' : ''}</span>
            </div>

            <!-- L2 rows -->
            <div class="divide-y divide-slate-200">
              {#each l1.l2 as l2}
                <div class="flex items-start gap-3 px-4 py-2.5 hover:bg-muted/20 transition-colors">
                  <!-- L2 name + count -->
                  <div class="w-52 flex-shrink-0 pt-0.5">
                    <span class="text-[12px] text-foreground font-medium">{l2.name}</span>
                    <span class="ml-1.5 text-[11px] text-muted-foreground">{l2.apps.length}</span>
                  </div>

                  <!-- App chips -->
                  <div class="flex flex-wrap gap-1.5 flex-1">
                    {#if l2.apps.length === 0}
                      <span class="text-[11px] text-muted-foreground italic">No applications</span>
                    {:else}
                      {#each l2.apps as app}
                        <button
                          class="inline-flex items-center px-2.5 py-1 rounded text-[11px] font-medium transition-opacity hover:opacity-80 cursor-default"
                          style="{chipStyle(app, overlay)}"
                          onmouseenter={(e) => showTooltip(e, app)}
                          onmouseleave={hideTooltip}
                          onfocus={(e) => showTooltip(e, app)}
                          onblur={hideTooltip}
                        >
                          {app.name}
                        </button>
                      {/each}
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>
