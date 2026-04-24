<script>
  import { onMount } from 'svelte';
  import { api } from '../../lib/api.js';
  import ApplicationDashboard from './ApplicationDashboard.svelte';
  import CapabilityLandscape from './CapabilityLandscape.svelte';
  import ApplicationLandscape from './ApplicationLandscape.svelte';
  import CapabilityTree from './CapabilityTree.svelte';
  import DependencyGraphView from './DependencyGraphView.svelte';

  export let params = {};
  $: wsId = params.wsId;
  $: svId = params.svId;

  let sv = null;
  let loading = true;
  let error = null;

  $: if (wsId && svId) load(wsId, svId);

  async function load(ws, id) {
    loading = true;
    error = null;
    sv = null;
    try {
      sv = await api.get('/workspaces/' + ws + '/saved-views/' + id);
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  // Build a params-like object for the child component so it receives wsId.
  $: childParams = { wsId };
</script>

{#if loading}
  <div class="flex items-center gap-2 text-muted-foreground p-6 text-sm">
    <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
    Loading…
  </div>
{:else if error}
  <div class="p-6 text-sm text-destructive">{error}</div>
{:else if sv}
  {#if sv.view_type === 'application-dashboard'}
    <ApplicationDashboard params={childParams} initialFilters={sv.filters} savedViewName={sv.name} />
  {:else if sv.view_type === 'capability-landscape'}
    <CapabilityLandscape params={childParams} initialFilters={sv.filters} savedViewName={sv.name} />
  {:else if sv.view_type === 'application-landscape'}
    <ApplicationLandscape params={childParams} initialFilters={sv.filters} savedViewName={sv.name} />
  {:else if sv.view_type === 'capability-tree'}
    <CapabilityTree params={childParams} initialFilters={sv.filters} savedViewName={sv.name} />
  {:else if sv.view_type === 'application-dependency'}
    <DependencyGraphView params={childParams} initialFilters={sv.filters} savedViewName={sv.name} />
  {:else}
    <div class="p-6 text-sm text-muted-foreground">Unknown view type: {sv.view_type}</div>
  {/if}
{/if}
