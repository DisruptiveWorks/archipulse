<script>
  import * as Dialog from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';

  export let title = 'About this view';
  export let open = false;
</script>

<button
  onclick={() => open = true}
  class="flex items-center gap-1 px-2.5 py-1.5 rounded-md border border-border text-[12px] text-muted-foreground hover:text-foreground hover:border-primary transition-colors"
  aria-label="View documentation"
  title="How to use this view"
>ⓘ</button>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-lg">
    <Dialog.Header>
      <Dialog.Title>{title}</Dialog.Title>
    </Dialog.Header>
    <div class="overflow-y-auto space-y-4 text-[13px] text-foreground leading-relaxed pr-1" style="max-height: 58vh">
      <slot />
    </div>
    <Dialog.Footer class="mt-2">
      <Button onclick={() => open = false}>Got it</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
