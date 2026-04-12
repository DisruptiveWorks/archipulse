<script>
  import { createEventDispatcher } from 'svelte';

  export let folder;
  export let collapsed;
  export let selectedDiagram;
  export let depth = 0;

  const dispatch = createEventDispatcher();

  $: isOpen = !collapsed[folder.id];
  $: indent = (depth + 1) * 12;
</script>

<!-- Folder row -->
<button
  class="w-full text-left flex items-center gap-1.5 py-1.5 text-[12px] hover:bg-muted/50 transition-colors text-foreground"
  style="padding-left: {4 + indent}px; padding-right: 8px;"
  onclick={() => dispatch('toggle', folder.id)}
>
  <span class="text-[9px] opacity-40 flex-shrink-0 transition-transform {isOpen ? '' : '-rotate-90'}">▼</span>
  <span class="text-amber-500 flex-shrink-0 text-[13px]">📁</span>
  <span class="truncate font-medium">{folder.name}</span>
</button>

{#if isOpen}
  <!-- Child folders (recursive) -->
  {#each (folder.children || []) as child}
    <svelte:self
      folder={child}
      {collapsed}
      {selectedDiagram}
      depth={depth + 1}
      on:toggle
      on:select
    />
  {/each}

  <!-- Diagrams in this folder -->
  {#each (folder.diagrams || []) as d}
    <button
      class="w-full text-left flex items-center gap-2 py-1.5 text-[12px] hover:bg-muted/50 transition-colors
        {selectedDiagram?.id === d.id ? 'bg-primary/10 text-primary font-medium' : 'text-foreground'}"
      style="padding-left: {4 + indent + 12}px; padding-right: 8px;"
      onclick={() => dispatch('select', d)}
    >
      <span class="text-[10px] opacity-40 flex-shrink-0">▪</span>
      <span class="truncate">{d.name || '(unnamed)'}</span>
    </button>
  {/each}
{/if}
