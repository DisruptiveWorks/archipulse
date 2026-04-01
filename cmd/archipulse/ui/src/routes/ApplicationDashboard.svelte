<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { PieChart, BarChart } from 'layerchart';

  export let params = {};

  $: wsId = params.wsId;

  let data = null;
  let loading = true;
  let error = null;

  // Lifecycle status color mapping
  const STATUS_COLORS = {
    'Production':    '#4ade80',  // green
    'Pilot':         '#60a5fa',  // blue
    'Planned':       '#a78bfa',  // violet
    'Retiring':      '#fb923c',  // orange
    'Decommissioned':'#f87171',  // red
    '(unset)':       '#6b7280',  // gray
  };

  // ApplicationComponent/Service/etc → short label
  function shortType(t) {
    return t.replace('Application', '').replace(/([A-Z])/g, ' $1').trim();
  }

  $: lifecycleData = (data?.lifecycle ?? []).map(d => ({
    key:   d.status,
    label: d.status,
    value: d.count,
    color: STATUS_COLORS[d.status] ?? '#8b8fa8',
  }));

  $: typeData = (data?.by_type ?? []).map(d => ({
    key:   d.type,
    label: shortType(d.type),
    value: d.count,
  }));

  onMount(async () => {
    loading = true;
    error = null;
    try {
      data = await api.get('/workspaces/' + wsId + '/views/application-dashboard/stats');
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
    <div class="mt-6 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">Error: {error}</div>
  {:else if data}
    <div class="mb-6">
      <h1 class="text-[18px] font-semibold">Application Dashboard</h1>
      <div class="text-muted-foreground text-[13px] mt-0.5">{data.total_apps} applications · lifecycle &amp; type breakdown</div>
    </div>

    {#if data.total_apps === 0}
      <div class="text-center py-16 px-6 text-muted-foreground">
        <div class="text-[40px] mb-3.5">📭</div>
        <p class="text-[14px] leading-relaxed">No application elements found.<br>Import a model first.</p>
      </div>
    {:else}
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

        <!-- Lifecycle donut -->
        <div class="bg-card border border-border rounded-xl p-5">
          <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-4">Lifecycle Status</div>
          <div class="h-[260px]">
            <PieChart
              data={lifecycleData}
              key="key"
              label="label"
              value="value"
              c="color"
              innerRadius={0.65}
              padAngle={0.025}
              cornerRadius={3}
              tooltip={{ offset: 12 }}
            />
          </div>
          <div class="mt-4 flex flex-wrap gap-x-4 gap-y-2">
            {#each lifecycleData as item}
              <div class="flex items-center gap-1.5 text-[12px]">
                <span class="size-2.5 rounded-full flex-shrink-0" style="background:{item.color}"></span>
                <span class="text-muted-foreground">{item.label}</span>
                <span class="font-semibold tabular-nums">{item.value}</span>
              </div>
            {/each}
          </div>
        </div>

        <!-- By type bar chart -->
        <div class="bg-card border border-border rounded-xl p-5">
          <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-4">By Element Type</div>
          <div class="h-[260px]">
            <BarChart
              data={typeData}
              x="label"
              y="value"
              orientation="horizontal"
              bandPadding={0.35}
              series={[{
                key: 'default',
                value: d => d.value,
                color: 'var(--color-primary, #7aa2f7)',
              }]}
              axis={{ x: { format: d => String(d) } }}
            />
          </div>
        </div>

      </div>
    {/if}
  {/if}
</div>
