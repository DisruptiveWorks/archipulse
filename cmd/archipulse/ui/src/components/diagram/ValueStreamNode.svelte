<script>
  // ValueStream: chevron/arrow shape per ArchiMate standard.
  // Container (has children): flat-left arrow, label at top.
  // Leaf step: shallow left-notch arrow, label centered.
  import { Handle, Position } from '@xyflow/svelte';
  export let data;

  const AMBER = { fill: '#FFFBEB', stroke: '#D97706', text: '#78350F' };

  // viewBox 0 0 100 40
  // Both container and leaf: flat left (vertical), arrow right.
  // Container uses a shallower arrow; leaf a proportionally deeper one.
  $: points = data.isContainer
    ? '0,0 88,0 100,20 88,40 0,40, 5,20'
    : '0,0 86,0 100,20 86,40 0,40 15,20';
</script>

<Handle type="source" position={Position.Right} style="opacity:0;pointer-events:none;" />
<Handle type="target" position={Position.Left} style="opacity:0;pointer-events:none;" />
<div style="width:100%;height:100%;position:relative;overflow:visible;pointer-events:none;">
  <svg
    viewBox="0 0 100 40"
    preserveAspectRatio="none"
    style="position:absolute;top:0;left:0;width:100%;height:100%;overflow:visible;pointer-events:none;"
  >
    <polygon
      {points}
      fill={AMBER.fill}
      stroke={AMBER.stroke}
      stroke-width="1.5"
      vector-effect="non-scaling-stroke"
    />
  </svg>

  {#if data.isContainer}
    <!-- Container: label at top, children rendered by XY Flow inside -->
    <div style="
      position: absolute;
      top: 6px;
      left: 0;
      right: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 0 14px;
      box-sizing: border-box;
      pointer-events: none;
    ">
      <span style="
        color: {AMBER.text};
        font-size: 11px;
        font-weight: 600;
        font-family: ui-sans-serif, system-ui, sans-serif;
        text-align: center;
        line-height: 1.3;
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
      padding: 4px 18px;
      box-sizing: border-box;
      pointer-events: none;
    ">
      <span style="
        color: {AMBER.text};
        font-size: 11px;
        font-weight: 500;
        font-family: ui-sans-serif, system-ui, sans-serif;
        text-align: center;
        line-height: 1.35;
        word-break: break-word;
        hyphens: auto;
      ">{data.label}</span>
    </div>
  {/if}
</div>
