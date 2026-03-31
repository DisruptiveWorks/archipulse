<script>
	import { cn } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	} = $props();
</script>

<div
	bind:this={ref}
	data-slot="card-content"
	class={cn("px-4 group-data-[size=sm]/card:px-3", className)}
	{...restProps}
>
	{@render children?.()}
</div>