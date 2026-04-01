<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { ArcChart } from 'layerchart';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';

  export let params = {};

  $: wsId = params.wsId;

  let data = null;
  let loading = true;
  let error = null;
  let selectedCapability = 'all';
  let selectedApp = null;   // source_id of highlighted app, or null

  // Drill-down modal state
  let drillKey = null;
  let drillValue = null;
  let drillOpen = false;

  // Human-readable labels for known property keys
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
    '#4ade80','#60a5fa','#a78bfa','#fb923c','#f87171',
    '#34d399','#f472b6','#facc15','#38bdf8','#c084fc',
  ];
  const UNSET_COLOR = '#4b5563';

  function colorFor(value, index) {
    if (value === '(unset)') return UNSET_COLOR;
    return PALETTE[index % PALETTE.length];
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
      key:   b.value,
      label: b.value,
      value: b.count,
      color: colorFor(b.value, i),
    }));
  }

  // Apps that match a given property key + value
  function appsForSlice(key, value) {
    if (!data?.apps) return [];
    return data.apps.filter(a => {
      const v = a.properties[key] ?? '';
      if (value === '(unset)') return v === '';
      return v === value;
    });
  }

  // Color for a specific app + property key (for highlighting in the panel)
  function appSliceColor(app, key) {
    const buckets = data?.properties?.[key];
    if (!buckets) return '#6b7280';
    const val = app.properties[key] ?? '';
    const effectiveVal = val === '' ? '(unset)' : val;
    const idx = buckets.findIndex(b => b.value === effectiveVal);
    return colorFor(effectiveVal, idx);
  }

  function openDrill(key, value) {
    drillKey = key;
    drillValue = value;
    drillOpen = true;
  }

  async function load(capability) {
    loading = true;
    error = null;
    selectedApp = null;
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

<!-- Drill-down modal -->
<Dialog.Root bind:open={drillOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="max-w-md">
      <Dialog.Header>
        <Dialog.Title>{drillValue} ({drillKey ? appsForSlice(drillKey, drillValue).length : 0})</Dialog.Title>
        <Dialog.Description>{propLabel(drillKey ?? '')} · {drillValue}</Dialog.Description>
      </Dialog.Header>
      <div class="mt-2 max-h-72 overflow-y-auto divide-y divide-border">
        {#each appsForSlice(drillKey, drillValue) as app}
          <div class="py-2.5 px-1 flex items-center gap-2 text-[13px]">
            <span
              class="size-2 rounded-full flex-shrink-0"
              style="background:{appSliceColor(app, drillKey)}"
            ></span>
            <span class="flex-1 truncate" title={app.name}>{app.name}</span>
            <span class="text-[11px] text-muted-foreground">{app.type.replace('Application','')}</span>
          </div>
        {/each}
      </div>
      <Dialog.Footer class="mt-4">
        <Button variant="outline" onclick={() => drillOpen = false}>Close</Button>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<div class="content" style="padding-left: 0; padding-right: 0;">
  {#if loading && !data}
    <div class="flex items-center gap-2 text-muted-foreground py-6 px-8">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="mt-6 mx-8 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">Error: {error}</div>
  {:else if data}
    <!-- Top bar -->
    <div class="flex items-center justify-between gap-4 px-6 py-4 border-b border-border flex-wrap">
      <div>
        <h1 class="text-[17px] font-semibold">Application Dashboard</h1>
        <div class="text-muted-foreground text-[12px] mt-0.5">
          Applications in scope: <span class="font-semibold text-foreground">{data.total_apps}</span>
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
      <div class="flex min-h-0">

        <!-- Left: app list -->
        <div class="w-52 flex-shrink-0 border-r border-border overflow-y-auto" style="max-height: calc(100vh - 140px);">
          {#if loading}
            <div class="flex items-center gap-2 text-muted-foreground text-[12px] px-3 py-3">
              <div class="size-3 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
              Updating…
            </div>
          {:else}
            {#each data.apps as app}
              <button
                class="w-full text-left px-3 py-2 text-[12px] flex items-center gap-2 transition-colors hover:bg-muted/50 border-b border-border/40 {selectedApp === app.id ? 'bg-primary/10 text-primary' : 'text-foreground'}"
                onclick={() => selectedApp = selectedApp === app.id ? null : app.id}
              >
                <span class="truncate flex-1" title={app.name}>{app.name}</span>
              </button>
            {/each}
          {/if}
        </div>

        <!-- Right: donuts grid -->
        <div class="flex-1 overflow-y-auto px-5 py-5" style="max-height: calc(100vh - 140px);">
          {#if loading}
            <div class="flex items-center gap-2 text-muted-foreground text-[13px] mb-4">
              <div class="size-3 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
              Updating…
            </div>
          {/if}

          <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
            {#each sortedPropKeys(data.properties) as key}
              {@const buckets = data.properties[key]}
              {@const slices = arcData(buckets)}
              <div class="bg-card border border-border rounded-xl p-5">
                <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-4">{propLabel(key)}</div>
                <div class="flex gap-5 items-center">
                  <!-- Donut -->
                  <div class="size-[160px] flex-shrink-0">
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
                  <!-- Legend — each row is clickable for drill-down -->
                  <div class="flex flex-col gap-1 min-w-0 flex-1">
                    {#each slices as slice, i}
                      {@const isHighlighted = selectedApp
                        ? (data.apps.find(a => a.id === selectedApp)?.properties?.[key] ?? '') === (slice.key === '(unset)' ? '' : slice.key)
                        : false}
                      <button
                        class="flex items-center gap-2 text-[12px] w-full text-left rounded px-1 py-0.5 transition-colors hover:bg-muted/50 {isHighlighted ? 'ring-1 ring-primary/50 bg-primary/5' : ''}"
                        onclick={() => openDrill(key, slice.key)}
                        title="Click to see apps"
                      >
                        <span class="size-2.5 rounded-full flex-shrink-0" style="background:{slice.color}"></span>
                        <span class="text-muted-foreground truncate flex-1" title={slice.label}>{slice.label}</span>
                        <span class="font-semibold tabular-nums text-foreground">{slice.value}</span>
                      </button>
                    {/each}
                  </div>
                </div>
              </div>
            {/each}
          </div>
        </div>

      </div>
    {/if}
  {/if}
</div>
