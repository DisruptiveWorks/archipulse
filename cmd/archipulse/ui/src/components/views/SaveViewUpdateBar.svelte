<script>
  import { toast } from 'svelte-sonner';
  import { api } from '../../lib/api.js';

  const { wsId, savedViewId, savedViewName, currentFilters, initialFilters } = $props();

  // Deep-sort object keys so key-order differences between API response and
  // derived state don't trigger false positives. Arrays are preserved as-is.
  function stable(v) {
    if (v == null) return 'null';
    return JSON.stringify(v, (_, val) => {
      if (val !== null && typeof val === 'object' && !Array.isArray(val)) {
        return Object.fromEntries(Object.entries(val).sort(([a], [b]) => a < b ? -1 : 1));
      }
      return val;
    });
  }

  // Reset baseline whenever the saved view is (re)loaded with new initialFilters.
  let lastSaved = $state('null');
  $effect(() => {
    lastSaved = stable(initialFilters);
  });

  let showConfirm = $state(false);
  let saving = $state(false);

  const changed = $derived(stable(currentFilters) !== lastSaved);

  async function doUpdate() {
    saving = true;
    try {
      await api.put('/workspaces/' + wsId + '/saved-views/' + savedViewId, {
        name: savedViewName,
        filters: currentFilters,
      });
      lastSaved = JSON.stringify(currentFilters ?? {});
      showConfirm = false;
      toast.success('Saved view updated');
    } catch (e) {
      toast.error('Could not update: ' + e.message);
    } finally {
      saving = false;
    }
  }

  function cancel() {
    showConfirm = false;
  }
</script>

{#if savedViewId && changed}
  <div class="flex items-center gap-2 px-3 py-1.5 mb-4 rounded-lg border border-amber-200 bg-amber-50 text-[12px] text-amber-800 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-300">
    {#if showConfirm}
      <span class="flex-1">Overwrite filters for <strong>"{savedViewName}"</strong>?</span>
      <button
        onclick={doUpdate}
        disabled={saving}
        class="px-2.5 py-1 rounded-md bg-amber-600 text-white font-medium hover:bg-amber-700 disabled:opacity-50 transition-colors">
        {saving ? 'Saving…' : 'Confirm'}
      </button>
      <button
        onclick={cancel}
        disabled={saving}
        class="px-2.5 py-1 rounded-md border border-amber-300 hover:bg-amber-100 transition-colors">
        Cancel
      </button>
    {:else}
      <span class="flex-1">Filters changed from saved view</span>
      <button
        onclick={() => showConfirm = true}
        class="px-2.5 py-1 rounded-md border border-amber-400 font-medium hover:bg-amber-100 transition-colors">
        Update saved view
      </button>
    {/if}
  </div>
{/if}
