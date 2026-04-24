<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { push, location } from 'svelte-spa-router';
  import { VIEWS, LAYER_GROUPS } from '../lib/views.js';
  import { api } from '../lib/api.js';
  import { user } from '../lib/auth.js';
  import { Badge } from '$lib/components/ui/badge';
  import { Separator } from '$lib/components/ui/separator';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';

  const ERROR_MESSAGES = {
    forbidden:            "You don't have permission to perform this action in this workspace.",
    unauthorized:         "Your session has expired. Please log in again.",
    'not authenticated':  "Your session has expired. Please log in again.",
    'internal server error': "An unexpected error occurred. Please try again.",
    'workspace not found': "The workspace could not be found.",
  };

  function friendlyError(raw) {
    const key = (raw || '').toLowerCase().trim();
    return ERROR_MESSAGES[key] ?? raw;
  }

  export let wsId;
  export let ws = null;
  export let open = false;

  $: canSettings = $user?.org_role === 'admin' || myWsRole === 'owner' || myWsRole === 'editor';
  let myWsRole = null;

  $: if (wsId) loadMyRole(wsId);

  async function loadMyRole(id) {
    myWsRole = null;
    try {
      const members = await api.get('/workspaces/' + id + '/members');
      const me = members.find(m => m.user_id === $user?.id);
      myWsRole = me?.role ?? null;
    } catch { /* ignore */ }
  }

  const dispatch = createEventDispatcher();

  // $: ensures Svelte tracks $location as a reactive dependency and
  // re-renders whenever the store changes.
  $: loc = $location;

  const dotColors = {
    'dot-biz':   '#d97706',
    'dot-app':   '#2563eb',
    'dot-tech':  '#16a34a',
    'dot-cross': '#64748b',
    'dot-mot':   '#7c3aed',
  };

  let importing = false;
  let dropOver = false;
  let importError = null;
  let showErrorDialog = false;

  // Preview state
  let pendingFile = null;
  let preview = null;
  let showPreview = false;
  let confirming = false;
  let expandedCategories = {};

  $: showErrorDialog = !!importError;

  function navTarget(key, v) {
    return v.graph ? key + '/graph' : v.tree ? key + '/tree' : v.map ? key + '/map' : key;
  }

  function handleFileInput(e) {
    const file = e.target.files[0];
    if (file) doPreview(file);
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
    if (file) doPreview(file);
  }

  async function doPreview(file) {
    importing = true;
    preview = null;
    pendingFile = file;
    expandedCategories = {};
    try {
      const fd = new FormData();
      fd.append('file', file);
      preview = await api.upload('/workspaces/' + wsId + '/import/preview', fd);
      showPreview = true;
    } catch (e) {
      importError = e.message;
    } finally {
      importing = false;
    }
  }

  async function confirmImport() {
    confirming = true;
    try {
      const fd = new FormData();
      fd.append('file', pendingFile);
      const data = await api.upload('/workspaces/' + wsId + '/import', fd);
      showPreview = false;
      preview = null;
      pendingFile = null;
      dispatch('imported', data);
    } catch (e) {
      showPreview = false;
      importError = e.message;
    } finally {
      confirming = false;
    }
  }

  function cancelPreview() {
    showPreview = false;
    preview = null;
    pendingFile = null;
  }

  function toggleCategory(key) {
    expandedCategories[key] = !expandedCategories[key];
    expandedCategories = expandedCategories;
  }

  function categoryLabel(key) {
    return { elements: 'Elements', relationships: 'Relationships', diagrams: 'Diagrams', property_definitions: 'Property definitions' }[key] || key;
  }
</script>

<aside class="sidebar {open ? 'open' : ''}">
  {#if ws}
    <div class="px-4 pt-4 pb-3 border-b border-border">
      <div class="text-[13px] font-semibold text-foreground whitespace-nowrap overflow-hidden text-ellipsis mb-1">{ws.name}</div>
      <Badge variant="outline" class="border-primary/30 bg-primary/10 text-primary text-[10px] font-semibold uppercase tracking-wide">{ws.purpose}</Badge>
    </div>
  {/if}

  <div
    class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm cursor-pointer transition-colors mx-2 mt-2 {loc === '/ws/' + wsId ? 'bg-white text-foreground font-medium shadow-sm' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
    onclick={() => push('/ws/' + wsId)}
    onkeydown={e => e.key === 'Enter' && push('/ws/' + wsId)}
    role="button"
    tabindex="0"
  >
    <span class="text-[14px] flex-shrink-0 w-[18px] text-center">⌂</span> Overview
  </div>
  <div
    class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm cursor-pointer transition-colors mx-2 mt-0.5 {loc === '/ws/' + wsId + '/history' ? 'bg-white text-foreground font-medium shadow-sm' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
    onclick={() => push('/ws/' + wsId + '/history')}
    onkeydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/history')}
    role="button"
    tabindex="0"
  >
    <span class="text-[14px] flex-shrink-0 w-[18px] text-center">◷</span> History
  </div>
  <div class="mx-2 mt-2">
    <Separator />
  </div>

  {#each LAYER_GROUPS as group}
    {@const items = Object.entries(VIEWS).filter(([, v]) => v.layer === group.key)}
    {#if items.length > 0}
      <div class="px-2 pt-3 pb-1">
        <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground px-2 mb-1">{group.label}</div>
        {#each items as [key, v]}
          {@const base = '/ws/' + wsId + '/view/' + key}
          {@const active = loc === base || loc.startsWith(base + '/')}
          <div
            class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm cursor-pointer transition-colors {active ? 'bg-white text-foreground font-medium shadow-sm' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
            onclick={() => push('/ws/' + wsId + '/view/' + navTarget(key, v))}
            onkeydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/view/' + navTarget(key, v))}
            role="button"
            tabindex="0"
          >
            <span class="size-1.5 rounded-full flex-shrink-0" style="background:{dotColors[group.dot] || '#8b8fa8'}"></span>
            {v.label}
          </div>
        {/each}
      </div>
    {/if}
  {/each}

  <div class="mx-2 mt-3">
    <Separator />
  </div>
  <div class="px-2 pt-2 pb-1">
    <div
      class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm cursor-pointer transition-colors {loc === '/ws/' + wsId + '/saved-views' || loc.startsWith('/ws/' + wsId + '/saved-view/') ? 'bg-white text-foreground font-medium shadow-sm' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
      onclick={() => push('/ws/' + wsId + '/saved-views')}
      onkeydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/saved-views')}
      role="button"
      tabindex="0"
    >
      <span class="text-[12px] flex-shrink-0 w-[18px] text-center">⊛</span>
      Saved Views
    </div>
  </div>

  <div class="mx-2 mt-1">
    <Separator />
  </div>
  <div class="px-2 pt-3 pb-1">
    <div class="text-[10px] font-bold tracking-[0.8px] uppercase text-muted-foreground px-2 mb-1">ArchiMate Editor</div>
    {#each [
      { path: '/ws/' + wsId + '/editor', label: 'Editor', icon: '✎' },
      { path: '/ws/' + wsId + '/diagrams', label: 'Diagram List', icon: '⊟' },
    ] as item}
      {@const active = loc === item.path || loc.startsWith(item.path + '/')}
      <div
        class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm cursor-pointer transition-colors {active ? 'bg-white text-foreground font-medium shadow-sm' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
        onclick={() => push(item.path)}
        onkeydown={e => e.key === 'Enter' && push(item.path)}
        role="button"
        tabindex="0"
      >
        <span class="text-[12px] flex-shrink-0 w-[18px] text-center">{item.icon}</span>
        {item.label}
      </div>
    {/each}
  </div>

  <div class="mx-2 mt-3">
    <Separator />
  </div>
  <div class="px-2 pt-3 pb-1">
    <div
      class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm cursor-pointer transition-colors {loc === '/ws/' + wsId + '/settings' ? 'bg-white text-foreground font-medium shadow-sm' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
      onclick={() => push('/ws/' + wsId + '/settings')}
      onkeydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/settings')}
      role="button"
      tabindex="0"
    >
      <span class="text-[14px] flex-shrink-0 w-[18px] text-center">⚙</span> Settings
    </div>
  </div>

  <div class="mt-auto px-2 py-3 border-t border-border">
    <div
      class="border-2 border-dashed border-border rounded-lg p-3.5 text-center text-muted-foreground cursor-pointer transition-colors {dropOver ? 'border-primary text-foreground' : 'hover:border-primary hover:text-foreground'}"
      onclick={() => document.getElementById('sb-file-input-' + wsId).click()}
      ondragover={handleDragOver}
      ondragleave={handleDragLeave}
      ondrop={handleDrop}
      role="button"
      tabindex="0"
      onkeydown={e => e.key === 'Enter' && document.getElementById('sb-file-input-' + wsId).click()}
    >
      <div class="text-2xl mb-1.5">↑</div>
      <p class="text-xs">Import model</p>
      <div class="text-[11px] mt-0.5 opacity-70">.xml (AOEF)</div>
    </div>
    <input
      type="file"
      id="sb-file-input-{wsId}"
      accept=".xml"
      style="display:none"
      onchange={handleFileInput}
    />
    {#if importing}
      <div class="flex items-center gap-2 text-muted-foreground py-2 mt-2">
        <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      </div>
    {/if}
  </div>
</aside>

<Dialog.Root bind:open={showErrorDialog} onOpenChange={(o) => { if (!o) importError = null; }}>
  <Dialog.Content class="max-w-sm">
    <Dialog.Header>
      <Dialog.Title>Import error</Dialog.Title>
      <Dialog.Description>{friendlyError(importError)}</Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer>
      <Button onclick={() => importError = null}>OK</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={showPreview} onOpenChange={(o) => { if (!o) cancelPreview(); }}>
  <Dialog.Content class="max-w-lg">
    <Dialog.Header>
      <Dialog.Title>Import preview</Dialog.Title>
      <Dialog.Description>Review what will change before confirming the import.</Dialog.Description>
    </Dialog.Header>
    {#if preview}
      <div class="space-y-2 max-h-[50vh] overflow-y-auto pr-1">
        {#each Object.entries(preview) as [key, cat]}
          {#if cat.added > 0 || cat.modified > 0 || cat.unchanged > 0}
            <div class="border border-border rounded-md overflow-hidden">
              <div
                class="flex items-center justify-between px-3 py-2 bg-muted/40 cursor-pointer select-none"
                role="button"
                tabindex="0"
                onclick={() => toggleCategory(key)}
                onkeydown={e => e.key === 'Enter' && toggleCategory(key)}
              >
                <span class="text-[13px] font-medium">{categoryLabel(key)}</span>
                <div class="flex items-center gap-2 text-[12px]">
                  {#if cat.added > 0}
                    <span class="text-green-600 font-medium">+{cat.added}</span>
                  {/if}
                  {#if cat.modified > 0}
                    <span class="text-amber-600 font-medium">~{cat.modified}</span>
                  {/if}
                  {#if cat.unchanged > 0}
                    <span class="text-muted-foreground">{cat.unchanged} unchanged</span>
                  {/if}
                  {#if cat.added > 0 || cat.modified > 0}
                    <span class="text-muted-foreground text-[11px] ml-1">{expandedCategories[key] ? '▲' : '▼'}</span>
                  {/if}
                </div>
              </div>
              {#if expandedCategories[key] && cat.details?.length > 0}
                <div class="divide-y divide-border max-h-48 overflow-y-auto">
                  {#each cat.details as item, i}
                    <div class="flex items-start gap-2 px-3 py-1.5 text-[12px]">
                      {#if i < cat.added}
                        <span class="flex-shrink-0 w-4 text-center text-green-600 font-bold">+</span>
                      {:else}
                        <span class="flex-shrink-0 w-4 text-center text-amber-600 font-bold">~</span>
                      {/if}
                      <span class="flex-1 truncate text-foreground">{item.name || item.source_id}</span>
                      {#if item.type}
                        <span class="text-muted-foreground flex-shrink-0">{item.type}</span>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}
        {/each}
      </div>
    {/if}
    <Dialog.Footer class="gap-2">
      <Button variant="outline" onclick={cancelPreview} disabled={confirming}>Cancel</Button>
      <Button onclick={confirmImport} disabled={confirming}>
        {#if confirming}
          <span class="flex items-center gap-1.5">
            <span class="size-3.5 rounded-full border-2 border-white/40 border-t-white animate-spin"></span>
            Importing…
          </span>
        {:else}
          Confirm import
        {/if}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
