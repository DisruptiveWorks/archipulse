<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { api } from '../lib/api.js';

  export let params = {};

  let workspaces = [];
  let loading = true;
  let error = null;

  // Modal state
  let showModal = false;
  let wsName = '';
  let wsPurpose = 'as-is';
  let wsDesc = '';
  let modalError = null;
  let creating = false;

  onMount(async () => {
    await loadWorkspaces();

    // Listen for create-ws event from Nav
    window.addEventListener('archipulse:create-ws', openModal);
    return () => window.removeEventListener('archipulse:create-ws', openModal);
  });

  async function loadWorkspaces() {
    loading = true;
    error = null;
    try {
      workspaces = await api.get('/workspaces');
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function openModal() {
    wsName = '';
    wsPurpose = 'as-is';
    wsDesc = '';
    modalError = null;
    showModal = true;
    setTimeout(() => document.getElementById('ws-name-input')?.focus(), 50);
  }

  function closeModal() {
    showModal = false;
  }

  async function createWs() {
    if (!wsName.trim()) {
      modalError = 'Name is required';
      return;
    }
    creating = true;
    modalError = null;
    try {
      const ws = await api.post('/workspaces', {
        name: wsName.trim(),
        purpose: wsPurpose,
        description: wsDesc.trim(),
      });
      closeModal();
      push('/ws/' + ws.id);
    } catch (e) {
      modalError = e.message;
    } finally {
      creating = false;
    }
  }

  function formatDate(str) {
    return new Date(str).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  }
</script>

<div class="content-full">
  {#if loading}
    <div class="loading"><div class="spinner"></div> Loading…</div>
  {:else if error}
    <div class="alert alert-error" style="margin-top:24px">Error: {error}</div>
  {:else if workspaces.length === 0}
    <div style="margin-bottom:28px">
      <h1 style="font-size:20px;font-weight:600">Workspaces</h1>
      <p style="color:var(--muted);margin-top:4px;font-size:13px">Your ArchiMate baselines</p>
    </div>
    <div class="empty-state">
      <div class="es-icon">🏛️</div>
      <p>No workspaces yet.<br>Create one and import your first ArchiMate model.</p>
      <br>
      <button class="btn btn-primary" on:click={openModal}>+ New workspace</button>
    </div>
  {:else}
    <div class="page-header">
      <div>
        <h1>Workspaces</h1>
        <div class="sub">{workspaces.length} baseline{workspaces.length !== 1 ? 's' : ''}</div>
      </div>
      <button class="btn btn-primary btn-sm" on:click={openModal}>+ New workspace</button>
    </div>
    <div class="ws-grid">
      {#each workspaces as ws}
        <div class="ws-card" on:click={() => push('/ws/' + ws.id)} role="button" tabindex="0" on:keydown={e => e.key === 'Enter' && push('/ws/' + ws.id)}>
          <div class="ws-card-top">
            <h2>{ws.name}</h2>
            <span class="purpose-badge">{ws.purpose}</span>
          </div>
          <div class="desc">{ws.description || 'No description'}</div>
          <div class="meta">Updated {formatDate(ws.updated_at)}</div>
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if showModal}
  <div class="modal-overlay" on:click={e => e.target === e.currentTarget && closeModal()} role="dialog" aria-modal="true">
    <div class="modal">
      <h2>New workspace</h2>
      <div class="form-grid">
        <label>
          Name
          <input id="ws-name-input" bind:value={wsName} placeholder="Q1-2026-AS-IS" on:keydown={e => e.key === 'Enter' && createWs()} />
        </label>
        <label>
          Purpose
          <select bind:value={wsPurpose}>
            <option value="as-is">as-is</option>
            <option value="to-be">to-be</option>
            <option value="migration">migration</option>
          </select>
        </label>
        <label>
          Description
          <textarea bind:value={wsDesc} placeholder="Optional description"></textarea>
        </label>
      </div>
      <div class="modal-actions">
        <button class="btn btn-ghost" on:click={closeModal}>Cancel</button>
        <button class="btn btn-primary" on:click={createWs} disabled={creating}>
          {creating ? 'Creating…' : 'Create'}
        </button>
      </div>
      {#if modalError}
        <div class="alert alert-error" style="margin-top:12px">{modalError}</div>
      {/if}
    </div>
  </div>
{/if}
