<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';

  export let params = {};
  $: wsId = params.wsId;

  let entries = [];
  let propertyKeys = [];
  let loading = true;
  let error = null;

  // ── Search & sort ──────────────────────────────────────────────────────────
  let search = '';
  let sortCol = 'name';   // 'name' | 'type' | any propKey
  let sortDir = 'asc';    // 'asc' | 'desc'

  // ── Column visibility ──────────────────────────────────────────────────────
  // Fixed columns always visible; property columns togglable.
  const FIXED_COLS = ['name', 'type'];
  // These property keys are shown by default when present.
  const DEFAULT_PROP_COLS = ['vendor', 'lifecycle_status', 'criticality', 'deployment_model', 'business_owner', 'user_count'];

  let visibleProps = new Set();
  let colMenuOpen = false;

  const PROP_LABELS = {
    lifecycle_status: 'Lifecycle Status',
    deployment_model: 'Deployment Model',
    criticality:      'Criticality',
    vendor:           'Vendor',
    business_owner:   'Business Owner',
    user_count:       'User Count',
  };
  function propLabel(k) {
    return PROP_LABELS[k] ?? k.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
  }

  const LIFECYCLE_COLORS = {
    'Production':    'bg-[#14301e] text-[#4ade80]',
    'Pilot':         'bg-[#1e2f55] text-[#60a5fa]',
    'Planned':       'bg-[#2a1f4a] text-[#a78bfa]',
    'Retiring':      'bg-[#3a2010] text-[#fb923c]',
    'Decommissioned':'bg-[#3a1010] text-[#f87171]',
  };
  const CRIT_COLORS = {
    'Critical': 'bg-[#3a1010] text-[#f87171]',
    'High':     'bg-[#3a2010] text-[#fb923c]',
    'Medium':   'bg-[#3a3010] text-[#facc15]',
    'Low':      'bg-[#14301e] text-[#4ade80]',
  };
  const DEPLOY_COLORS = {
    'On-Premise':   'bg-[#1a2a1a] text-[#4ade80]',
    'Public Cloud': 'bg-[#1e2f55] text-[#60a5fa]',
    'SaaS':         'bg-[#14302a] text-[#34d399]',
    'Hybrid':       'bg-[#2a1f4a] text-[#a78bfa]',
  };

  function badgeClass(key, val) {
    if (key === 'lifecycle_status') return LIFECYCLE_COLORS[val] ?? 'bg-muted text-muted-foreground';
    if (key === 'criticality')      return CRIT_COLORS[val]      ?? 'bg-muted text-muted-foreground';
    if (key === 'deployment_model') return DEPLOY_COLORS[val]    ?? 'bg-muted text-muted-foreground';
    return null;
  }

  // ── Type display ───────────────────────────────────────────────────────────
  function typeLabel(t) {
    return t.replace(/^Application/, '').replace(/([A-Z])/g, ' $1').trim();
  }

  // ── Filtering & sorting ────────────────────────────────────────────────────
  $: activePropCols = propertyKeys.filter(k => visibleProps.has(k));

  $: filtered = (() => {
    if (!search) return entries;
    const q = search.toLowerCase();
    return entries.filter(e =>
      e.name.toLowerCase().includes(q) ||
      e.type.toLowerCase().includes(q) ||
      e.documentation.toLowerCase().includes(q) ||
      Object.values(e.properties).some(v => v.toLowerCase().includes(q))
    );
  })();

  $: sorted = (() => {
    const dir = sortDir === 'asc' ? 1 : -1;
    return [...filtered].sort((a, b) => {
      const av = sortCol === 'name' ? a.name : sortCol === 'type' ? a.type : (a.properties[sortCol] ?? '');
      const bv = sortCol === 'name' ? b.name : sortCol === 'type' ? b.type : (b.properties[sortCol] ?? '');
      return av.localeCompare(bv) * dir;
    });
  })();

  function setSort(col) {
    if (sortCol === col) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      sortCol = col;
      sortDir = 'asc';
    }
  }

  function sortIcon(col) {
    if (sortCol !== col) return '↕';
    return sortDir === 'asc' ? '↑' : '↓';
  }

  function toggleProp(k) {
    const next = new Set(visibleProps);
    if (next.has(k)) next.delete(k); else next.add(k);
    visibleProps = next;
  }

  function exportCSV() {
    const cols = ['Name', 'Type', ...activePropCols.map(propLabel), 'Description'];
    const lines = [cols.join(',')];
    sorted.forEach(e => {
      const vals = [
        e.name,
        typeLabel(e.type),
        ...activePropCols.map(k => e.properties[k] ?? ''),
        e.documentation,
      ];
      lines.push(vals.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(','));
    });
    const blob = new Blob([lines.join('\n')], { type: 'text/csv' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'application-catalogue.csv';
    a.click();
  }

  onMount(async () => {
    try {
      const data = await api.get('/workspaces/' + wsId + '/views/application-catalogue/entries');
      entries = data.entries ?? [];
      propertyKeys = data.property_keys ?? [];
      // Default-visible: keys present in data that are in DEFAULT_PROP_COLS, in that order.
      visibleProps = new Set(DEFAULT_PROP_COLS.filter(k => propertyKeys.includes(k)));
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });
</script>

<svelte:window on:click={() => { colMenuOpen = false; }} />

<div class="content">
  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">{error}</div>
  {:else}

    <!-- Header -->
    <div class="flex items-center justify-between gap-4 mb-5 flex-wrap">
      <div>
        <h1 class="text-[18px] font-semibold">Application Catalogue</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">
          Showing {sorted.length} of {entries.length} applications
        </div>
      </div>

      <!-- Toolbar -->
      <div class="flex items-center gap-2 flex-wrap">
        <!-- Search -->
        <input
          type="search"
          bind:value={search}
          placeholder="Search…"
          class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary w-48"
        />

        <!-- Column visibility -->
        <div class="relative" on:click|stopPropagation>
          <button
            class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground hover:bg-muted transition-colors"
            on:click={() => colMenuOpen = !colMenuOpen}
          >
            Columns ▾
          </button>
          {#if colMenuOpen}
            <div class="absolute right-0 top-full mt-1 z-30 bg-popover border border-border rounded-lg shadow-lg p-2 w-52">
              <div class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wide px-2 mb-1.5">Property Columns</div>
              {#each propertyKeys as k}
                <label class="flex items-center gap-2 px-2 py-1 rounded hover:bg-muted cursor-pointer">
                  <input type="checkbox" checked={visibleProps.has(k)} on:change={() => toggleProp(k)} class="accent-primary" />
                  <span class="text-[12px] text-foreground">{propLabel(k)}</span>
                </label>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Export -->
        <button
          class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground hover:bg-muted transition-colors"
          on:click={exportCSV}
        >↓ CSV</button>
      </div>
    </div>

    {#if entries.length === 0}
      <div class="text-center py-16 text-muted-foreground">
        <div class="text-[40px] mb-3">📭</div>
        <p class="text-[14px]">No applications found. Import a model first.</p>
      </div>
    {:else}
      <div class="overflow-x-auto border border-border rounded-lg">
        <table class="w-full text-[13px]">
          <thead>
            <tr class="border-b border-border bg-muted/60">
              <!-- Name -->
              <th
                class="text-left px-3 py-2.5 font-semibold text-muted-foreground whitespace-nowrap cursor-pointer hover:text-foreground select-none"
                on:click={() => setSort('name')}
              >Name <span class="text-[11px] opacity-60">{sortIcon('name')}</span></th>

              <!-- Type -->
              <th
                class="text-left px-3 py-2.5 font-semibold text-muted-foreground whitespace-nowrap cursor-pointer hover:text-foreground select-none"
                on:click={() => setSort('type')}
              >Type <span class="text-[11px] opacity-60">{sortIcon('type')}</span></th>

              <!-- Active property columns -->
              {#each activePropCols as k}
                <th
                  class="text-left px-3 py-2.5 font-semibold text-muted-foreground whitespace-nowrap cursor-pointer hover:text-foreground select-none"
                  on:click={() => setSort(k)}
                >{propLabel(k)} <span class="text-[11px] opacity-60">{sortIcon(k)}</span></th>
              {/each}

              <!-- Description (not sortable, last) -->
              <th class="text-left px-3 py-2.5 font-semibold text-muted-foreground">Description</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border/50">
            {#each sorted as entry}
              <tr class="hover:bg-muted/30 transition-colors">
                <td class="px-3 py-2.5 font-medium text-foreground whitespace-nowrap">{entry.name}</td>
                <td class="px-3 py-2.5 text-muted-foreground whitespace-nowrap">{typeLabel(entry.type)}</td>

                {#each activePropCols as k}
                  {@const val = entry.properties[k] ?? ''}
                  {@const bc = val ? badgeClass(k, val) : null}
                  <td class="px-3 py-2.5 whitespace-nowrap">
                    {#if bc}
                      <span class="inline-block px-2 py-0.5 rounded text-[11px] font-medium {bc}">{val}</span>
                    {:else if val}
                      <span class="text-foreground">{val}</span>
                    {:else}
                      <span class="text-muted-foreground">—</span>
                    {/if}
                  </td>
                {/each}

                <td class="px-3 py-2.5 text-muted-foreground max-w-xs truncate" title={entry.documentation}>
                  {entry.documentation || '—'}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <div class="mt-3 text-[12px] text-muted-foreground">
        Showing {sorted.length} of {entries.length} entries
      </div>
    {/if}
  {/if}
</div>
