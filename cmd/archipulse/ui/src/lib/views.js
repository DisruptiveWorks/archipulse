export const VIEWS = {
  'element-catalogue':      { icon: '◈', label: 'Element Catalogue',      desc: 'All elements across layers',                                   layer: 'cross' },
  'capability-tree':        { icon: '◈', label: 'Capability Tree',         desc: 'Business capability hierarchy',                                layer: 'business', tree: true },
  'application-dashboard':  { icon: '◉', label: 'Application Dashboard',   desc: 'Lifecycle status & type distribution charts',                  layer: 'application', dashboard: true },
  'application-catalogue':  { icon: '◈', label: 'Application Catalogue',   desc: 'Application components',                                       layer: 'application' },
  'application-landscape':  { icon: '◈', label: 'Application Landscape',   desc: 'Apps mapped to business processes',                            layer: 'application' },
  'application-dependency': { icon: '◈', label: 'Dependency Graph',        desc: 'Interactive dependency graph',                                 layer: 'application', graph: true },
  'integration-map':        { icon: '⇄', label: 'Integration Map',         desc: 'Integration topology — services, components and data flows',   layer: 'application', graph: true },
  'technology-catalogue':   { icon: '◈', label: 'Technology Catalogue',    desc: 'Infrastructure & technology elements',                         layer: 'technology' },
};

export const LAYER_GROUPS = [
  { key: 'cross',       label: 'Cross-cutting', dot: 'dot-cross' },
  { key: 'business',    label: 'Business',      dot: 'dot-biz'   },
  { key: 'application', label: 'Application',   dot: 'dot-app'   },
  { key: 'technology',  label: 'Technology',    dot: 'dot-tech'  },
];
