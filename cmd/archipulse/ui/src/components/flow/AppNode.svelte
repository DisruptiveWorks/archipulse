<script>
  import { Handle, Position } from '@xyflow/svelte';

  let { data = {} } = $props();

  const LIFECYCLE_STYLE = {
    'Production':     { border: '#16a34a', text: '#166534', bg: '#f0fdf4' },
    'Pilot':          { border: '#2563eb', text: '#1d4ed8', bg: '#eff6ff' },
    'Planned':        { border: '#7c3aed', text: '#6d28d9', bg: '#f5f3ff' },
    'Retiring':       { border: '#ea580c', text: '#c2410c', bg: '#fff7ed' },
    'Decommissioned': { border: '#dc2626', text: '#b91c1c', bg: '#fef2f2' },
  };
  const DEFAULT_STYLE = { border: '#2563eb', text: '#1e3a8a', bg: '#eff6ff' };

  const style       = $derived(LIFECYCLE_STYLE[data.lifecycle] ?? DEFAULT_STYLE);
  const isComponent = $derived(data.tier === 'component');
  const bw          = $derived(isComponent ? '2.5px' : '1.5px');
  const bs          = $derived(
    data.tier === 'service'   ? 'dashed' :
    data.tier === 'interface' || data.tier === 'function' ? 'dotted' : 'solid'
  );
</script>

<div style="
  background:{style.bg};
  border:{bw} {bs} {style.border};
  color:{style.text};
  font-weight:{isComponent ? 700 : 400};
  font-size:{isComponent ? '12px' : '11px'};
  min-width:{isComponent ? '148px' : '118px'};
  max-width:{isComponent ? '190px' : '160px'};
  padding:{isComponent ? '10px 14px' : '7px 11px'};
  border-radius:8px;
  text-align:center;
  line-height:1.35;
  box-shadow:0 1px 4px rgba(0,0,0,0.10), 0 0 0 1px rgba(0,0,0,0.04);
  cursor:default;
  user-select:none;
">
  <Handle type="target" position={Position.Left}  style="background:{style.border}; width:9px; height:9px; border:none; border-radius:50%;" />
  <div style="word-break:break-word;">{data.label}</div>
  {#if data.badge}
    <div style="font-size:9px; opacity:0.6; margin-top:3px; font-weight:400; letter-spacing:0.3px;">{data.badge}</div>
  {/if}
  <Handle type="source" position={Position.Right} style="background:{style.border}; width:9px; height:9px; border:none; border-radius:50%;" />
</div>
