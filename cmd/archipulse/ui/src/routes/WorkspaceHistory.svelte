<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { user } from '../lib/auth.js';

  export let params = {};
  $: wsId = params.wsId;

  let events = [];
  let snapshots = [];
  let loading = true;
  let snapsLabel = '';
  let creating = false;
  let restoring = null;
  let error = null;

  $: canEdit = $user?.org_role === 'admin' || ['owner', 'editor'].includes(myWsRole);
  $: canDelete = $user?.org_role === 'admin' || myWsRole === 'owner';
  let myWsRole = null;

  $: if (wsId) load(wsId);

  async function load(id) {
    loading = true;
    error = null;
    myWsRole = null;
    try {
      const [evts, snaps, members] = await Promise.all([
        api.get('/workspaces/' + id + '/events?limit=50'),
        api.get('/workspaces/' + id + '/snapshots'),
        api.get('/workspaces/' + id + '/members'),
      ]);
      events = evts || [];
      snapshots = snaps || [];
      const me = members.find(m => m.user_id === $user?.id);
      myWsRole = me?.role ?? null;
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function createSnapshot() {
    creating = true;
    try {
      const snap = await api.post('/workspaces/' + wsId + '/snapshots', { label: snapsLabel.trim() || null });
      snapshots = [snap, ...snapshots];
      snapsLabel = '';
    } catch (e) {
      error = e.message;
    } finally {
      creating = false;
    }
  }

  async function deleteSnapshot(id) {
    if (!confirm('Delete this snapshot?')) return;
    try {
      await api.delete('/workspaces/' + wsId + '/snapshots/' + id);
      snapshots = snapshots.filter(s => s.id !== id);
    } catch (e) {
      error = e.message;
    }
  }

  async function restoreSnapshot(snap) {
    if (!confirm(`Restore snapshot "${snap.label || formatDate(snap.created_at)}"? Current model will be replaced.`)) return;
    restoring = snap.id;
    try {
      await api.post('/workspaces/' + wsId + '/snapshots/' + snap.id + '/restore', {});
      await load(wsId);
    } catch (e) {
      error = e.message;
    } finally {
      restoring = null;
    }
  }

  function formatDate(iso) {
    return new Date(iso).toLocaleString(undefined, {
      month: 'short', day: 'numeric', year: 'numeric',
      hour: '2-digit', minute: '2-digit',
    });
  }

  function actionLabel(e) {
    const labels = {
      create: 'created',
      update: 'updated',
      delete: 'deleted',
      import: 'imported model',
      add_member: 'added member',
      remove_member: 'removed member',
      update_member_role: 'updated member role',
      create_snapshot: 'created snapshot',
      restore_snapshot: 'restored snapshot',
    };
    return labels[e.action] ?? e.action;
  }

  function entityLabel(e) {
    if (e.entity_name) return `"${e.entity_name}"`;
    if (e.entity_type === 'member') return e.entity_id;
    return e.entity_type;
  }

  function metaDetail(e) {
    if (!e.meta) return '';
    try {
      const m = typeof e.meta === 'string' ? JSON.parse(e.meta) : e.meta;
      if (e.action === 'import') return `· ${m.elements} elements, ${m.relationships} relationships`;
      if (e.action === 'add_member' || e.action === 'update_member_role') return `· as ${m.role}`;
    } catch { /* ignore */ }
    return '';
  }
</script>

<div class="content">
  <h1 class="text-[18px] font-semibold mb-6">History</h1>

  {#if loading}
    <div class="flex items-center gap-2 text-muted-foreground py-6">
      <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
      Loading…
    </div>
  {:else}

    {#if error}
      <div class="mb-4 text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">{error}</div>
    {/if}

    <!-- Snapshots -->
    <div class="mb-8">
      <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-3">Snapshots</div>

      {#if canEdit}
        <div class="flex gap-2 mb-4">
          <input
            type="text"
            bind:value={snapsLabel}
            placeholder="Label (optional)"
            class="flex-1 text-sm border border-border rounded-md px-3 py-1.5 bg-background focus:outline-none focus:ring-1 focus:ring-primary"
          />
          <button
            onclick={createSnapshot}
            disabled={creating}
            class="text-sm bg-primary text-primary-foreground rounded-md px-4 py-1.5 hover:bg-primary/90 disabled:opacity-50 transition-colors"
          >
            {creating ? 'Saving…' : 'Save snapshot'}
          </button>
        </div>
      {/if}

      {#if snapshots.length === 0}
        <div class="text-sm text-muted-foreground py-4 text-center border border-dashed border-border rounded-lg">
          No snapshots yet. Snapshots are created automatically before each import, or manually here.
        </div>
      {:else}
        <div class="divide-y divide-border border border-border rounded-lg overflow-hidden">
          {#each snapshots as snap}
            <div class="flex items-center justify-between px-4 py-3 bg-card hover:bg-muted/30 transition-colors">
              <div>
                <div class="text-[13px] font-medium">
                  {#if snap.label}{snap.label}{:else}<span class="text-muted-foreground italic">Auto</span>{/if}
                  <span class="ml-2 text-[10px] font-normal px-1.5 py-0.5 rounded bg-muted text-muted-foreground uppercase">{snap.trigger}</span>
                </div>
                <div class="text-[11px] text-muted-foreground mt-0.5">{formatDate(snap.created_at)} · {snap.created_by_email}</div>
              </div>
              <div class="flex items-center gap-2 flex-shrink-0">
                {#if canEdit}
                  <button
                    onclick={() => restoreSnapshot(snap)}
                    disabled={restoring === snap.id}
                    class="text-[12px] px-3 py-1 rounded-md border border-border hover:border-primary hover:text-primary transition-colors disabled:opacity-50"
                  >
                    {restoring === snap.id ? 'Restoring…' : 'Restore'}
                  </button>
                {/if}
                {#if canDelete}
                  <button
                    onclick={() => deleteSnapshot(snap.id)}
                    class="text-[12px] px-2 py-1 rounded-md text-muted-foreground hover:text-destructive transition-colors"
                    title="Delete snapshot"
                  >✕</button>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Activity feed -->
    <div>
      <div class="text-[11px] font-bold tracking-[0.6px] uppercase text-muted-foreground mb-3">Activity</div>

      {#if events.length === 0}
        <div class="text-sm text-muted-foreground py-4 text-center border border-dashed border-border rounded-lg">
          No activity recorded yet.
        </div>
      {:else}
        <div class="divide-y divide-border border border-border rounded-lg overflow-hidden">
          {#each events as e}
            <div class="flex items-start gap-3 px-4 py-3 bg-card">
              <div class="size-6 rounded-full bg-primary/10 text-primary text-[10px] font-bold flex items-center justify-center flex-shrink-0 mt-0.5">
                {e.user_email[0].toUpperCase()}
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-[13px]">
                  <span class="font-medium">{e.user_email}</span>
                  <span class="text-muted-foreground"> {actionLabel(e)} {entityLabel(e)} {metaDetail(e)}</span>
                </div>
                <div class="text-[11px] text-muted-foreground mt-0.5">{formatDate(e.created_at)}</div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

  {/if}
</div>
