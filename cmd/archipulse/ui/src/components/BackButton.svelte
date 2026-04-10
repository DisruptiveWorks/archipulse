<script>
  export let label = 'Back';
  export let onclick = () => {};
</script>

<button
  class="flex items-center gap-1.5 text-[12px] text-muted-foreground hover:text-foreground transition-colors mb-4 group"
  on:click={onclick}
>
  <span class="group-hover:-translate-x-0.5 transition-transform">←</span>
  {label}
</button>
