import { describe, it, expect } from 'vitest';
import { VIEWS, LAYER_GROUPS } from './views.js';

describe('VIEWS', () => {
  it('has all expected view keys', () => {
    const keys = Object.keys(VIEWS);
    expect(keys).toContain('element-catalogue');
    expect(keys).toContain('capability-tree');
    expect(keys).toContain('application-dashboard');
    expect(keys).toContain('application-landscape');
    expect(keys).toContain('application-dependency');
    expect(keys).toContain('application-catalogue');
    expect(keys).toContain('technology-catalogue');
  });

  it('each view has label and icon', () => {
    for (const [key, view] of Object.entries(VIEWS)) {
      expect(view.label, `${key} missing label`).toBeTruthy();
      expect(view.icon, `${key} missing icon`).toBeTruthy();
    }
  });

  it('graph views have graph: true', () => {
    expect(VIEWS['application-dependency'].graph).toBe(true);
  });

  it('tree views have tree: true', () => {
    expect(VIEWS['capability-tree'].tree).toBe(true);
  });

  it('map views have map: true', () => {
    expect(VIEWS['application-landscape'].map).toBe(true);
  });

  it('dashboard views have dashboard: true', () => {
    expect(VIEWS['application-dashboard'].dashboard).toBe(true);
  });

  it('catalogue views have catalogue property', () => {
    expect(VIEWS['application-catalogue'].catalogue).toBe('application');
    expect(VIEWS['technology-catalogue'].catalogue).toBe('technology');
    expect(VIEWS['element-catalogue'].catalogue).toBe('element');
  });

  it('each view has a layer', () => {
    const validLayers = new Set(LAYER_GROUPS.map(g => g.key));
    for (const [key, view] of Object.entries(VIEWS)) {
      expect(validLayers.has(view.layer), `${key} has unknown layer: ${view.layer}`).toBe(true);
    }
  });
});

describe('LAYER_GROUPS', () => {
  it('has expected groups', () => {
    const keys = LAYER_GROUPS.map(g => g.key);
    expect(keys).toContain('cross');
    expect(keys).toContain('business');
    expect(keys).toContain('application');
    expect(keys).toContain('technology');
    expect(keys).toContain('catalogue');
  });

  it('each group has label and dot', () => {
    for (const group of LAYER_GROUPS) {
      expect(group.label).toBeTruthy();
      expect(group.dot).toBeTruthy();
    }
  });
});
