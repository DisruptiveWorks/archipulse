<script>
  import { Handle, Position } from '@xyflow/svelte';
  import { getIcon } from '../diagram/archimate-icons.js';

  let { data = {} } = $props();

  const icon = getIcon('Capability');
</script>

<div style="
  position:relative;
  background:#fffbeb;
  border:2px solid #d97706;
  color:#78350f;
  font-weight:600;
  font-size:12px;
  min-width:160px;
  max-width:210px;
  padding:10px 28px 10px 14px;
  border-radius:8px;
  text-align:center;
  line-height:1.35;
  box-shadow:0 1px 4px rgba(0,0,0,0.10), 0 0 0 1px rgba(0,0,0,0.04);
  cursor:default;
  user-select:none;
">
  <Handle type="target" position={Position.Left}  style="background:#d97706; width:9px; height:9px; border:none; border-radius:50%;" />
  {#if icon}
    <svg viewBox="0 0 16 16" width="13" height="13"
      style="position:absolute; top:4px; right:4px; stroke:#d97706; fill:none; opacity:0.75; pointer-events:none; overflow:visible;">
      <!-- eslint-disable-next-line svelte/no-at-html-tags -->
      {@html icon}
    </svg>
  {/if}
  <div style="word-break:break-word;">{data.label}</div>
  {#if data.appCount > 0}
    <div style="font-size:9px; opacity:0.55; margin-top:3px; font-weight:400; letter-spacing:0.3px;">{data.appCount} app{data.appCount > 1 ? 's' : ''}</div>
  {/if}
  <Handle type="source" position={Position.Right} style="background:#d97706; width:9px; height:9px; border:none; border-radius:50%;" />
</div>
