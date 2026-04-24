<script>
  import { createEventDispatcher } from 'svelte';
  import { push } from 'svelte-spa-router';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { api } from '../../lib/api.js';
  import { savedViewsRevision } from '../../lib/saved-views-events.js';

  export let open = false;
  export let wsId;
  export let viewType;
  export let filters = {};

  const dispatch = createEventDispatcher();

  let name = '';
  let saving = false;
  let error = null;

  function reset() {
    name = '';
    saving = false;
    error = null;
  }

  $: if (!open) reset();

  async function save() {
    if (!name.trim()) return;
    saving = true;
    error = null;
    try {
      const sv = await api.post('/workspaces/' + wsId + '/saved-views', {
        view_type: viewType,
        name: name.trim(),
        filters,
      });
      savedViewsRevision.update(n => n + 1);
      dispatch('saved', sv);
      open = false;
      push('/ws/' + wsId + '/saved-view/' + sv.id);
    } catch (e) {
      error = e.message;
    } finally {
      saving = false;
    }
  }

  function onKeydown(e) {
    if (e.key === 'Enter' && name.trim() && !saving) save();
  }
</script>

<Dialog.Root bind:open onOpenChange={(o) => { if (!o) reset(); }}>
  <Dialog.Content class="max-w-sm">
    <Dialog.Header>
      <Dialog.Title>Save view</Dialog.Title>
      <Dialog.Description>Give this view a name to find it later in the sidebar.</Dialog.Description>
    </Dialog.Header>

    <div class="space-y-3 py-1">
      <div class="space-y-1.5">
        <Label for="sv-name">Name</Label>
        <Input
          id="sv-name"
          bind:value={name}
          placeholder="e.g. Finance Apps — Cloud"
          onkeydown={onKeydown}
          autofocus
        />
      </div>
      {#if error}
        <p class="text-[12px] text-destructive">{error}</p>
      {/if}
    </div>

    <Dialog.Footer class="gap-2">
      <Button variant="outline" onclick={() => open = false} disabled={saving}>Cancel</Button>
      <Button onclick={save} disabled={!name.trim() || saving}>
        {#if saving}
          <span class="flex items-center gap-1.5">
            <span class="size-3.5 rounded-full border-2 border-white/40 border-t-white animate-spin"></span>
            Saving…
          </span>
        {:else}
          Save
        {/if}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
