<script>
  import { Handle, Position } from '@xyflow/svelte';
  import { getIcon, getColor } from './archimate-icons.js';

  export let data;

  $: c = getColor(data.elementType);
  $: icon = getIcon(data.elementType);
  $: isContainer = data.isContainer || false;
  $: dashed = c.dashed || false;
</script>

<Handle type="source" position={Position.Right} style="opacity:0;pointer-events:none;" />
<Handle type="target" position={Position.Left} style="opacity:0;pointer-events:none;" />
<div
  style="
    width: 100%;
    height: 100%;
    background: {c.fill};
    border: 1.5px {dashed ? 'dashed' : 'solid'} {c.stroke};
    border-radius: 4px;
    position: relative;
    font-family: ui-sans-serif, system-ui, sans-serif;
    box-sizing: border-box;
    overflow: visible;
    box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  "
>
  <!-- Type icon — always top-right -->
  <svg
    viewBox="0 0 16 16"
    width="16"
    height="16"
    style="
      position: absolute;
      top: 4px;
      right: 4px;
      color: {c.stroke};
      stroke: {c.stroke};
      fill: none;
      flex-shrink: 0;
      overflow: visible;
      pointer-events: none;
    "
  >
    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
    {@html icon}
  </svg>

  {#if isContainer}
    <!-- Container: label at top, children rendered by XY Flow inside -->
    <div style="
      position: absolute;
      top: 6px;
      left: 0;
      right: 22px;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 0 8px;
      box-sizing: border-box;
    ">
      <span style="
        color: {c.text};
        font-size: 11px;
        font-weight: 600;
        text-align: center;
        line-height: 1.3;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 100%;
      ">{data.label}</span>
    </div>
  {:else}
    <!-- Leaf: label centered -->
    <div style="
      position: absolute;
      inset: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 4px 22px 4px 6px;
      box-sizing: border-box;
    ">
      <span style="
        color: {c.text};
        font-size: 11px;
        font-weight: 500;
        text-align: center;
        line-height: 1.35;
        word-break: break-word;
        hyphens: auto;
      ">{data.label}</span>
    </div>
  {/if}
</div>
