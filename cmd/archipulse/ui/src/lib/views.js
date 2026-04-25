export const VIEWS = {
  'element-catalogue':      { icon: '◈', label: 'Element Catalogue',        desc: 'All elements across layers',                                   layer: 'cross', catalogue: 'element' },
  'capability-tree':        { icon: '◈', label: 'Capability Tree',           desc: 'Business capability hierarchy',                                layer: 'business', tree: true },
  'capability-landscape':   { icon: '◈', label: 'Landscape by Capability',   desc: 'Applications mapped to realizing capabilities — hierarchical',  layer: 'application', map: true },
  'process-application':    { icon: '◈', label: 'Process–App Usage',         desc: 'Which applications support each business process',              layer: 'business', matrix: true },
  'application-dashboard':  { icon: '◉', label: 'Application Dashboard',     desc: 'Lifecycle status & type distribution charts',                  layer: 'application', dashboard: true },
  'application-landscape':  { icon: '◈', label: 'Landscape by Domain',       desc: 'Applications grouped by business domain',                      layer: 'application', map: true },
  'application-dependency': { icon: '◈', label: 'Dependency Graph',          desc: 'Interactive dependency graph',                                 layer: 'application', graph: true },
  'technology-stack':       { icon: '◈', label: 'Technology Stack',          desc: 'Applications mapped to the technology they run on',            layer: 'technology', matrix: true },
  'application-catalogue':  { icon: '◈', label: 'Application Catalogue',     desc: 'Application components with properties',                       layer: 'catalogue', catalogue: 'application' },
  'technology-catalogue':   { icon: '◈', label: 'Technology Catalogue',      desc: 'Infrastructure & technology elements',                         layer: 'catalogue', catalogue: 'technology' },
};

export const LAYER_GROUPS = [
  { key: 'cross',       label: 'Cross-cutting', dot: 'dot-cross' },
  { key: 'business',    label: 'Business',      dot: 'dot-biz'   },
  { key: 'application', label: 'Application',   dot: 'dot-app'   },
  { key: 'technology',  label: 'Technology',    dot: 'dot-tech'  },
  { key: 'catalogue',   label: 'Catalogues',    dot: 'dot-cross' },
];
