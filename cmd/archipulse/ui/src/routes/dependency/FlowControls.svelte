<script>
  import { useSvelteFlow, Panel } from '@xyflow/svelte';

  let { onReady } = $props();

  const { fitView } = useSvelteFlow();

  $effect(() => {
    onReady(fitView);
  });
</script>
