<script>
  import { createEventDispatcher } from 'svelte';
  import { push, location } from 'svelte-spa-router';
  import { VIEWS, LAYER_GROUPS } from '../lib/views.js';
  import { api } from '../lib/api.js';
  import { Badge } from '$lib/components/ui/badge';
  import { Separator } from '$lib/components/ui/separator';

  export let wsId;
  export let ws = null;
  export let open = false;

  const dispatch = createEventDispatcher();

  // $: ensures Svelte tracks $location as a reactive dependency and
  // re-renders whenever the store changes.
  $: loc = $location;

  const dotColors = {
    'dot-biz':   '#d97706',
    'dot-app':   '#2563eb',
    'dot-tech':  '#16a34a',
    'dot-cross': '#64748b',
    'dot-mot':   '#7c3aed',
  };

  let importResult = null;
  let importing = false;
  let dropOver = false;

  function navTarget(key, v) {
    return v.graph ? key + '/graph' : v.tree ? key + '/tree' : key;
  }

  function handleFileInput(e) {
    const file = e.target.files[0];
    if (file) doImport(file);
    e.target.value = '';
  }

  function handleDragOver(e) {
    e.preventDefault();
    dropOver = true;
  }

  function handleDragLeave() {
    dropOver = false;
  }

  function handleDrop(e) {
    e.preventDefault();
    dropOver = false;
    const file = e.dataTransfer.files[0];
    if (file) doImport(file);
  }

  async function doImport(file) {
    importing = true;
    importResult = null;
    try {
      const fd = new FormData();
      fd.append('file', file);
      const data = await api.upload('/workspaces/' + wsId + '/import', fd);
      importResult = { ok: true, msg: `✓ ${data.elements} elements · ${data.relationships} relationships` };
      setTimeout(() => {
        dispatch('imported');
      }, 1400);
    } catch (e) {
      importResult = { ok: false, msg: '✗ ' + e.message };
    } finally {
      importing = false;
    }
  }
</script>

<aside class="sidebar {open ? 'open' : ''}">
  {#if ws}
    <div class="px-4 pt-4 pb-3 border-b border-border">
      <div class="text-[13px] font-semibold text-foreground whitespace-nowrap overflow-hidden text-ellipsis mb-1">{ws.name}</div>
      <Badge variant="outline" class="border-primary/30 bg-primary/10 text-primary text-[10px] font-semibold uppercase tracking-wide">{ws.purpose}</Badge>
    </div>
  {/if}

  <div
    class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm cursor-pointer transition-colors mx-2 mt-2 {loc === '/ws/' + wsId ? 'bg-white text-foreground font-medium shadow-sm' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
    on:click={() => push('/ws/' + wsId)}
    on:keydown={e => e.key === 'Enter' && push('/ws/' + wsId)}
    role="button"
    tabindex="0"
  >
    <span class="text-[14px] flex-shrink-0 w-[18px] text-center">⌂</span> Overview
  </div>
  <div class="mx-2 mt-2">
    <Separator />
  </div>

  {#each LAYER_GROUPS as group}
    {@const items = Object.entries(VIEWS).filter(([, v]) => v.layer === group.key)}
    {#if items.length > 0}
      <div class="px-2 pt-3 pb-1">
        <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground px-2 mb-1">{group.label}</div>
        {#each items as [key, v]}
          {@const base = '/ws/' + wsId + '/view/' + key}
          {@const active = loc === base || loc.startsWith(base + '/')}
          <div
            class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm cursor-pointer transition-colors {active ? 'bg-white text-foreground font-medium shadow-sm' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
            on:click={() => push('/ws/' + wsId + '/view/' + navTarget(key, v))}
            on:keydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/view/' + navTarget(key, v))}
            role="button"
            tabindex="0"
          >
            <span class="size-1.5 rounded-full flex-shrink-0" style="background:{dotColors[group.dot] || '#8b8fa8'}"></span>
            {v.label}
          </div>
        {/each}
      </div>
    {/if}
  {/each}

  <div class="mt-auto px-2 py-3 border-t border-border">
    <div
      class="border-2 border-dashed border-border rounded-lg p-3.5 text-center text-muted-foreground cursor-pointer transition-colors {dropOver ? 'border-primary text-foreground' : 'hover:border-primary hover:text-foreground'}"
      on:click={() => document.getElementById('sb-file-input-' + wsId).click()}
      on:dragover={handleDragOver}
      on:dragleave={handleDragLeave}
      on:drop={handleDrop}
      role="button"
      tabindex="0"
      on:keydown={e => e.key === 'Enter' && document.getElementById('sb-file-input-' + wsId).click()}
    >
      <div class="text-2xl mb-1.5">↑</div>
      <p class="text-xs">Import model</p>
      <div class="text-[11px] mt-0.5 opacity-70">.xml · .ajx · .json</div>
    </div>
    <input
      type="file"
      id="sb-file-input-{wsId}"
      accept=".xml,.ajx,.json"
      style="display:none"
      on:change={handleFileInput}
    />
    {#if importing}
      <div class="flex items-center gap-2 text-muted-foreground py-2 mt-2">
        <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      </div>
    {:else if importResult}
      <div class="mt-2 text-[12px] px-3 py-2 rounded-md {importResult.ok ? 'bg-success/10 border border-success/30 text-success' : 'bg-destructive/10 border border-destructive/30 text-destructive'}">
        {importResult.msg}
      </div>
    {/if}
  </div>
</aside>
