<script>
  import { onMount } from 'svelte';
  import { api } from '../../lib/api.js';
  import { ArcChart } from 'layerchart';

  export let params = {};

  $: wsId = params.wsId;

  let data = null;
  let loading = true;
  let error = null;
  let selectedCapability = 'all';
  let selectedApp = null; // source_id of highlighted app

  // Tooltip state
  let tooltip = null; // { key, value, apps, x, y }

  const PROP_LABELS = {
    lifecycle_status: 'Lifecycle Status',
    deployment_model: 'Deployment Model',
    criticality:      'Business Criticality',
    vendor:           'Vendor',
    business_owner:   'Business Owner',
    user_count:       'User Count Range',
  };

  const PROP_ORDER = [
    'lifecycle_status',
    'deployment_model',
    'criticality',
    'vendor',
    'business_owner',
    'user_count',
  ];

  const PALETTE = [
    '#16a34a','#2563eb','#7c3aed','#ea580c','#dc2626',
    '#0891b2','#db2777','#ca8a04','#0284c7','#9333ea',
  ];
  const UNSET_COLOR = '#94a3b8';

  function colorFor(value, index) {
    return value === '(unset)' ? UNSET_COLOR : PALETTE[index % PALETTE.length];
  }

  function propLabel(key) {
    return PROP_LABELS[key] ?? key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
  }

  function sortedPropKeys(properties) {
    if (!properties) return [];
    const known = PROP_ORDER.filter(k => k in properties);
    const other = Object.keys(properties).filter(k => !PROP_ORDER.includes(k)).sort();
    return [...known, ...other];
  }

  function arcData(buckets) {
    return buckets.map((b, i) => ({
      key: b.value, label: b.value, value: b.count, color: colorFor(b.value, i),
    }));
  }

  function appsForSlice(key, value) {
    return (data?.apps ?? []).filter(a => {
      const v = a.properties?.[key] ?? '';
      return value === '(unset)' ? v === '' : v === value;
    });
  }

  // Show tooltip anchored to the legend row element
  function showTooltip(e, key, slice) {
    const apps = appsForSlice(key, slice.key);
    const rect = e.currentTarget.getBoundingClientRect();
    // Position to the right of the legend row, or left if near edge
    const x = rect.right + 8;
    const y = rect.top;
    tooltip = { key, value: slice.key, color: slice.color, apps, x, y };
  }

  function hideTooltip() {
    tooltip = null;
  }

  // Color for an app's value within a given property key
  function appDotColor(app, key) {
    const buckets = data?.properties?.[key];
    if (!buckets) return '#6b7280';
    const v = app.properties?.[key] ?? '';
    const effectiveVal = v === '' ? '(unset)' : v;
    const idx = buckets.findIndex(b => b.value === effectiveVal);
    return colorFor(effectiveVal, idx);
  }

  async function load(capability) {
    loading = true;
    error = null;
    selectedApp = null;
    tooltip = null;
    try {
      const qs = capability && capability !== 'all'
        ? '?capability=' + encodeURIComponent(capability)
        : '';
      data = await api.get('/workspaces/' + wsId + '/views/application-dashboard/stats' + qs);
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  onMount(() => load('all'));

  function onCapabilityChange(e) {
    selectedCapability = e.target.value;
    load(selectedCapability);
  }
</script>

<!-- Rich tooltip (rendered at body level via fixed position) -->
{#if tooltip}
  <div
    class="fixed z-50 bg-popover border border-border rounded-lg shadow-lg p-3 w-64 pointer-events-none"
    style="left:{Math.min(tooltip.x, window.innerWidth - 272)}px; top:{Math.min(tooltip.y, window.innerHeight - 320)}px"
  >
    <div class="flex items-center gap-2 mb-2 pb-2 border-b border-border">
      <span class="size-2.5 rounded-full flex-shrink-0" style="background:{tooltip.color}"></span>
      <span class="text-[13px] font-semibold text-foreground truncate">{tooltip.value}</span>
      <span class="ml-auto text-[12px] text-muted-foreground">{tooltip.apps.length}</span>
    </div>
    <div class="space-y-1 max-h-52 overflow-y-auto">
      {#each tooltip.apps as app}
        <div class="flex items-center gap-2 text-[12px]">
          <span class="size-1.5 rounded-full flex-shrink-0 bg-muted-foreground/50"></span>
          <span class="truncate text-foreground" title={app.name}>{app.name}</span>
        </div>
      {/each}
    </div>
  </div>
{/if}

<div class="content">
  {#if loading && !data}
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="mt-6 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">Error: {error}</div>
  {:else if data}

    <!-- Header -->
    <div class="flex items-start justify-between gap-4 mb-6 flex-wrap">
      <div>
        <h1 class="text-[18px] font-semibold">Application Dashboard</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">
          Applications in scope: <span class="font-semibold text-foreground">{data.total_apps}</span>
          {#if loading}<span class="ml-2 text-[11px]">Updating…</span>{/if}
        </div>
      </div>

      {#if data.capabilities?.length > 0}
        <div class="flex items-center gap-2">
          <label for="cap-filter" class="text-[12px] text-muted-foreground whitespace-nowrap">Business Capability</label>
          <select
            id="cap-filter"
            onchange={onCapabilityChange}
            class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary min-w-[200px]"
          >
            <option value="all" selected={selectedCapability === 'all'}>All</option>
            {#each data.capabilities as cap}
              <option value={cap} selected={selectedCapability === cap}>{cap}</option>
            {/each}
          </select>
        </div>
      {/if}
    </div>

    {#if data.total_apps === 0}
      <div class="text-center py-16 px-6 text-muted-foreground">
        <div class="text-[40px] mb-3.5">📭</div>
        <p class="text-[14px] leading-relaxed">No applications found for this scope.</p>
      </div>
    {:else}
      <!-- Two-column layout: app list | donuts -->
      <div class="flex gap-5 items-start">

        <!-- App list panel -->
        <div class="flex-shrink-0 w-48 bg-card border border-border rounded-xl overflow-hidden shadow-sm">
          <div class="text-[10px] font-bold tracking-[0.6px] uppercase text-muted-foreground px-3 py-2.5 border-b border-border">
            Applications ({data.apps.length})
          </div>
          <div class="overflow-y-auto" style="max-height: 560px;">
            {#each data.apps as app}
              <button
                class="w-full text-left px-3 py-1.5 text-[12px] flex items-center gap-2 border-b border-border/30 transition-colors hover:bg-muted/50
                  {selectedApp === app.id ? 'bg-primary/10 text-primary font-medium' : 'text-foreground'}"
                onclick={() => { selectedApp = selectedApp === app.id ? null : app.id; tooltip = null; }}
              >
                <span class="size-1.5 rounded-full flex-shrink-0 bg-muted-foreground/40"></span>
                <span class="truncate" title={app.name}>{app.name}</span>
              </button>
            {/each}
          </div>
        </div>

        <!-- Donuts grid -->
        <div class="flex-1 grid grid-cols-1 xl:grid-cols-2 gap-4">
          {#each sortedPropKeys(data.properties) as key}
            {@const buckets = data.properties[key]}
            {@const slices = arcData(buckets)}
            <div class="bg-card border border-border rounded-xl p-4 shadow-sm">
              <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-3">{propLabel(key)}</div>
              <div class="flex gap-4 items-center">
                <!-- Donut -->
                <div class="size-[150px] flex-shrink-0">
                  <ArcChart
                    data={slices}
                    key="key"
                    label="label"
                    value="value"
                    c="color"
                    innerRadius={0.60}
                    padAngle={0.02}
                    cornerRadius={2}
                  />
                </div>
                <!-- Legend with hover tooltip -->
                <div class="flex flex-col gap-0.5 min-w-0 flex-1">
                  {#each slices as slice}
                    {@const isHighlighted = selectedApp != null
                      && (data.apps.find(a => a.id === selectedApp)?.properties?.[key] ?? '') === (slice.key === '(unset)' ? '' : slice.key)}
                    <div
                      role="button"
                      tabindex="0"
                      class="flex items-center gap-2 text-[12px] rounded px-1.5 py-1 cursor-default transition-colors hover:bg-muted/60
                        {isHighlighted ? 'bg-primary/8 ring-1 ring-inset ring-primary/30' : ''}"
                      onmouseenter={(e) => showTooltip(e, key, slice)}
                      onmouseleave={hideTooltip}
                      onfocus={(e) => showTooltip(e, key, slice)}
                      onblur={hideTooltip}
                    >
                      <span class="size-2.5 rounded-full flex-shrink-0" style="background:{slice.color}"></span>
                      <span class="text-muted-foreground truncate flex-1" title={slice.label}>{slice.label}</span>
                      <span class="font-semibold tabular-nums text-foreground">{slice.value}</span>
                    </div>
                  {/each}
                </div>
              </div>
            </div>
          {/each}
        </div>

      </div>
    {/if}
  {/if}
</div>
