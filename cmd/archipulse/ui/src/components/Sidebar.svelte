<script>
  import { createEventDispatcher } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { VIEWS, LAYER_GROUPS } from '../lib/views.js';
  import { api } from '../lib/api.js';

  export let wsId;
  export let ws = null;
  export let activeView = null;

  const dispatch = createEventDispatcher();

  let importResult = null;
  let importing = false;
  let dropOver = false;

  function navTarget(key, v) {
    return v.graph ? key + '/graph' : v.tree ? key + '/tree' : key;
  }

  function isActive(key, v) {
    const target = navTarget(key, v);
    return activeView === target || activeView === key;
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

<aside class="sidebar">
  {#if ws}
    <div class="sidebar-ws-header">
      <div class="sidebar-ws-name">{ws.name}</div>
      <span class="purpose-badge">{ws.purpose}</span>
    </div>
  {/if}

  <div
    class="sidebar-item {!activeView ? 'active' : ''}"
    style="margin:8px 8px 0"
    on:click={() => push('/ws/' + wsId)}
    on:keydown={e => e.key === 'Enter' && push('/ws/' + wsId)}
    role="button"
    tabindex="0"
  >
    <span class="si-icon">⌂</span> Overview
  </div>
  <div class="divider" style="margin:8px 8px 0"></div>

  {#each LAYER_GROUPS as group}
    {@const items = Object.entries(VIEWS).filter(([, v]) => v.layer === group.key)}
    {#if items.length > 0}
      <div class="sidebar-section">
        <div class="sidebar-section-label">{group.label}</div>
        {#each items as [key, v]}
          <div
            class="sidebar-item {isActive(key, v) ? 'active' : ''}"
            on:click={() => push('/ws/' + wsId + '/view/' + navTarget(key, v))}
            on:keydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/view/' + navTarget(key, v))}
            role="button"
            tabindex="0"
          >
            <span class="sidebar-layer-dot {group.dot}"></span>
            {v.label}
          </div>
        {/each}
      </div>
    {/if}
  {/each}

  <div class="sidebar-footer">
    <div
      class="drop-zone {dropOver ? 'over' : ''}"
      style="padding:14px"
      on:click={() => document.getElementById('sb-file-input-' + wsId).click()}
      on:dragover={handleDragOver}
      on:dragleave={handleDragLeave}
      on:drop={handleDrop}
      role="button"
      tabindex="0"
      on:keydown={e => e.key === 'Enter' && document.getElementById('sb-file-input-' + wsId).click()}
    >
      <div class="dz-icon">↑</div>
      <p>Import model</p>
      <div class="hint">.xml · .ajx · .json</div>
    </div>
    <input
      type="file"
      id="sb-file-input-{wsId}"
      accept=".xml,.ajx,.json"
      style="display:none"
      on:change={handleFileInput}
    />
    {#if importing}
      <div class="loading" style="margin-top:8px"><div class="spinner"></div></div>
    {:else if importResult}
      <div class="alert {importResult.ok ? 'alert-success' : 'alert-error'}" style="margin-top:8px;font-size:12px">
        {importResult.msg}
      </div>
    {/if}
  </div>
</aside>
