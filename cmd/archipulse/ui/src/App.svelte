<script>
  import Router, { location, querystring } from 'svelte-spa-router';
  import { push } from 'svelte-spa-router';
  import Nav from './components/Nav.svelte';
  import Sidebar from './components/Sidebar.svelte';
  import Home from './routes/Home.svelte';
  import WorkspaceOverview from './routes/WorkspaceOverview.svelte';
  import ViewRouter from './routes/ViewRouter.svelte';
  import GraphView from './routes/GraphView.svelte';
  import CapabilityTree from './routes/CapabilityTree.svelte';
  import ApplicationLandscapeMap from './routes/ApplicationLandscapeMap.svelte';

  import { api } from './lib/api.js';
  import { VIEWS } from './lib/views.js';
  import { onMount } from 'svelte';

  // Route definitions
  const routes = {
    '/': Home,
    '/ws/:wsId': WorkspaceOverview,
    '/ws/:wsId/view/:viewName': ViewRouter,
    '/ws/:wsId/view/:viewName/graph': GraphView,
    '/ws/:wsId/view/:viewName/tree': CapabilityTree,
    '/ws/:wsId/view/:viewName/map': ApplicationLandscapeMap,
  };

  // Current route state derived from location
  let currentParams = {};
  let ws = null;
  let wsLoaded = false;

  $: {
    const loc = $location;
    currentParams = extractParams(loc);
  }

  $: wsId = currentParams.wsId || null;
  $: viewName = currentParams.viewName || null;
  $: activeView = currentParams.activeView || null;

  $: if (wsId) {
    loadWs(wsId);
  } else {
    ws = null;
    wsLoaded = false;
  }

  async function loadWs(id) {
    try {
      ws = await api.get('/workspaces/' + id);
    } catch (_) {
      ws = null;
    }
    wsLoaded = true;
  }

  function extractParams(loc) {
    // Match /ws/:wsId/view/:viewName/graph or /tree or /map
    let m = loc.match(/^\/ws\/([^/]+)\/view\/([^/]+)\/(graph|tree|map)$/);
    if (m) return { wsId: m[1], viewName: m[2], activeView: m[2] + '/' + m[3] };

    // Match /ws/:wsId/view/:viewName
    m = loc.match(/^\/ws\/([^/]+)\/view\/([^/]+)$/);
    if (m) {
      const vn = m[2];
      const v = VIEWS[vn];
      const target = v && v.graph ? vn + '/graph' : v && v.tree ? vn + '/tree' : v && v.map ? vn + '/map' : vn;
      return { wsId: m[1], viewName: vn, activeView: target };
    }

    // Match /ws/:wsId
    m = loc.match(/^\/ws\/([^/]+)$/);
    if (m) return { wsId: m[1], viewName: null, activeView: null };

    return {};
  }

  $: viewLabel = viewName ? (VIEWS[viewName] ? VIEWS[viewName].label : viewName) : null;

  function handleImported() {
    // Force reload current route by navigating to same path
    if (wsId) {
      loadWs(wsId);
    }
  }

  function routeEvent(e) {
    currentParams = extractParams($location);
  }
</script>

<Nav wsId={wsId} wsName={ws ? ws.name : null} {viewLabel} />

<div class="app-shell">
  {#if wsId}
    <Sidebar
      {wsId}
      {ws}
      on:imported={handleImported}
    />
    <div style="flex:1;display:flex;flex-direction:column;min-width:0">
      <Router {routes} on:routeEvent={routeEvent} />
    </div>
  {:else}
    <Router {routes} />
  {/if}
</div>
