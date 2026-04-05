<script>
  import { Handle, Position } from '@xyflow/svelte';

  let { data = {} } = $props();
</script>

<div style="
  background:#fffbeb;
  border:2px solid #d97706;
  color:#78350f;
  font-weight:600;
  font-size:12px;
  min-width:160px;
  max-width:210px;
  padding:10px 14px;
  border-radius:8px;
  text-align:center;
  line-height:1.35;
  box-shadow:0 1px 4px rgba(0,0,0,0.10), 0 0 0 1px rgba(0,0,0,0.04);
  cursor:default;
  user-select:none;
">
  <Handle type="target" position={Position.Left}  style="background:#d97706; width:9px; height:9px; border:none; border-radius:50%;" />
  <div style="word-break:break-word;">{data.label}</div>
  {#if data.appCount > 0}
    <div style="font-size:9px; opacity:0.55; margin-top:3px; font-weight:400; letter-spacing:0.3px;">{data.appCount} app{data.appCount > 1 ? 's' : ''}</div>
  {/if}
  <Handle type="source" position={Position.Right} style="background:#d97706; width:9px; height:9px; border:none; border-radius:50%;" />
</div>
