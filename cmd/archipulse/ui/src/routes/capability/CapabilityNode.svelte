<script>
  import { Handle, Position } from '@xyflow/svelte';

  let { data = {} } = $props();
</script>

<div style="
  background:#201808;
  border:2px solid #e0af68;
  color:#fcd990;
  font-weight:600;
  font-size:12px;
  min-width:160px;
  max-width:210px;
  padding:10px 14px;
  border-radius:8px;
  text-align:center;
  line-height:1.35;
  box-shadow:0 2px 14px rgba(0,0,0,0.55);
  cursor:default;
  user-select:none;
">
  <Handle type="target" position={Position.Left}  style="background:#e0af68; width:9px; height:9px; border:none; border-radius:50%;" />
  <div style="word-break:break-word;">{data.label}</div>
  {#if data.appCount > 0}
    <div style="font-size:9px; opacity:0.5; margin-top:3px; font-weight:400; letter-spacing:0.3px;">{data.appCount} app{data.appCount > 1 ? 's' : ''}</div>
  {/if}
  <Handle type="source" position={Position.Right} style="background:#e0af68; width:9px; height:9px; border:none; border-radius:50%;" />
</div>
