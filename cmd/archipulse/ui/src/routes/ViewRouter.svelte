<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { VIEWS } from '../lib/views.js';
  import TableView from './TableView.svelte';

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

{#if !redirected && (!view || (!view.graph && !view.tree))}
  <TableView {params} />
{:else}
  <div class="loading"><div class="spinner"></div> Redirecting…</div>
{/if}
