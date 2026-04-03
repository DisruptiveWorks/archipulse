<script>
  import { push } from 'svelte-spa-router';
  import { VIEWS } from '../lib/views.js';
  import TableView from '../components/views/TableView.svelte';
  import ApplicationDashboard from '../components/views/ApplicationDashboard.svelte';
  import ApplicationLandscapeMap from './ApplicationLandscapeMap.svelte';
  import ApplicationCatalogueView from '../components/views/ApplicationCatalogueView.svelte';
  import TechnologyCatalogueView from '../components/views/TechnologyCatalogueView.svelte';
  import ElementCatalogueView from '../components/views/ElementCatalogueView.svelte';

  export let params = {};

  $: wsId = params.wsId;
  $: viewName = params.viewName;
  $: view = VIEWS[viewName];

  let redirected = false;

  // Use a reactive block instead of onMount so re-navigation between views
  // (same ViewRouter component, different params) also triggers the redirect.
  $: if (viewName) {
    if (view?.graph) {
      push('/ws/' + wsId + '/view/' + viewName + '/graph');
      redirected = true;
    } else if (view?.tree) {
      push('/ws/' + wsId + '/view/' + viewName + '/tree');
      redirected = true;
    } else if (view?.map) {
      push('/ws/' + wsId + '/view/' + viewName + '/map');
      redirected = true;
    } else {
      redirected = false;
    }
  }
</script>

{#if view?.dashboard}
  <ApplicationDashboard {params} />
{:else if view?.catalogue === 'element'}
  <ElementCatalogueView {params} />
{:else if view?.catalogue === 'application'}
  <ApplicationCatalogueView {params} />
{:else if view?.catalogue === 'technology'}
  <TechnologyCatalogueView {params} />
{:else if redirected}
  <div class="flex items-center gap-2 text-muted-foreground py-6">
    <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
    Redirecting…
  </div>
{:else if view}
  <TableView {params} />
{/if}
