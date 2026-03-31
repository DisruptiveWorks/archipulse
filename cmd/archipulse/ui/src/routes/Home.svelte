<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { api } from '../lib/api.js';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';

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
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="mt-6 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">Error: {error}</div>
  {:else if workspaces.length === 0}
    <div class="mb-7">
      <h1 class="text-[20px] font-semibold">Workspaces</h1>
      <p class="text-muted-foreground mt-1 text-[13px]">Your ArchiMate baselines</p>
    </div>
    <div class="text-center py-16 px-6 text-muted-foreground">
      <div class="text-[40px] mb-3.5">🏛️</div>
      <p class="text-[14px] leading-relaxed">No workspaces yet.<br>Create one and import your first ArchiMate model.</p>
      <br>
      <Button onclick={openModal}>+ New workspace</Button>
    </div>
  {:else}
    <div class="flex items-start justify-between mb-6 gap-4">
      <div>
        <h1 class="text-[18px] font-semibold">Workspaces</h1>
        <div class="text-muted-foreground text-[13px] mt-0.5">{workspaces.length} baseline{workspaces.length !== 1 ? 's' : ''}</div>
      </div>
      <Button size="sm" onclick={openModal}>+ New workspace</Button>
    </div>
    <div class="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-4 mt-6">
      {#each workspaces as ws}
        <div class="bg-card border border-border rounded-lg p-5 cursor-pointer transition-all hover:border-primary hover:-translate-y-px flex flex-col gap-2"
          role="button" tabindex="0"
          onclick={() => push('/ws/' + ws.id)}
          onkeydown={e => e.key === 'Enter' && push('/ws/' + ws.id)}>
          <div class="flex items-start justify-between gap-2">
            <h2 class="text-[15px] font-semibold">{ws.name}</h2>
            <Badge variant="outline" class="border-primary/30 bg-primary/10 text-primary text-[10px] font-semibold uppercase tracking-wide shrink-0">{ws.purpose}</Badge>
          </div>
          <div class="text-muted-foreground text-[13px] leading-relaxed flex-1">{ws.description || 'No description'}</div>
          <div class="text-muted-foreground text-[11px] pt-2 border-t border-border">Updated {formatDate(ws.updated_at)}</div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<Dialog.Root bind:open={showModal}>
  <Dialog.Content class="bg-card border-border text-foreground max-w-md">
    <Dialog.Header>
      <Dialog.Title>New workspace</Dialog.Title>
    </Dialog.Header>
    <div class="flex flex-col gap-4 py-2">
      <div class="flex flex-col gap-1.5">
        <Label for="ws-name-input">Name</Label>
        <Input id="ws-name-input" bind:value={wsName} placeholder="Q1-2026-AS-IS" onkeydown={e => e.key === 'Enter' && createWs()} class="bg-background border-border" />
      </div>
      <div class="flex flex-col gap-1.5">
        <Label for="ws-purpose">Purpose</Label>
        <select id="ws-purpose" bind:value={wsPurpose} class="bg-background border border-border rounded-md text-foreground text-sm px-3 py-2 outline-none focus:border-ring">
          <option value="as-is">as-is</option>
          <option value="to-be">to-be</option>
          <option value="migration">migration</option>
        </select>
      </div>
      <div class="flex flex-col gap-1.5">
        <Label for="ws-desc">Description</Label>
        <textarea id="ws-desc" bind:value={wsDesc} placeholder="Optional description" class="bg-background border border-border rounded-md text-foreground text-sm px-3 py-2 outline-none focus:border-ring resize-vertical min-h-[72px]"></textarea>
      </div>
    </div>
    {#if modalError}
      <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">{modalError}</div>
    {/if}
    <Dialog.Footer>
      <Button variant="outline" onclick={closeModal}>Cancel</Button>
      <Button onclick={createWs} disabled={creating}>{creating ? 'Creating…' : 'Create'}</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
