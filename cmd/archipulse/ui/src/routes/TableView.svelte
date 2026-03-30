<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { VIEWS } from '../lib/views.js';

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
    if (val === 'Application') return 'tag tag-app';
    if (val === 'Business') return 'tag tag-biz';
    if (val === 'Technology') return 'tag tag-tech';
    if (val === 'Motivation') return 'tag tag-mot';
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
    <div class="loading"><div class="spinner"></div> Loading…</div>
  {:else if error}
    <div class="alert alert-error" style="margin-top:24px">Error: {error}</div>
  {:else}
    <div class="page-header">
      <div>
        <h1>{meta.label}</h1>
        <div class="sub">{rows.length} rows · {meta.desc}</div>
      </div>
      <button class="btn btn-ghost btn-sm" on:click={exportCSV}>↓ Export CSV</button>
    </div>

    {#if rows.length === 0}
      <div class="empty-state">
        <div class="es-icon">📭</div>
        <p>No data — import a model first.</p>
      </div>
    {:else}
      <div class="table-wrap">
        <table>
          <thead>
            <tr>{#each columns as col}<th>{col}</th>{/each}</tr>
          </thead>
          <tbody>
            {#each rows as row}
              <tr>
                {#each row as cell, i}
                  {@const col = columns[i]}
                  {@const val = cell == null ? '' : String(cell)}
                  {@const tagClass = layerCols.has(col) ? layerTagClass(val) : null}
                  <td>
                    {#if tagClass}
                      <span class={tagClass}>{val}</span>
                    {:else}
                      {val}
                    {/if}
                  </td>
                {/each}
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {/if}
</div>
