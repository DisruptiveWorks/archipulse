<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { VIEWS } from '../lib/views.js';
  import * as Table from '$lib/components/ui/table';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';

  export let params = {};

  let rows = [];
  let columns = [];
  let loading = true;
  let error = null;

  $: wsId = params.wsId;
  $: viewName = params.viewName;
  $: meta = VIEWS[viewName] || { label: viewName, desc: '' };

  const layerCols = new Set(['Layer', 'layer']);

  function layerTagClass(val) {
    if (val === 'Application') return 'bg-[#1e2f55] text-[#7aa2f7] border-0 text-[11px]';
    if (val === 'Business')    return 'bg-[#2a2414] text-[#e0af68] border-0 text-[11px]';
    if (val === 'Technology')  return 'bg-[#1a2a1a] text-[#9ece6a] border-0 text-[11px]';
    if (val === 'Motivation')  return 'bg-[#2a1a2a] text-[#bb9af7] border-0 text-[11px]';
    return null;
  }

  onMount(async () => {
    await load();
  });

  async function load() {
    loading = true;
    error = null;
    try {
      const data = await api.get('/workspaces/' + wsId + '/views/' + viewName);
      rows = data.rows || [];
      columns = data.columns || [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function exportCSV() {
    const lines = [columns.join(',')];
    rows.forEach(r => lines.push(r.map(c => '"' + String(c ?? '').replace(/"/g, '""') + '"').join(',')));
    const blob = new Blob([lines.join('\n')], { type: 'text/csv' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = viewName + '.csv';
    a.click();
  }
</script>

<div class="content">
  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="mt-6 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">Error: {error}</div>
  {:else}
    <div class="flex items-start justify-between mb-6 gap-4">
      <div>
        <h1 class="text-[18px] font-semibold">{meta.label}</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">{rows.length} rows · {meta.desc}</div>
      </div>
      <Button variant="outline" size="sm" onclick={exportCSV}>↓ Export CSV</Button>
    </div>

    {#if rows.length === 0}
      <div class="text-center py-16 px-6 text-muted-foreground">
        <div class="text-[40px] mb-3.5">📭</div>
        <p class="text-[14px] leading-relaxed">No data — import a model first.</p>
      </div>
    {:else}
      <div class="overflow-x-auto border border-border rounded-lg">
        <Table.Root>
          <Table.Header>
            <Table.Row class="border-border hover:bg-transparent">
              {#each columns as col}
                <Table.Head class="bg-muted text-muted-foreground font-semibold whitespace-nowrap">{col}</Table.Head>
              {/each}
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each rows as row}
              <Table.Row class="border-border hover:bg-muted/50">
                {#each row as cell, i}
                  {@const col = columns[i]}
                  {@const val = cell == null ? '' : String(cell)}
                  {@const tagClass = layerCols.has(col) ? layerTagClass(val) : null}
                  <Table.Cell class="text-foreground">
                    {#if tagClass}
                      <Badge class={tagClass}>{val}</Badge>
                    {:else if val === ''}
                      <span class="text-muted-foreground">—</span>
                    {:else}
                      {val}
                    {/if}
                  </Table.Cell>
                {/each}
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>
      </div>
    {/if}
  {/if}
</div>
