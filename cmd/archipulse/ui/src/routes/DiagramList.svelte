<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { api } from '../lib/api.js';
  import BackButton from '../components/BackButton.svelte';

  export let params = {};
  $: wsId = params.wsId;

  let diagrams = [];
  let loading = true;
  let error = null;

  onMount(async () => {
    try {
      const result = await api.get('/workspaces/' + wsId + '/diagrams');
      diagrams = result || [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });
</script>

<div class="content">
  <BackButton onclick={() => push('/ws/' + wsId)} label="Overview" />
  <h1 class="text-[18px] font-semibold mb-1">Diagrams</h1>
  <p class="text-[13px] text-muted-foreground mb-5">ArchiMate views imported from the model file</p>

  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">{error}</div>
  {:else if diagrams.length === 0}
    <div class="text-center py-16 text-muted-foreground">
      <div class="text-[40px] mb-3">🗂️</div>
      <p class="text-[14px]">No diagrams found.<br>Import a model with views to see them here.</p>
    </div>
  {:else}
    <div class="border border-border rounded-lg overflow-hidden">
      <table class="w-full text-[13px]">
        <thead>
          <tr class="border-b border-border bg-muted/40">
            <th class="text-left px-4 py-2.5 font-medium text-muted-foreground">Name</th>
            <th class="text-left px-4 py-2.5 font-medium text-muted-foreground w-40">Source ID</th>
          </tr>
        </thead>
        <tbody>
          {#each diagrams as d}
            <tr
              class="border-b border-border last:border-0 hover:bg-muted/30 cursor-pointer transition-colors"
              onclick={() => push('/ws/' + wsId + '/diagrams/' + d.id)}
              role="button"
              tabindex="0"
              onkeydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/diagrams/' + d.id)}
            >
              <td class="px-4 py-2.5 font-medium">{d.name || '(unnamed)'}</td>
              <td class="px-4 py-2.5 text-muted-foreground font-mono text-[11px] truncate max-w-[160px]">{d.source_id}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
