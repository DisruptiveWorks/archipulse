<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { VIEWS } from '../lib/views.js';
  import TableView from './TableView.svelte';
  import ApplicationDashboard from './ApplicationDashboard.svelte';

  export let params = {};

  $: wsId = params.wsId;
  $: viewName = params.viewName;
  $: view = VIEWS[viewName];

  let redirected = false;

  onMount(() => {
    if (view && view.graph) {
      push('/ws/' + wsId + '/view/' + viewName + '/graph');
      redirected = true;
    } else if (view && view.tree) {
      push('/ws/' + wsId + '/view/' + viewName + '/tree');
      redirected = true;
    }
  });
</script>

{#if view?.dashboard}
  <ApplicationDashboard {params} />
{:else if !redirected && (!view || (!view.graph && !view.tree))}
  <TableView {params} />
{:else}
  <div class="flex items-center gap-2 text-muted-foreground py-6">
    <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
    Redirecting…
  </div>
{/if}
