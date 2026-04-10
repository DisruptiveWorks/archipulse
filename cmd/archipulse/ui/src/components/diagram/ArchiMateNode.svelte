<script>
  import { NodeResizer } from '@xyflow/svelte';

  export let data;

  const LAYER_COLORS = {
    ApplicationComponent:     { fill: '#1a2b4a', stroke: '#7aa2f7', text: '#7aa2f7' },
    ApplicationService:       { fill: '#1a2b4a', stroke: '#7aa2f7', text: '#7aa2f7' },
    ApplicationFunction:      { fill: '#1a2b4a', stroke: '#7aa2f7', text: '#7aa2f7' },
    ApplicationInterface:     { fill: '#1a2b4a', stroke: '#7aa2f7', text: '#7aa2f7' },
    ApplicationCollaboration: { fill: '#1a2b4a', stroke: '#7aa2f7', text: '#7aa2f7' },
    BusinessActor:            { fill: '#2a1f0a', stroke: '#e0af68', text: '#e0af68' },
    BusinessRole:             { fill: '#2a1f0a', stroke: '#e0af68', text: '#e0af68' },
    BusinessProcess:          { fill: '#2a1f0a', stroke: '#e0af68', text: '#e0af68' },
    BusinessFunction:         { fill: '#2a1f0a', stroke: '#e0af68', text: '#e0af68' },
    BusinessService:          { fill: '#2a1f0a', stroke: '#e0af68', text: '#e0af68' },
    BusinessObject:           { fill: '#2a1f0a', stroke: '#e0af68', text: '#e0af68' },
    Capability:               { fill: '#2a1f0a', stroke: '#e0af68', text: '#e0af68' },
    TechnologyService:        { fill: '#0f2a1a', stroke: '#9ece6a', text: '#9ece6a' },
    TechnologyComponent:      { fill: '#0f2a1a', stroke: '#9ece6a', text: '#9ece6a' },
    Node:                     { fill: '#0f2a1a', stroke: '#9ece6a', text: '#9ece6a' },
    SystemSoftware:           { fill: '#0f2a1a', stroke: '#9ece6a', text: '#9ece6a' },
    Grouping:                 { fill: 'transparent', stroke: '#565f89', text: '#a9b1d6' },
  };

  const DEFAULT_COLOR = { fill: '#1a1b26', stroke: '#565f89', text: '#a9b1d6' };

  $: c = LAYER_COLORS[data.elementType] || DEFAULT_COLOR;
</script>

<div
  class="archimate-node"
  style="
    width: 100%;
    height: 100%;
    background: {c.fill};
    border: 1.5px solid {c.stroke};
    border-radius: 3px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4px 6px;
    box-sizing: border-box;
    overflow: hidden;
  "
>
  <span
    style="
      color: {c.text};
      font-size: 11px;
      font-family: ui-sans-serif, system-ui, sans-serif;
      text-align: center;
      line-height: 1.3;
      word-break: break-word;
      hyphens: auto;
    "
  >{data.label}</span>
</div>
