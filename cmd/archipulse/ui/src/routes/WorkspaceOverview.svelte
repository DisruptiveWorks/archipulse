<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { api } from '../lib/api.js';
  import { VIEWS } from '../lib/views.js';
  import { importRevision } from '../lib/workspace-events.js';

  export let params = {};

  let ws = null;
  let loading = true;
  let statsLoading = true;
  let error = null;
  let elements = 0, biz = 0, app = 0, tech = 0, mot = 0;

  $: wsId = params.wsId;

  onMount(async () => {
    await load();
  });

  // Reload when a new model is imported (even if we're already on this route).
  $: if ($importRevision >= 0 && wsId) load();

  async function load() {
    loading = true;
    error = null;
    try {
      ws = await api.get('/workspaces/' + wsId);
    } catch (e) {
      error = e.message;
      loading = false;
      return;
    }
    loading = false;

    // Load stats
    statsLoading = true;
    try {
      const cat = await api.get('/workspaces/' + wsId + '/views/element-catalogue');
      const rows = cat.rows || [];
      elements = rows.length;
      rows.forEach(r => {
        const layer = r[0];
        if (layer === 'Business') biz++;
        else if (layer === 'Application') app++;
        else if (layer === 'Technology') tech++;
        else if (layer === 'Motivation') mot++;
      });
    } catch(_) {}
    statsLoading = false;
  }

  function navTarget(key, v) {
    return v.graph ? key + '/graph' : v.tree ? key + '/tree' : key;
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
  {:else if ws}
    <div class="flex items-start justify-between mb-6 gap-4">
      <div>
        <h1 class="text-[18px] font-semibold">{ws.name}</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">{ws.description || 'No description'}</div>
      </div>
    </div>

    {#if statsLoading}
      <div class="flex items-center gap-2 text-muted-foreground py-6">
        <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
        Loading model stats…
      </div>
    {:else if elements === 0}
      <div class="text-center py-16 px-6 text-muted-foreground">
        <div class="text-[40px] mb-3.5">📭</div>
        <p class="text-[14px] leading-relaxed">No model imported yet.<br>Use the import panel on the left to upload an AOEF or AJX file.</p>
      </div>
    {:else}
      <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-3">Model overview</div>
      <div class="grid grid-cols-[repeat(auto-fill,minmax(180px,1fr))] gap-3 mb-7">
        <div class="bg-card border border-border rounded-lg p-4">
          <div class="text-[11px] text-muted-foreground uppercase tracking-wide mb-1.5">Total elements</div>
          <div class="text-2xl font-bold">{elements}</div>
        </div>
        <div class="bg-card border border-border rounded-lg p-4">
          <div class="text-[11px] text-muted-foreground uppercase tracking-wide mb-1.5">Business</div>
          <div class="text-2xl font-bold" style="color:#e0af68">{biz}</div>
        </div>
        <div class="bg-card border border-border rounded-lg p-4">
          <div class="text-[11px] text-muted-foreground uppercase tracking-wide mb-1.5">Application</div>
          <div class="text-2xl font-bold" style="color:#7aa2f7">{app}</div>
        </div>
        <div class="bg-card border border-border rounded-lg p-4">
          <div class="text-[11px] text-muted-foreground uppercase tracking-wide mb-1.5">Technology</div>
          <div class="text-2xl font-bold" style="color:#9ece6a">{tech}</div>
        </div>
        {#if mot > 0}
          <div class="bg-card border border-border rounded-lg p-4">
            <div class="text-[11px] text-muted-foreground uppercase tracking-wide mb-1.5">Motivation</div>
            <div class="text-2xl font-bold" style="color:#bb9af7">{mot}</div>
          </div>
        {/if}
      </div>

      <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-3 mt-2">Available views</div>
      <div class="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-2.5">
        {#each Object.entries(VIEWS) as [key, v]}
          {@const target = navTarget(key, v)}
          <div
            class="bg-card border border-border rounded-lg px-4 py-3.5 cursor-pointer transition-colors hover:border-primary"
            onclick={() => push('/ws/' + wsId + '/view/' + target)}
            role="button"
            tabindex="0"
            onkeydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/view/' + target)}
          >
            <div class="text-[13px] font-semibold mb-0.5">{v.label}</div>
            <div class="text-[12px] text-muted-foreground">{v.desc}</div>
          </div>
        {/each}
      </div>

      <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-3 mt-6">Model diagrams</div>
      <div
        class="bg-card border border-border rounded-lg px-4 py-3.5 cursor-pointer transition-colors hover:border-primary inline-flex items-center gap-2"
        onclick={() => push('/ws/' + wsId + '/diagrams')}
        role="button"
        tabindex="0"
        onkeydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/diagrams')}
      >
        <span class="text-[13px] font-semibold">Browse diagrams</span>
        <span class="text-[12px] text-muted-foreground">— original ArchiMate views from the imported file</span>
      </div>
    {/if}
  {/if}
</div>
