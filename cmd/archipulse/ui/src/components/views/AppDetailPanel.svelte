<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { api } from '../../lib/api.js';

  export let app = null; // { id, name, type, properties: {} }
  export let wsId = null;

  const dispatch = createEventDispatcher();

  let detail = null;
  let loading = false;

  const PROP_LABELS = {
    lifecycle_status:  'Lifecycle',
    deployment_model:  'Deployment',
    criticality:       'Criticality',
    vendor:            'Vendor',
    business_owner:    'Owner',
    user_count:        'Users',
    domain:            'Domain',
    business_domain:   'Domain',
    tco:               'TCO (yr)',
    go_live:           'Go-live',
    go_live_year:      'Go-live',
  };

  const CRIT_PILL = {
    'Critical': { bg: '#fee2e2', color: '#dc2626', border: '#fca5a5' },
    'High':     { bg: '#fef3c7', color: '#b45309', border: '#fcd34d' },
    'Medium':   { bg: '#e0f2fe', color: '#0369a1', border: '#7dd3fc' },
    'Low':      { bg: '#f0fdf4', color: '#166534', border: '#86efac' },
  };
  const LC_PILL = {
    'Production':     { bg: '#f0fdf4', color: '#166534', border: '#86efac' },
    'Pilot':          { bg: '#f5f3ff', color: '#6d28d9', border: '#c4b5fd' },
    'Planned':        { bg: '#e0f2fe', color: '#0369a1', border: '#7dd3fc' },
    'Retiring':       { bg: '#fef3c7', color: '#b45309', border: '#fcd34d' },
    'Decommissioned': { bg: '#f8fafc', color: '#64748b', border: '#cbd5e1' },
  };

  const TECH_TYPE_LABEL = {
    'Node': 'IaaS', 'Device': 'Hardware', 'SystemSoftware': 'PaaS',
    'Artifact': 'Artifact', 'TechnologyService': 'Service',
  };

  const REL_LABEL = {
    'serving': 'serves', 'flow': 'flow', 'access': 'access',
    'triggering': 'triggers', 'association': 'associates',
  };

  function propLabel(key) {
    return PROP_LABELS[key] ?? key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
  }

  function techTypeLabel(type) {
    return TECH_TYPE_LABEL[type] ?? type.replace(/([A-Z])/g, ' $1').trim();
  }

  function isPillProp(key) {
    return key === 'criticality' || key === 'lifecycle_status';
  }

  function pillStyle(key, value) {
    const p = key === 'criticality' ? CRIT_PILL[value] : LC_PILL[value];
    if (!p) return '';
    return `background:${p.bg}; color:${p.color}; border:1px solid ${p.border}`;
  }

  $: subtypeName = app?.type?.replace('Application', '') ?? '';

  $: plainProps = Object.entries(detail?.properties ?? app?.properties ?? {}).filter(([k, v]) =>
    v && !isPillProp(k)
  );
  $: pillProps = Object.entries(detail?.properties ?? app?.properties ?? {}).filter(([k, v]) =>
    v && isPillProp(k)
  );

  $: if (app && wsId) {
    loadDetail(app.id, wsId);
  } else {
    detail = null;
  }

  async function loadDetail(appId, ws) {
    loading = true;
    detail = null;
    try {
      detail = await api.get('/workspaces/' + ws + '/elements/' + encodeURIComponent(appId) + '/app-detail');
    } catch (_) {
      // silently fall back to basic properties from app prop
    } finally {
      loading = false;
    }
  }

  function onKeydown(e) {
    if (e.key === 'Escape') dispatch('close');
  }
</script>

<svelte:window on:keydown={onKeydown} />

{#if app}
  <!-- Backdrop -->
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-40" onclick={() => dispatch('close')}></div>

  <div
    class="fixed right-4 z-50 w-[300px] overflow-y-auto bg-white border border-border rounded-xl shadow-xl flex flex-col"
    style="top: calc(var(--nav-h, 52px) + 12px); max-height: calc(100vh - var(--nav-h, 52px) - 24px)"
    role="dialog"
    aria-modal="true"
  >
    <!-- Header -->
    <div class="flex items-start gap-2 px-4 py-3 border-b border-border flex-shrink-0">
      <span class="size-2 rounded-full flex-shrink-0 mt-1.5" style="background:#7cc8ec"></span>
      <div class="flex-1 min-w-0">
        <div class="text-[13.5px] font-semibold text-foreground leading-tight">{detail?.name ?? app.name}</div>
        {#if subtypeName}
          <div class="text-[11px] text-muted-foreground mt-0.5">Application {subtypeName}</div>
        {/if}
      </div>
      <button
        onclick={() => dispatch('close')}
        class="flex-shrink-0 text-muted-foreground hover:text-foreground p-1 rounded hover:bg-muted transition-colors text-[16px] leading-none"
        aria-label="Close"
      >✕</button>
    </div>

    {#if loading}
      <div class="flex items-center justify-center py-6">
        <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin"></div>
      </div>
    {:else}
      <!-- Pill properties (criticality, lifecycle) -->
      {#if pillProps.length > 0}
        <div class="px-4 pt-3 pb-1 grid gap-2">
          {#each pillProps as [key, value]}
            <div class="grid grid-cols-[110px_1fr] gap-2 text-[12.5px] items-center">
              <span class="text-muted-foreground">{propLabel(key)}</span>
              <span
                class="inline-flex items-center px-2 py-0.5 rounded-full text-[11.5px] font-medium w-fit"
                style={pillStyle(key, value)}
              >{value}</span>
            </div>
          {/each}
        </div>
      {/if}

      <!-- Plain properties -->
      {#if plainProps.length > 0}
        <div class="px-4 {pillProps.length > 0 ? 'pt-1' : 'pt-3'} pb-3 grid gap-2 border-b border-border">
          {#each plainProps as [key, value]}
            <div class="grid grid-cols-[110px_1fr] gap-2 text-[12.5px] items-center">
              <span class="text-muted-foreground">{propLabel(key)}</span>
              <span class="text-foreground">{value}</span>
            </div>
          {/each}
        </div>
      {:else if pillProps.length > 0}
        <div class="border-b border-border"></div>
      {/if}

      <!-- Runs On -->
      {#if detail?.runs_on?.length > 0}
        <div class="px-4 py-3 border-b border-border">
          <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-2">
            Runs On ({detail.runs_on.length})
          </div>
          <div class="space-y-1.5">
            {#each detail.runs_on as tech}
              <div class="flex items-center justify-between text-[12.5px]">
                <span class="text-foreground">{tech.name}</span>
                <span class="text-muted-foreground text-[11px]">{techTypeLabel(tech.type)}</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Interfaces -->
      {#if detail?.interfaces?.length > 0}
        <div class="px-4 py-3 border-b border-border">
          <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-2">
            Interfaces ({detail.interfaces.length})
          </div>
          <div class="space-y-1.5">
            {#each detail.interfaces as iface}
              <div class="flex items-center justify-between text-[12.5px]">
                <span class="text-foreground">
                  {#if iface.direction === 'out'}→{:else}←{/if}
                  {iface.target_name}
                </span>
                <span class="text-muted-foreground text-[11px]">{REL_LABEL[iface.rel_type] ?? iface.rel_type}</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Used In Processes -->
      {#if detail?.processes?.length > 0}
        <div class="px-4 py-3">
          <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground mb-2">
            Used In Processes ({detail.processes.length})
          </div>
          <div class="space-y-1.5">
            {#each detail.processes as proc}
              <div class="text-[12.5px] text-foreground">{proc.name}</div>
            {/each}
          </div>
        </div>
      {/if}

      {#if pillProps.length === 0 && plainProps.length === 0 && !detail?.runs_on?.length && !detail?.interfaces?.length && !detail?.processes?.length}
        <div class="px-4 py-4 text-[12px] text-muted-foreground italic">No properties recorded.</div>
      {/if}
    {/if}
  </div>
{/if}
