<script>
  import { onMount } from 'svelte';
  import { api } from '../../lib/api.js';

  export let params = {};
  $: wsId = params.wsId;

  let entries = [];
  let loading = true;
  let error = null;

  // ── Search & sort ──────────────────────────────────────────────────────────
  let search = '';
  let sortCol = 'name';   // 'name' | 'type' | 'apps'
  let sortDir = 'asc';

  // ── Type → category label ──────────────────────────────────────────────────
  const TYPE_CATEGORY = {
    'Node':              'Infrastructure Node',
    'Device':            'Device',
    'SystemSoftware':    'System Software',
    'TechnologyService': 'Technology Service',
    'Artifact':          'Artifact',
    'Path':              'Network Path',
    'CommunicationNetwork': 'Network',
  };
  function categoryLabel(t) {
    return TYPE_CATEGORY[t] ?? t.replace(/([A-Z])/g, ' $1').trim();
  }

  const CATEGORY_BADGE = {
    'Node':              { bg: '#dbeafe', text: '#1e40af' },
    'Device':            { bg: '#dbeafe', text: '#1e40af' },
    'SystemSoftware':    { bg: '#dcfce7', text: '#166534' },
    'TechnologyService': { bg: '#ffedd5', text: '#9a3412' },
    'Artifact':          { bg: '#ede9fe', text: '#5b21b6' },
    'CommunicationNetwork': { bg: '#f1f5f9', text: '#475569' },
    'Path':              { bg: '#f1f5f9', text: '#475569' },
  };
  function categoryBadgeStyle(t) {
    const b = CATEGORY_BADGE[t] ?? { bg: '#f1f5f9', text: '#475569' };
    return `background:${b.bg}; color:${b.text};`;
  }

  // ── Filtering & sorting ────────────────────────────────────────────────────
  $: filtered = (() => {
    if (!search) return entries;
    const q = search.toLowerCase();
    return entries.filter(e =>
      e.name.toLowerCase().includes(q) ||
      e.type.toLowerCase().includes(q) ||
      e.documentation.toLowerCase().includes(q) ||
      e.used_by_apps.some(a => a.toLowerCase().includes(q))
    );
  })();

  $: sorted = (() => {
    const dir = sortDir === 'asc' ? 1 : -1;
    return [...filtered].sort((a, b) => {
      const av = sortCol === 'name' ? a.name : sortCol === 'type' ? a.type : String(a.used_by_apps.length);
      const bv = sortCol === 'name' ? b.name : sortCol === 'type' ? b.type : String(b.used_by_apps.length);
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

  function exportCSV() {
    const cols = ['Name', 'Category', 'Description', 'Hosted Applications'];
    const lines = [cols.join(',')];
    sorted.forEach(e => {
      const vals = [e.name, categoryLabel(e.type), e.documentation, e.used_by_apps.join('; ')];
      lines.push(vals.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(','));
    });
    const blob = new Blob([lines.join('\n')], { type: 'text/csv' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'technology-catalogue.csv';
    a.click();
  }

  onMount(async () => {
    try {
      const data = await api.get('/workspaces/' + wsId + '/views/technology-catalogue/entries');
      entries = data.entries ?? [];
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
  {:else}

    <!-- Header -->
    <div class="flex items-center justify-between gap-4 mb-5 flex-wrap">
      <div>
        <h1 class="text-[18px] font-semibold">Technology Catalogue</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">
          Showing {sorted.length} of {entries.length} technology elements
        </div>
      </div>

      <div class="flex items-center gap-2">
        <input
          type="search"
          bind:value={search}
          placeholder="Search…"
          class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary w-48"
        />
        <button
          class="bg-card border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground hover:bg-muted transition-colors"
          on:click={exportCSV}
        >↓ CSV</button>
      </div>
    </div>

    {#if entries.length === 0}
      <div class="text-center py-16 text-muted-foreground">
        <div class="text-[40px] mb-3">📭</div>
        <p class="text-[14px]">No technology elements found. Import a model first.</p>
      </div>
    {:else}
      <div class="overflow-x-auto border border-border rounded-lg">
        <table class="w-full text-[13px]">
          <thead>
            <tr class="border-b border-border bg-muted/60">
              <th
                class="text-left px-3 py-2.5 font-semibold text-muted-foreground whitespace-nowrap cursor-pointer hover:text-foreground select-none"
                on:click={() => setSort('name')}
              >Name <span class="text-[11px] opacity-60">{sortIcon('name')}</span></th>

              <th
                class="text-left px-3 py-2.5 font-semibold text-muted-foreground whitespace-nowrap cursor-pointer hover:text-foreground select-none"
                on:click={() => setSort('type')}
              >Category <span class="text-[11px] opacity-60">{sortIcon('type')}</span></th>

              <th class="text-left px-3 py-2.5 font-semibold text-muted-foreground">Description</th>

              <th
                class="text-left px-3 py-2.5 font-semibold text-muted-foreground whitespace-nowrap cursor-pointer hover:text-foreground select-none"
                on:click={() => setSort('apps')}
              >Hosted Applications <span class="text-[11px] opacity-60">{sortIcon('apps')}</span></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border/50">
            {#each sorted as entry}
              <tr class="hover:bg-muted/30 transition-colors">
                <td class="px-3 py-2.5 font-medium text-foreground whitespace-nowrap">{entry.name}</td>
                <td class="px-3 py-2.5 whitespace-nowrap">
                  <span class="inline-block px-2 py-0.5 rounded text-[11px] font-medium" style="{categoryBadgeStyle(entry.type)}">
                    {categoryLabel(entry.type)}
                  </span>
                </td>
                <td class="px-3 py-2.5 text-muted-foreground max-w-xs truncate" title={entry.documentation}>
                  {entry.documentation || '—'}
                </td>
                <td class="px-3 py-2.5">
                  {#if entry.used_by_apps.length === 0}
                    <span class="text-muted-foreground">—</span>
                  {:else}
                    <div class="flex flex-wrap gap-1">
                      {#each entry.used_by_apps as app}
                        <span class="inline-block px-1.5 py-0.5 rounded text-[11px]" style="background:#dbeafe; color:#1e40af;">{app}</span>
                      {/each}
                    </div>
                  {/if}
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
