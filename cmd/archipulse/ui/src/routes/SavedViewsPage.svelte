<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { api } from '../lib/api.js';
  import { savedViewsRevision } from '../lib/saved-views-events.js';
  import { Button } from '$lib/components/ui/button';
  import * as Dialog from '$lib/components/ui/dialog';

  export let params = {};
  $: wsId = params.wsId;

  let views = [];
  let loading = true;
  let error = null;
  let deleteTarget = null;
  let showDeleteDialog = false;
  let deleting = false;

  $: showDeleteDialog = !!deleteTarget;

  const VIEW_TYPE_LABELS = {
    'application-dashboard':  'Application Dashboard',
    'capability-landscape':   'Capability Landscape',
    'application-landscape':  'Application Landscape',
    'capability-tree':        'Capability Tree',
    'application-dependency': 'Dependency Graph',
  };

  const VIEW_TYPE_DOTS = {
    'application-dashboard':  '#2563eb',
    'capability-landscape':   '#d97706',
    'application-landscape':  '#2563eb',
    'capability-tree':        '#d97706',
    'application-dependency': '#2563eb',
  };

  const ALL_RELS = ['serving', 'flow', 'access', 'triggering', 'association'];

  function filterSummary(sv) {
    const f = sv.filters ?? {};
    const parts = [];
    if (f.capability) parts.push(f.capability);
    if (f.overlay) parts.push(f.overlay.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()));
    if (f.colorBy) parts.push(f.colorBy.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()));
    if (f.heatmap && f.heatmap !== 'none') parts.push({ appCount: 'Coverage heatmap', gap: 'Gap heatmap', avgCrit: 'Criticality heatmap' }[f.heatmap] ?? f.heatmap);
    if (f.focusedCapabilityName) parts.push(f.focusedCapabilityName);
    if (f.focusedApps?.length) parts.push(`${f.focusedApps.length} app${f.focusedApps.length !== 1 ? 's' : ''} focused`);
    if (f.activeRels?.length && f.activeRels.length < ALL_RELS.length) {
      parts.push(f.activeRels.join(', '));
    }
    return parts.length ? parts.join(' · ') : 'No filters';
  }

  function formatDate(iso) {
    return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
  }

  async function load() {
    loading = true;
    error = null;
    try {
      views = await api.get('/workspaces/' + wsId + '/saved-views');
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return;
    deleting = true;
    try {
      await api.delete('/workspaces/' + wsId + '/saved-views/' + deleteTarget.id);
      savedViewsRevision.update(n => n + 1);
      deleteTarget = null;
      await load();
    } catch (e) {
      error = e.message;
    } finally {
      deleting = false;
    }
  }

  onMount(() => load());
</script>

<div class="content">
  <div class="flex items-center justify-between mb-6">
    <div>
      <h1 class="text-[18px] font-semibold">Saved Views</h1>
      <p class="text-muted-foreground text-[13px] mt-0.5">Bookmarked automatic views with their filter state</p>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else if error}
    <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">{error}</div>
  {:else if views.length === 0}
    <div class="text-center py-20 text-muted-foreground">
      <div class="text-[40px] mb-3">⊛</div>
      <p class="text-[14px]">No saved views yet.</p>
      <p class="text-[13px] mt-1 opacity-70">Open any automatic view and click <strong>⊕ Save view</strong> to bookmark it.</p>
    </div>
  {:else}
    <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {#each views as sv}
        <div
          class="group bg-card border border-border rounded-lg p-4 cursor-pointer hover:border-primary hover:shadow-sm transition-all"
          onclick={() => push('/ws/' + wsId + '/saved-view/' + sv.id)}
          onkeydown={e => e.key === 'Enter' && push('/ws/' + wsId + '/saved-view/' + sv.id)}
          role="button"
          tabindex="0"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="flex items-center gap-2 min-w-0">
              <span class="size-2 rounded-full flex-shrink-0 mt-0.5" style="background:{VIEW_TYPE_DOTS[sv.view_type] ?? '#64748b'}"></span>
              <span class="text-[13px] font-medium text-foreground truncate">{sv.name}</span>
            </div>
            <button
              class="opacity-0 group-hover:opacity-100 flex-shrink-0 text-muted-foreground hover:text-destructive transition-all p-0.5 rounded"
              onclick={(e) => { e.stopPropagation(); deleteTarget = sv; }}
              aria-label="Delete saved view"
            >✕</button>
          </div>

          <div class="mt-2 space-y-1">
            <div class="text-[11px] text-muted-foreground">{VIEW_TYPE_LABELS[sv.view_type] ?? sv.view_type}</div>
            <div class="text-[11px] text-foreground/70">{filterSummary(sv)}</div>
          </div>

          <div class="mt-3 text-[11px] text-muted-foreground">{formatDate(sv.created_at)}</div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<Dialog.Root bind:open={showDeleteDialog} onOpenChange={(o) => { if (!o) deleteTarget = null; }}>
  <Dialog.Content class="max-w-sm">
    <Dialog.Header>
      <Dialog.Title>Delete saved view</Dialog.Title>
      <Dialog.Description>
        Delete <strong>{deleteTarget?.name}</strong>? This can't be undone.
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer class="gap-2">
      <Button variant="outline" onclick={() => deleteTarget = null} disabled={deleting}>Cancel</Button>
      <Button variant="destructive" onclick={confirmDelete} disabled={deleting}>
        {deleting ? 'Deleting…' : 'Delete'}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
