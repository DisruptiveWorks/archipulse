<script>
  import Router, { location, querystring } from 'svelte-spa-router';
  import { push } from 'svelte-spa-router';
  import Nav from './components/Nav.svelte';
  import Sidebar from './components/Sidebar.svelte';
  import Home from './routes/Home.svelte';
  import WorkspaceOverview from './routes/WorkspaceOverview.svelte';
  import ViewRouter from './components/views/ViewRouter.svelte';
  import DependencyGraphView from './components/views/DependencyGraphView.svelte';
  import CapabilityTree from './components/views/CapabilityTree.svelte';
  import ApplicationLandscapeMap from './components/views/ApplicationLandscapeMap.svelte';
  import Login from './routes/Login.svelte';
  import DiagramList from './routes/DiagramList.svelte';
  import DiagramViewer from './routes/DiagramViewer.svelte';
  import EditorPlaceholder from './routes/EditorPlaceholder.svelte';

  import { api } from './lib/api.js';
  import { VIEWS } from './lib/views.js';
  import { user, fetchMe } from './lib/auth.js';
  import { onMount } from 'svelte';

  // Route definitions
  const routes = {
    '/login': Login,
    '/': Home,
    '/ws/:wsId': WorkspaceOverview,
    '/ws/:wsId/editor': EditorPlaceholder,
    '/ws/:wsId/diagrams': DiagramList,
    '/ws/:wsId/diagrams/:diagId': DiagramViewer,
    '/ws/:wsId/view/:viewName': ViewRouter,
    '/ws/:wsId/view/application-dependency/graph': DependencyGraphView,
    '/ws/:wsId/view/:viewName/tree': CapabilityTree,
    '/ws/:wsId/view/:viewName/map': ApplicationLandscapeMap,
  };

  // Auth state
  let authLoading = true;

  onMount(async () => {
    await fetchMe();
    authLoading = false;
    // If not logged in and not already on login page, redirect.
    if (!$user && $location !== '/login') {
      push('/login');
    }
  });

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
    if (wsId) loadWs(wsId);
  }

  function routeEvent(e) {
    currentParams = extractParams($location);
  }

  let sidebarOpen = false;

  function toggleSidebar() {
    sidebarOpen = !sidebarOpen;
  }

  function closeSidebar() {
    sidebarOpen = false;
  }

  // Close sidebar on navigation
  $: if ($location) sidebarOpen = false;

  // Whether to show the shell (nav + sidebar) — hide on the login page.
  $: isLoginPage = $location === '/login';
</script>

{#if authLoading}
  <!-- Blank while we check the session to avoid flash. -->
  <div></div>
{:else if isLoginPage}
  <Router {routes} on:routeEvent={routeEvent} />
{:else}
  <Nav wsId={wsId} wsName={ws ? ws.name : null} {viewLabel} on:toggleSidebar={toggleSidebar} />

  <div class="app-shell">
    {#if wsId}
      <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
      <div class="sidebar-overlay {sidebarOpen ? 'open' : ''}" onclick={closeSidebar}></div>
      <Sidebar
        {wsId}
        {ws}
        open={sidebarOpen}
        on:imported={handleImported}
      />
      <div style="flex:1;display:flex;flex-direction:column;min-width:0">
        <Router {routes} on:routeEvent={routeEvent} />
      </div>
    {:else}
      <Router {routes} />
    {/if}
  </div>
{/if}
