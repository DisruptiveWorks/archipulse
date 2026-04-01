<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { ArcChart } from 'layerchart';

  export let params = {};

  $: wsId = params.wsId;

  let data = null;
  let loading = true;
  let error = null;
  let selectedCapability = 'all';

  // Human-readable labels for known property keys
  const PROP_LABELS = {
    lifecycle_status:   'Lifecycle Status',
    deployment_model:   'Deployment Model',
    criticality:        'Business Criticality',
    vendor:             'Vendor',
    business_owner:     'Business Owner',
    user_count:         'User Count Range',
  };

  // Preferred order for property sections
  const PROP_ORDER = [
    'lifecycle_status',
    'deployment_model',
    'criticality',
    'vendor',
    'business_owner',
    'user_count',
  ];

  // Color palette for property values (cycles), unset always last gray
  const PALETTE = [
    '#4ade80', '#60a5fa', '#a78bfa', '#fb923c', '#f87171',
    '#34d399', '#f472b6', '#facc15', '#38bdf8', '#c084fc',
  ];
  const UNSET_COLOR = '#4b5563';

  function colorFor(value, index) {
    if (value === '(unset)') return UNSET_COLOR;
    return PALETTE[index % PALETTE.length];
  }

  function propLabel(key) {
    return PROP_LABELS[key] ?? key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
  }

  // Return property sections in preferred order, unknown keys appended at end
  function sortedProps(properties) {
    if (!properties) return [];
    const known = PROP_ORDER.filter(k => k in properties);
    const unknown = Object.keys(properties).filter(k => !PROP_ORDER.includes(k)).sort();
    return [...known, ...unknown];
  }

  // Build arc data with color pre-assigned
  function arcData(buckets) {
    return buckets.map((b, i) => ({
      key:   b.value,
      label: b.value,
      value: b.count,
      color: colorFor(b.value, i),
    }));
  }

  async function load(capability) {
    loading = true;
    error = null;
    try {
      const qs = capability && capability !== 'all' ? `?capability=${encodeURIComponent(capability)}` : '';
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
    <div class="flex items-start justify-between mb-6 gap-4 flex-wrap">
      <div>
        <h1 class="text-[18px] font-semibold">Application Dashboard</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">
          Applications in scope: <span class="font-semibold text-foreground">{data.total_apps}</span>
        </div>
      </div>

      <!-- Capability filter -->
      {#if data.capabilities?.length > 0}
        <div class="flex items-center gap-2">
          <label for="cap-filter" class="text-[12px] text-muted-foreground whitespace-nowrap">Business Capability</label>
          <select
            id="cap-filter"
            value={selectedCapability}
            onchange={onCapabilityChange}
            class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary min-w-[200px]"
          >
            <option value="all">All</option>
            {#each data.capabilities as cap}
              <option value={cap}>{cap}</option>
            {/each}
          </select>
        </div>
      {/if}
    </div>

    {#if data.total_apps === 0}
      <div class="text-center py-16 px-6 text-muted-foreground">
        <div class="text-[40px] mb-3.5">📭</div>
        <p class="text-[14px] leading-relaxed">No applications found for this scope.<br>Try selecting a different capability or import a model first.</p>
      </div>
    {:else}
      <!-- Loading overlay while refetching -->
      {#if loading}
        <div class="flex items-center gap-2 text-muted-foreground text-[13px] mb-4">
          <div class="size-3 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
          Updating…
        </div>
      {/if}

      <!-- Property donuts grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
        {#each sortedProps(data.properties) as key}
          {@const buckets = data.properties[key]}
          {@const slices = arcData(buckets)}
          <div class="bg-card border border-border rounded-xl p-5">
            <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-4">{propLabel(key)}</div>
            <div class="flex gap-6 items-center">
              <!-- Donut -->
              <div class="size-[180px] flex-shrink-0">
                <ArcChart
                  data={slices}
                  key="key"
                  label="label"
                  value="value"
                  c="color"
                  innerRadius={0.62}
                  padAngle={0.02}
                  cornerRadius={2}
                />
              </div>
              <!-- Legend -->
              <div class="flex flex-col gap-1.5 min-w-0">
                {#each slices as slice}
                  <div class="flex items-center gap-2 text-[12px]">
                    <span class="size-2.5 rounded-full flex-shrink-0" style="background:{slice.color}"></span>
                    <span class="text-muted-foreground truncate" title={slice.label}>{slice.label}</span>
                    <span class="font-semibold tabular-nums ml-auto pl-2">{slice.value}</span>
                  </div>
                {/each}
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>
