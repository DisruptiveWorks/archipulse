import cytoscape from 'cytoscape';
import cytoscapeDagre from 'cytoscape-dagre';
import dagre from 'dagre';

let registered = false;
function ensureRegistered() {
  if (!registered) {
    cytoscapeDagre(cytoscape, dagre);
    registered = true;
  }
}

// ── Relationship metadata ────────────────────────────────────────────────────

const REL_META = {
  serving:     { label: 'Serves',         color: '#7aa2f7', dash: false },
  flow:        { label: 'Data Flow',       color: '#9ece6a', dash: true  },
  access:      { label: 'Accesses',        color: '#bb9af7', dash: true  },
  association: { label: 'Associated with', color: '#6b7280', dash: true  },
  triggering:  { label: 'Triggers',        color: '#E85D3A', dash: false },
};

function relKey(rel) {
  const r = (rel || '').toLowerCase();
  if (r.includes('serving'))     return 'serving';
  if (r.includes('flow'))        return 'flow';
  if (r.includes('access'))      return 'access';
  if (r.includes('triggering'))  return 'triggering';
  if (r.includes('association')) return 'association';
  return 'association';
}

function relMeta(rel) {
  return REL_META[relKey(rel)] ?? REL_META.association;
}

// ── Lifecycle → node color ────────────────────────────────────────────────────

const LIFECYCLE_COLOR = {
  'Production':     { bg: '#0d2a1a', border: '#4ade80', text: '#4ade80' },
  'Pilot':          { bg: '#0d1f38', border: '#60a5fa', text: '#60a5fa' },
  'Planned':        { bg: '#1a1228', border: '#a78bfa', text: '#a78bfa' },
  'Retiring':       { bg: '#2a1800', border: '#fb923c', text: '#fb923c' },
  'Decommissioned': { bg: '#2a0d0d', border: '#f87171', text: '#f87171' },
};

function nodeColors(lifecycleStatus) {
  return LIFECYCLE_COLOR[lifecycleStatus] ?? { bg: '#131929', border: '#3d59a1', text: '#7aa2f7' };
}

// ── Type visual tier ──────────────────────────────────────────────────────────

function nodeTier(type) {
  if (type === 'ApplicationComponent') return 'component';
  if (type === 'ApplicationService')   return 'service';
  if (type === 'ApplicationInterface') return 'interface';
  if (type === 'ApplicationFunction')  return 'function';
  return 'other';
}

function nodeSize(tier) {
  if (tier === 'component')  return { w: 160, h: 44 };
  if (tier === 'service')    return { w: 140, h: 38 };
  if (tier === 'interface')  return { w: 130, h: 36 };
  if (tier === 'function')   return { w: 130, h: 36 };
  return { w: 140, h: 38 };
}

function typeBadge(type) {
  return type.replace('Application', '');
}

// ── Dependency Graph ──────────────────────────────────────────────────────────

export function makeDependencyGraph(container, data, { onEdgeHover, onEdgeLeave, onNodeClick } = {}) {
  ensureRegistered();

  const elements = [
    ...(data.nodes || []).map(n => {
      const tier = nodeTier(n.type);
      const sz   = nodeSize(tier);
      const clr  = nodeColors(n.lifecycle_status);
      return {
        data: {
          id:        n.id,
          label:     n.name,
          badge:     typeBadge(n.type),
          tier,
          lifecycle: n.lifecycle_status || '',
          bg:        clr.bg,
          border:    clr.border,
          textColor: clr.text,
          w:         sz.w,
          h:         sz.h,
        },
      };
    }),
    ...(data.edges || []).map(e => {
      const meta = relMeta(e.relationship);
      const rk   = relKey(e.relationship);
      return {
        data: {
          id:       e.id,
          source:   e.source,
          target:   e.target,
          relKey:   rk,
          relLabel: meta.label,
          color:    meta.color,
          dashed:   meta.dash,
          // Raw names filled in after nodes are indexed
          sourceName: '',
          targetName: '',
        },
      };
    }),
  ];

  // Index node names for edge tooltips
  const nameById = {};
  (data.nodes || []).forEach(n => { nameById[n.id] = n.name; });
  elements.forEach(el => {
    if (el.data.relKey) {
      el.data.sourceName = nameById[el.data.source] ?? el.data.source;
      el.data.targetName = nameById[el.data.target] ?? el.data.target;
    }
  });

  const cy = cytoscape({
    container,
    elements,
    style: [
      // ── Nodes ──
      {
        selector: 'node',
        style: {
          shape: 'round-rectangle',
          label: 'data(label)',
          'text-valign': 'center',
          'text-halign': 'center',
          'font-size': '11px',
          'font-family': 'system-ui, sans-serif',
          color: 'data(textColor)',
          'text-wrap': 'wrap',
          'text-max-width': '140px',
          width: 'data(w)',
          height: 'data(h)',
          'background-color': 'data(bg)',
          'border-color': 'data(border)',
          'border-width': 2,
          'transition-property': 'opacity, border-width',
          'transition-duration': '150ms',
        },
      },
      {
        selector: 'node[tier="component"]',
        style: { 'border-width': 2.5, 'font-size': '12px', 'font-weight': 'bold' },
      },
      {
        selector: 'node[tier="service"]',
        style: { 'border-style': 'dashed', 'font-size': '11px' },
      },
      {
        selector: 'node[tier="interface"]',
        style: { 'border-style': 'dotted', 'font-size': '10px' },
      },
      {
        selector: 'node[tier="function"]',
        style: { 'border-style': 'dotted', 'font-size': '10px' },
      },
      {
        selector: 'node.selected',
        style: { 'border-width': 4, 'border-color': '#fff' },
      },
      {
        selector: 'node.faded',
        style: { opacity: 0.15 },
      },
      // ── Edges ──
      {
        selector: 'edge',
        style: {
          width: 1.5,
          'line-color': 'data(color)',
          'target-arrow-color': 'data(color)',
          'target-arrow-shape': 'triangle',
          'curve-style': 'bezier',
          opacity: 0.75,
          'transition-property': 'opacity, width',
          'transition-duration': '150ms',
        },
      },
      {
        selector: 'edge[?dashed]',
        style: { 'line-style': 'dashed', 'line-dash-pattern': [6, 3] },
      },
      {
        selector: 'edge.faded',
        style: { opacity: 0.06 },
      },
      {
        selector: 'edge.hovered',
        style: { width: 3, opacity: 1 },
      },
    ],
    layout: {
      name: 'dagre',
      rankDir: 'LR',
      nodeSep: 28,
      rankSep: 90,
      edgeSep: 10,
      align: 'UL',
      animate: false,
      padding: 48,
    },
    userZoomingEnabled: true,
    userPanningEnabled: true,
    boxSelectionEnabled: false,
    minZoom: 0.2,
    maxZoom: 3,
  });

  // ── Interactions ──────────────────────────────────────────────────────────

  // Node click → highlight neighbourhood
  cy.on('tap', 'node', evt => {
    const n = evt.target;
    cy.elements().addClass('faded');
    n.removeClass('faded').addClass('selected');
    n.neighborhood().removeClass('faded');
    if (onNodeClick) onNodeClick(n.data());
  });

  // Tap background → clear selection
  cy.on('tap', evt => {
    if (evt.target === cy) {
      cy.elements().removeClass('faded selected');
      if (onNodeClick) onNodeClick(null);
    }
  });

  // Edge hover → tooltip
  cy.on('mouseover', 'edge', evt => {
    evt.target.addClass('hovered');
    if (onEdgeHover) onEdgeHover(evt.target.data(), evt.originalEvent);
  });
  cy.on('mouseout', 'edge', evt => {
    evt.target.removeClass('hovered');
    if (onEdgeLeave) onEdgeLeave();
  });
  cy.on('mousemove', 'edge', evt => {
    if (onEdgeHover) onEdgeHover(evt.target.data(), evt.originalEvent);
  });

  return cy;
}

// ── Legacy graph (integration map) ───────────────────────────────────────────

function relColor(rel) {
  const r = (rel || '').toLowerCase();
  if (r.includes('serving'))     return '#7aa2f7';
  if (r.includes('triggering'))  return '#E85D3A';
  if (r.includes('flow'))        return '#9ece6a';
  if (r.includes('access'))      return '#bb9af7';
  if (r.includes('assignment'))  return '#e0af68';
  if (r.includes('association')) return '#6b7280';
  return '#3a3d55';
}

export function makeGraph(container, data) {
  const cy = cytoscape({
    container,
    elements: [
      ...(data.nodes || []).map(n => ({
        data: { id: n.id, label: n.label || n.name, layer: n.layer, nodeType: n.type },
      })),
      ...(data.edges || []).map(e => ({
        data: {
          id: e.id,
          source: e.source,
          target: e.target,
          label: e.label || e.relationship,
          relColor: relColor(e.relationship),
        },
      })),
    ],
    style: [
      {
        selector: 'node',
        style: {
          label: 'data(label)',
          'text-valign': 'bottom',
          'text-halign': 'center',
          'font-size': '10px',
          color: '#e2e4f0',
          'text-margin-y': '4px',
          width: 34,
          height: 34,
          'background-color': '#555',
          'border-width': 2,
          'border-color': '#2a2d3e',
          'text-wrap': 'wrap',
          'text-max-width': '100px',
        },
      },
      { selector: 'node[layer="Application"]', style: { 'background-color': '#7aa2f7', 'border-color': '#3d59a1' } },
      { selector: 'node[layer="Business"]',    style: { 'background-color': '#e0af68', 'border-color': '#a87c36' } },
      { selector: 'node[layer="Technology"]',  style: { 'background-color': '#9ece6a', 'border-color': '#5a8a2a' } },
      { selector: 'node[layer="Motivation"]',  style: { 'background-color': '#bb9af7', 'border-color': '#7555a0' } },
      { selector: 'node[nodeType="ApplicationService"]',   style: { 'background-color': '#4a9eff', 'border-color': '#2a6abf', 'border-style': 'dashed' } },
      { selector: 'node[nodeType="ApplicationFunction"]',  style: { 'background-color': '#3a5a80', 'border-color': '#2a3a50' } },
      { selector: 'node[nodeType="ApplicationInterface"]', style: { 'background-color': '#3a9a9a', 'border-color': '#1a6a6a' } },
      { selector: 'node[nodeType="DataObject"]',           style: { 'background-color': '#7c5cbf', 'border-color': '#4a2a8a', width: 28, height: 28 } },
      {
        selector: 'edge',
        style: {
          width: 1.5,
          'line-color': 'data(relColor)',
          'target-arrow-color': 'data(relColor)',
          'target-arrow-shape': 'triangle',
          'curve-style': 'bezier',
          'font-size': '9px',
          color: '#8b8fa8',
          label: 'data(label)',
          'text-rotation': 'autorotate',
          'text-margin-y': '-6px',
          opacity: 0.8,
        },
      },
      { selector: ':selected', style: { 'border-color': '#fff', 'border-width': 3 } },
      { selector: '.faded',    style: { opacity: 0.12 } },
    ],
    layout: { name: 'cose', animate: false, padding: 40, nodeRepulsion: 8000 },
  });

  cy.on('tap', 'node', evt => {
    const n = evt.target;
    cy.elements().addClass('faded');
    n.removeClass('faded');
    n.neighborhood().removeClass('faded');
  });
  cy.on('tap', evt => {
    if (evt.target === cy) cy.elements().removeClass('faded');
  });

  return cy;
}

// ── Capability Tree ───────────────────────────────────────────────────────────

function appSubtype(t) {
  if (!t) return 'other';
  if (t === 'ApplicationComponent') return 'component';
  if (t === 'ApplicationService')   return 'service';
  if (t === 'ApplicationFunction')  return 'function';
  if (t === 'ApplicationInterface') return 'interface';
  return 'other';
}

function appShortType(t) {
  if (!t) return '';
  return t.replace('Application', '').replace('Composite', '');
}

export function makeCapabilityTree(container, nodes) {
  ensureRegistered();

  const elems = [];
  const appNodesSeen = new Set();

  nodes.forEach(n => {
    elems.push({ data: { id: n.id, label: n.name, kind: 'capability', apps: n.supporting_apps || [] } });
    if (n.parent_id) {
      elems.push({ data: { id: 'e-' + n.parent_id + '-' + n.id, source: n.parent_id, target: n.id, kind: 'composition' } });
    }
    (n.supporting_apps || []).forEach(a => {
      if (!appNodesSeen.has(a.id)) {
        appNodesSeen.add(a.id);
        const sub = appSubtype(a.type);
        elems.push({
          data: {
            id: 'app-' + a.id,
            label: a.name + '\n' + appShortType(a.type),
            kind: 'app',
            appType: a.type,
            appSub: sub,
          },
        });
      }
      elems.push({ data: { id: 'srv-' + n.id + '-' + a.id, source: 'app-' + a.id, target: n.id, kind: 'serving' } });
    });
  });

  const cy = cytoscape({
    container,
    elements: elems,
    style: [
      {
        selector: 'node',
        style: {
          shape: 'round-rectangle',
          label: 'data(label)',
          'text-valign': 'center',
          'text-halign': 'center',
          'font-size': '12px',
          'font-family': 'system-ui, sans-serif',
          color: '#e2e8f0',
          'text-wrap': 'wrap',
          'text-max-width': '150px',
          width: 160,
          height: 44,
          padding: '10px',
        },
      },
      { selector: 'node[kind="capability"]', style: { 'background-color': '#2a2010', 'border-color': '#e0af68', 'border-width': 2, color: '#e0af68' } },
      { selector: 'node[kind="app"][appSub="component"]',  style: { 'background-color': '#0d1f2e', 'border-color': '#7aa2f7', 'border-width': 2,   color: '#7aa2f7', width: 150, height: 44, 'font-size': '11px' } },
      { selector: 'node[kind="app"][appSub="service"]',    style: { 'background-color': '#0d1a28', 'border-color': '#4a9eff', 'border-width': 1.5, color: '#4a9eff', width: 140, height: 40, 'font-size': '11px', 'border-style': 'dashed' } },
      { selector: 'node[kind="app"][appSub="function"]',   style: { 'background-color': '#0d1520', 'border-color': '#3a5a80', 'border-width': 1,   color: '#5a80a8', width: 130, height: 38, 'font-size': '10px' } },
      { selector: 'node[kind="app"][appSub="interface"]',  style: { 'background-color': '#0a1e20', 'border-color': '#3a9a9a', 'border-width': 1.5, color: '#3a9a9a', width: 130, height: 38, 'font-size': '10px' } },
      { selector: 'node[kind="app"][appSub="other"]',      style: { 'background-color': '#111827', 'border-color': '#4b5563', 'border-width': 1,   color: '#6b7280', width: 130, height: 38, 'font-size': '10px' } },
      { selector: 'node:selected', style: { 'border-width': 3, 'border-color': '#fff' } },
      { selector: 'edge', style: { 'curve-style': 'taxi', 'taxi-direction': 'horizontal', width: 1.5, opacity: 0.6 } },
      { selector: 'edge[kind="composition"]', style: { 'line-color': '#a87c36', 'target-arrow-color': '#a87c36', 'target-arrow-shape': 'triangle' } },
      { selector: 'edge[kind="serving"]',     style: { 'line-color': '#4a6fa5', 'line-style': 'dashed', 'line-dash-pattern': [5, 3], 'target-arrow-color': '#4a6fa5', 'target-arrow-shape': 'triangle' } },
      { selector: '.faded', style: { opacity: 0.15 } },
    ],
    layout: { name: 'dagre', rankDir: 'LR', nodeSep: 30, rankSep: 100, align: 'UL', animate: false, padding: 40 },
    userZoomingEnabled: true,
    userPanningEnabled: true,
    boxSelectionEnabled: false,
  });

  return cy;
}
