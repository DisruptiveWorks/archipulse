<script>
  import { Handle, Position } from '@xyflow/svelte';

  let { data = {} } = $props();

  const LIFECYCLE_STYLE = {
    'Production':     { border: '#4ade80', text: '#4ade80', bg: '#0d2a1a' },
    'Pilot':          { border: '#60a5fa', text: '#60a5fa', bg: '#0d1f38' },
    'Planned':        { border: '#a78bfa', text: '#a78bfa', bg: '#1a1228' },
    'Retiring':       { border: '#fb923c', text: '#fb923c', bg: '#2a1800' },
    'Decommissioned': { border: '#f87171', text: '#f87171', bg: '#2a0d0d' },
  };
  const DEFAULT_STYLE = { border: '#3d59a1', text: '#7aa2f7', bg: '#131929' };

  const style       = $derived(LIFECYCLE_STYLE[data.lifecycle] ?? DEFAULT_STYLE);
  const isComponent = $derived(data.tier === 'component');
  const borderStyle = $derived(
    data.tier === 'service'   ? 'dashed' :
    data.tier === 'interface' || data.tier === 'function' ? 'dotted' : 'solid'
  );
</script>

<div
  style="
    background: {style.bg};
    border: {isComponent ? '2.5px' : '1.5px'} {borderStyle} {style.border};
    color: {style.text};
    font-weight: {isComponent ? '700' : '400'};
    font-size: {isComponent ? '12px' : '11px'};
    min-width: {isComponent ? '140px' : '110px'};
    max-width: {isComponent ? '180px' : '156px'};
    padding: {isComponent ? '9px 13px' : '6px 10px'};
    border-radius: 8px;
    text-align: center;
    line-height: 1.3;
    box-shadow: 0 2px 8px rgba(0,0,0,0.4);
    cursor: default;
    user-select: none;
  "
>
  <Handle type="target" position={Position.Left}  style="background:{style.border}; width:8px; height:8px; border:none;" />
  <div style="word-break:break-word;">{data.label}</div>
  {#if data.badge && data.badge !== 'Component'}
    <div style="font-size:9px; opacity:0.55; margin-top:2px; font-weight:400;">{data.badge}</div>
  {/if}
  <Handle type="source" position={Position.Right} style="background:{style.border}; width:8px; height:8px; border:none;" />
</div>
