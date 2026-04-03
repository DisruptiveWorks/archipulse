<script>
  import { Handle, Position } from '@xyflow/svelte';

  let { data = {} } = $props();

  const LIFECYCLE_STYLE = {
    'Production':     { border: '#22c55e', text: '#86efac', bg: '#14352a' },
    'Pilot':          { border: '#3b82f6', text: '#93c5fd', bg: '#172444' },
    'Planned':        { border: '#8b5cf6', text: '#c4b5fd', bg: '#22173a' },
    'Retiring':       { border: '#f97316', text: '#fdba74', bg: '#3a2010' },
    'Decommissioned': { border: '#ef4444', text: '#fca5a5', bg: '#3a1414' },
  };
  const DEFAULT_STYLE = { border: '#4a6fa5', text: '#93b4f0', bg: '#1e2d45' };

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
  box-shadow:0 2px 12px rgba(0,0,0,0.5);
  cursor:default;
  user-select:none;
">
  <Handle type="target" position={Position.Left}  style="background:{style.border}; width:9px; height:9px; border:none; border-radius:50%;" />
  <div style="word-break:break-word;">{data.label}</div>
  {#if data.badge && data.badge !== 'Component'}
    <div style="font-size:9px; opacity:0.6; margin-top:3px; font-weight:400; letter-spacing:0.3px;">{data.badge}</div>
  {/if}
  <Handle type="source" position={Position.Right} style="background:{style.border}; width:9px; height:9px; border:none; border-radius:50%;" />
</div>
