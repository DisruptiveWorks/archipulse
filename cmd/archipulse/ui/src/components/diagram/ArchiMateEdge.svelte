<script>
  // Custom XY Flow edge that renders ArchiMate relationship types faithfully.
  //
  // Instead of using XY Flow's auto-routed handle positions, this component
  // computes the path directly from the source/target node bounds and bendpoints
  // stored in `data`. Markers (arrows, diamonds, circles) are defined globally
  // in DiagramView via a hidden <svg><defs> and referenced by ID.
  //
  // Props passed by XY Flow: id, sourceX, sourceY, targetX, targetY, data, ...
  // We only use `data`.

  import { getRelationshipStyle, intersectNodeBoundary } from './archimate-edges.js';

  export let data;

  $: style     = getRelationshipStyle(data?.relationshipType);
  $: bps       = data?.bendpoints || [];
  $: srcBounds = data?.sourceBounds;
  $: tgtBounds = data?.targetBounds;

  $: srcCenter = srcBounds
    ? { x: srcBounds.x + srcBounds.w / 2, y: srcBounds.y + srcBounds.h / 2 }
    : { x: 0, y: 0 };
  $: tgtCenter = tgtBounds
    ? { x: tgtBounds.x + tgtBounds.w / 2, y: tgtBounds.y + tgtBounds.h / 2 }
    : { x: 0, y: 0 };

  // Use orthogonal snapping only when bendpoints define a routing lane.
  // Direct connections (no bendpoints) use natural diagonal intersection.
  $: hasBps = bps.length > 0;

  $: startPt = intersectNodeBoundary(
    srcBounds,
    hasBps ? bps[0].x : tgtCenter.x,
    hasBps ? bps[0].y : tgtCenter.y,
    hasBps,
  );
  $: endPt = intersectNodeBoundary(
    tgtBounds,
    hasBps ? bps[bps.length - 1].x : srcCenter.x,
    hasBps ? bps[bps.length - 1].y : srcCenter.y,
    hasBps,
  );

  $: allPts = [startPt, ...bps, endPt];

  // Build a smooth polyline: straight segments but with small rounded corners
  // at intermediate bendpoints so the path looks clean rather than jagged.
  function smoothPath(pts) {
    if (pts.length < 2) return '';
    if (pts.length === 2) return `M${pts[0].x},${pts[0].y} L${pts[1].x},${pts[1].y}`;
    const r = 8; // corner radius in diagram units
    let d = `M${pts[0].x},${pts[0].y}`;
    for (let i = 1; i < pts.length - 1; i++) {
      const prev = pts[i - 1];
      const cur  = pts[i];
      const next = pts[i + 1];
      const d1 = Math.hypot(cur.x - prev.x, cur.y - prev.y);
      const d2 = Math.hypot(next.x - cur.x, next.y - cur.y);
      const t1 = Math.min(r, d1 / 2) / d1;
      const t2 = Math.min(r, d2 / 2) / d2;
      const bx = cur.x - (cur.x - prev.x) * t1;
      const by = cur.y - (cur.y - prev.y) * t1;
      const cx2 = cur.x + (next.x - cur.x) * t2;
      const cy2 = cur.y + (next.y - cur.y) * t2;
      d += ` L${bx.toFixed(1)},${by.toFixed(1)} Q${cur.x},${cur.y} ${cx2.toFixed(1)},${cy2.toFixed(1)}`;
    }
    d += ` L${pts[pts.length - 1].x},${pts[pts.length - 1].y}`;
    return d;
  }

  $: pathD = smoothPath(allPts);

  $: markerEnd   = style.end   ? `url(#am-${style.end})`   : undefined;
  $: markerStart = style.start ? `url(#am-${style.start})` : undefined;
</script>

<!-- Invisible wider stroke for future hover/selection interactions -->
<path d={pathD} fill="none" stroke="transparent" stroke-width="14" />

<!-- Main edge path — stroke color propagates to markers via context-stroke -->
<path
  d={pathD}
  fill="none"
  stroke={style.stroke}
  stroke-width={style.width}
  stroke-dasharray={style.dash || undefined}
  marker-end={markerEnd}
  marker-start={markerStart}
/>
