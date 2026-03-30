<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { api } from '../lib/api.js';
  import { VIEWS } from '../lib/views.js';

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
    <div class="loading"><div class="spinner"></div> Loading…</div>
  {:else if error}
    <div class="alert alert-error" style="margin-top:24px">Error: {error}</div>
  {:else if ws}
    <div class="page-header">
      <div>
        <h1>{ws.name}</h1>
        <div class="sub">{ws.description || 'No description'}</div>
      </div>
    </div>

    {#if statsLoading}
      <div class="loading"><div class="spinner"></div> Loading model stats…</div>
    {:else if elements === 0}
      <div class="empty-state">
        <div class="es-icon">📭</div>
        <p>No model imported yet.<br>Use the import panel on the left to upload an AOEF or AJX file.</p>
      </div>
    {:else}
      <div class="section-label">Model overview</div>
      <div class="overview-grid">
        <div class="overview-card">
          <div class="ov-label">Total elements</div>
          <div class="ov-value">{elements}</div>
        </div>
        <div class="overview-card">
          <div class="ov-label">Business</div>
          <div class="ov-value" style="color:#e0af68">{biz}</div>
        </div>
        <div class="overview-card">
          <div class="ov-label">Application</div>
          <div class="ov-value" style="color:#7aa2f7">{app}</div>
        </div>
        <div class="overview-card">
          <div class="ov-label">Technology</div>
          <div class="ov-value" style="color:#9ece6a">{tech}</div>
        </div>
        {#if mot > 0}
          <div class="overview-card">
            <div class="ov-label">Motivation</div>
            <div class="ov-value" style="color:#bb9af7">{mot}</div>
          </div>
        {/if}
      </div>

      <div class="section-label" style="margin-top:8px">Available views</div>
      <div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:10px">
        {#each Object.entries(VIEWS) as [key, v]}
          {@const target = navTarget(key, v)}
          <div
            style="background:var(--surface);border:1px solid var(--border);border-radius:8px;padding:14px 16px;cursor:pointer;transition:border-color .15s"
            on:click={() => push('/ws/' + wsId + '/view/' + target)}
            on:mouseover={e => e.currentTarget.style.borderColor='var(--accent)'}
            on:mouseout={e => e.currentTarget.style.borderColor='var(--border)'}
            on:focus={e => e.currentTarget.style.borderColor='var(--accent)'}
            on:blur={e => e.currentTarget.style.borderColor='var(--border)'}
            role="button"
            tabindex="0"
            on:keydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/view/' + target)}
          >
            <div style="font-size:13px;font-weight:600;margin-bottom:3px">{v.label}</div>
            <div style="font-size:12px;color:var(--muted)">{v.desc}</div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>
