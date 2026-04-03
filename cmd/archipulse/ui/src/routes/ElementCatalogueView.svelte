<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';

  export let params = {};
  $: wsId = params.wsId;

  let rows    = [];
  let loading = true;
  let error   = null;

  // ── Search & filters ──────────────────────────────────────────────────────
  let search        = '';
  let activeLayers  = new Set(); // empty = show all

  // ── Sort ──────────────────────────────────────────────────────────────────
  let sortCol = 'name'; // 'layer' | 'type' | 'name'
  let sortDir = 'asc';

  // ── Layer meta ────────────────────────────────────────────────────────────
  const LAYER_META = {
    'Application':           { bg: '#0d1b38', text: '#7aa2f7', border: '#2a4080' },
    'Business':              { bg: '#211800', text: '#e0af68', border: '#5a4010' },
    'Technology':            { bg: '#0d1f0d', text: '#9ece6a', border: '#2a4a1a' },
    'Motivation':            { bg: '#1e1030', text: '#bb9af7', border: '#4a2a80' },
    'Strategy':              { bg: '#0d2020', text: '#4fd1c5', border: '#1a5555' },
    'Physical':              { bg: '#201408', text: '#d4956a', border: '#5a3a18' },
    'ImplementationMigration': { bg: '#10181e', text: '#7dcfff', border: '#1e4060' },
    'Composite':             { bg: '#161b22', text: '#8b949e', border: '#30363d' },
  };

  const LAYER_LABELS = {
    'ImplementationMigration': 'Impl. & Migration',
  };

  function layerLabel(l) { return LAYER_LABELS[l] ?? l; }

  // ── Derived data ──────────────────────────────────────────────────────────
  $: layers = [...new Set(rows.map(r => r.layer))].sort();

  $: filtered = rows.filter(r => {
    if (activeLayers.size > 0 && !activeLayers.has(r.layer)) return false;
    if (!search) return true;
    const q = search.toLowerCase();
    return r.name.toLowerCase().includes(q) || r.type.toLowerCase().includes(q);
  });

  $: sorted = [...filtered].sort((a, b) => {
    const av = a[sortCol] ?? '';
    const bv = b[sortCol] ?? '';
    const cmp = av.localeCompare(bv);
    return sortDir === 'asc' ? cmp : -cmp;
  });

  function toggleLayer(l) {
    const next = new Set(activeLayers);
    if (next.has(l)) next.delete(l); else next.add(l);
    activeLayers = next;
  }

  function setSort(col) {
    if (sortCol === col) sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    else { sortCol = col; sortDir = 'asc'; }
  }

  function sortIcon(col) {
    if (sortCol !== col) return '⇅';
    return sortDir === 'asc' ? '↑' : '↓';
  }

  function exportCSV() {
    const lines = [['Layer', 'Type', 'Name', 'Documentation'].join(',')];
    sorted.forEach(r => {
      lines.push([r.layer, r.type, r.name, r.documentation]
        .map(c => '"' + String(c ?? '').replace(/"/g, '""') + '"').join(','));
    });
    const blob = new Blob([lines.join('\n')], { type: 'text/csv' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'element-catalogue.csv';
    a.click();
  }

  onMount(async () => {
    try {
      const data = await api.get('/workspaces/' + wsId + '/views/element-catalogue');
      // data.rows: [[layer, type, name, doc], ...]
      rows = (data.rows ?? []).map(r => ({
        layer:         r[0] || '',
        type:          r[1] || '',
        name:          r[2] || '',
        documentation: r[3] || '',
      }));
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
    <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">Error: {error}</div>
  {:else}

    <!-- Header -->
    <div class="flex items-start justify-between gap-4 mb-5">
      <div>
        <h1 class="text-[18px] font-semibold">Element Catalogue</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">
          {sorted.length} of {rows.length} elements
        </div>
      </div>
      <button
        class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] hover:bg-muted transition-colors flex-shrink-0"
        onclick={exportCSV}
      >↓ Export CSV</button>
    </div>

    <!-- Search + layer filters -->
    <div class="flex flex-wrap items-center gap-2 mb-4">
      <input
        type="search"
        bind:value={search}
        placeholder="Search name or type…"
        class="bg-background border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary w-64"
      />

      <div class="flex flex-wrap gap-1.5">
        {#each layers as l}
          {@const m = LAYER_META[l] ?? LAYER_META['Composite']}
          {@const active = activeLayers.size === 0 || activeLayers.has(l)}
          <button
            onclick={() => toggleLayer(l)}
            class="px-2.5 py-0.5 rounded-full text-[11px] font-medium border transition-opacity {activeLayers.size > 0 && !activeLayers.has(l) ? 'opacity-30' : ''}"
            style="background:{m.bg}; color:{m.text}; border-color:{m.border};"
          >
            {layerLabel(l)}
          </button>
        {/each}
        {#if activeLayers.size > 0}
          <button
            onclick={() => { activeLayers = new Set(); }}
            class="px-2 py-0.5 rounded-full text-[11px] text-muted-foreground hover:text-foreground border border-border transition-colors"
          >✕ clear</button>
        {/if}
      </div>
    </div>

    {#if sorted.length === 0}
      <div class="text-center py-16 text-muted-foreground">
        <div class="text-[36px] mb-3">📭</div>
        <p class="text-[14px]">{rows.length === 0 ? 'No elements — import a model first.' : 'No results match your filters.'}</p>
      </div>
    {:else}
      <div class="overflow-x-auto border border-border rounded-lg">
        <table class="w-full text-[13px]">
          <thead>
            <tr class="border-b border-border bg-muted">
              <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold whitespace-nowrap">
                <button class="flex items-center gap-1 hover:text-foreground transition-colors" onclick={() => setSort('layer')}>
                  Layer <span class="text-[10px]">{sortIcon('layer')}</span>
                </button>
              </th>
              <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold whitespace-nowrap">
                <button class="flex items-center gap-1 hover:text-foreground transition-colors" onclick={() => setSort('type')}>
                  Type <span class="text-[10px]">{sortIcon('type')}</span>
                </button>
              </th>
              <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold whitespace-nowrap">
                <button class="flex items-center gap-1 hover:text-foreground transition-colors" onclick={() => setSort('name')}>
                  Name <span class="text-[10px]">{sortIcon('name')}</span>
                </button>
              </th>
              <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold">Documentation</th>
            </tr>
          </thead>
          <tbody>
            {#each sorted as row}
              {@const m = LAYER_META[row.layer] ?? LAYER_META['Composite']}
              <tr class="border-b border-border hover:bg-muted/40 transition-colors">
                <td class="px-3 py-2 whitespace-nowrap">
                  <span class="inline-block px-2 py-0.5 rounded-full text-[11px] font-medium border"
                        style="background:{m.bg}; color:{m.text}; border-color:{m.border};">
                    {layerLabel(row.layer)}
                  </span>
                </td>
                <td class="px-3 py-2 text-muted-foreground whitespace-nowrap">{row.type || '—'}</td>
                <td class="px-3 py-2 font-medium text-foreground">{row.name}</td>
                <td class="px-3 py-2 text-muted-foreground max-w-xs truncate" title={row.documentation}>
                  {row.documentation || '—'}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {/if}
</div>
