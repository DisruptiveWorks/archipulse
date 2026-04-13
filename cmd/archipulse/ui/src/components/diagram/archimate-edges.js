// ArchiMate relationship visual styles.
// Each entry defines the stroke, line style, and marker types for source/end.
//
// Marker types:
//   filled-arrow   — solid filled arrowhead (Triggering, Flow, Assignment)
//   open-arrow     — open V arrowhead (Association, Serving, Influence)
//   open-triangle  — hollow closed triangle (Realization, Specialization)
//   filled-diamond — solid diamond at source (Composition)
//   open-diamond   — hollow diamond at source (Aggregation)
//   filled-circle  — solid circle at source (Assignment)
//
// Access relationship markers depend on accessType:
//   Access    → open-arrow at end only
//   Read      → open-arrow at start (towards source = data flows from target to source)
//   Write     → open-arrow at end  (towards target = data flows from source to target)
//   ReadWrite → open-arrow at both ends

export const RELATIONSHIP_STYLES = {
  TriggeringRelationship:     { stroke: '#374151', width: 1.5, dash: null,  end: 'filled-arrow',  start: null },
  FlowRelationship:           { stroke: '#374151', width: 1.5, dash: '8 4', end: 'filled-arrow',  start: null },
  AssociationRelationship:    { stroke: '#9CA3AF', width: 1.2, dash: null,  end: null,            start: null },
  CompositionRelationship:    { stroke: '#374151', width: 1.5, dash: null,  end: null,            start: 'filled-diamond' },
  AggregationRelationship:    { stroke: '#374151', width: 1.5, dash: null,  end: null,            start: 'open-diamond' },
  AssignmentRelationship:     { stroke: '#374151', width: 1.5, dash: null,  end: 'filled-arrow',  start: 'filled-circle' },
  RealizationRelationship:    { stroke: '#6B7280', width: 1.5, dash: '8 4', end: 'open-triangle', start: null },
  RealisationRelationship:    { stroke: '#6B7280', width: 1.5, dash: '8 4', end: 'open-triangle', start: null },
  ServingRelationship:        { stroke: '#374151', width: 1.2, dash: null,  end: 'open-arrow',    start: null },
  AccessRelationship:         { stroke: '#9CA3AF', width: 1.2, dash: '4 3', end: 'open-arrow',    start: null },
  InfluenceRelationship:      { stroke: '#9CA3AF', width: 1.2, dash: '6 3', end: 'open-arrow',    start: null },
  SpecializationRelationship: { stroke: '#374151', width: 1.5, dash: null,  end: 'open-triangle', start: null },
};

const DEFAULT_STYLE = { stroke: '#6B7280', width: 1.2, dash: null, end: 'filled-arrow', start: null };

/**
 * Returns the visual style for a relationship, taking into account type-specific
 * semantic attributes from the OEF standard:
 *   - accessType (Access): Access | Read | Write | ReadWrite
 *   - isDirected (Association): true adds an arrowhead at the target end
 */
export function getRelationshipStyle(relType, { accessType, isDirected } = {}) {
  let base;
  if (!relType) {
    base = { ...DEFAULT_STYLE };
  } else if (RELATIONSHIP_STYLES[relType]) {
    base = { ...RELATIONSHIP_STYLES[relType] };
  } else {
    // Partial match for truncated type names
    const found = Object.entries(RELATIONSHIP_STYLES).find(([key]) =>
      relType.includes(key.replace('Relationship', ''))
    );
    base = found ? { ...found[1] } : { ...DEFAULT_STYLE };
  }

  // Access relationship: markers depend on accessType.
  if (relType === 'AccessRelationship' || relType === 'Access') {
    switch (accessType) {
      case 'Read':
        base.start = 'open-arrow-rev';
        base.end   = null;
        break;
      case 'Write':
        base.start = null;
        base.end   = 'open-arrow';
        break;
      case 'ReadWrite':
        base.start = 'open-arrow-rev';
        base.end   = 'open-arrow';
        break;
      default: // 'Access' or unset → arrow at end
        base.start = null;
        base.end   = 'open-arrow';
    }
  }

  // Association relationship: directed flag adds an arrowhead at the target.
  if (
    (relType === 'AssociationRelationship' || relType === 'Association') &&
    isDirected
  ) {
    base.end = 'open-arrow';
  }

  return base;
}

// Compute where the edge exits/enters a node boundary toward point (px, py).
//
// orthogonal=true: bendpoint-driven routing — snap the exit/entry coordinate
//   to the bendpoint lane so segments remain horizontal or vertical.
// orthogonal=false: direct connection — use a diagonal ray from center so the
//   edge exits through the most natural point on the border.
export function intersectNodeBoundary(bounds, px, py, orthogonal = false) {
  if (!bounds) return { x: px, y: py };
  const cx = bounds.x + bounds.w / 2;
  const cy = bounds.y + bounds.h / 2;
  const dx = px - cx;
  const dy = py - cy;
  if (Math.abs(dx) < 0.1 && Math.abs(dy) < 0.1) return { x: cx, y: cy };
  const hw = bounds.w / 2;
  const hh = bounds.h / 2;
  const left   = bounds.x;
  const right  = bounds.x + bounds.w;
  const top    = bounds.y;
  const bottom = bounds.y + bounds.h;
  const sx = dx !== 0 ? hw / Math.abs(dx) : Infinity;
  const sy = dy !== 0 ? hh / Math.abs(dy) : Infinity;

  if (orthogonal) {
    if (sy <= sx) {
      return {
        x: +(Math.max(left, Math.min(right, px))).toFixed(2),
        y: dy < 0 ? top : bottom,
      };
    } else {
      return {
        x: dx < 0 ? left : right,
        y: +(Math.max(top, Math.min(bottom, py))).toFixed(2),
      };
    }
  } else {
    const s = Math.min(sx, sy);
    return {
      x: +(cx + dx * s).toFixed(2),
      y: +(cy + dy * s).toFixed(2),
    };
  }
}
