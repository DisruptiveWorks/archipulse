<script>
  import { onMount } from 'svelte';
  import { toast } from 'svelte-sonner';
  import { api } from '../lib/api.js';
  import { ICONS } from '../components/diagram/archimate-icons.js';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';

  export let params = {};
  $: wsId = params.wsId;

  // ── Tabs ──────────────────────────────────────────────────────────────────
  let activeTab = 'elements'; // 'elements' | 'relationships'

  // ── Element type / layer data ─────────────────────────────────────────────
  const ELEMENT_TYPES = {
    Motivation:    ['Stakeholder','Driver','Assessment','Goal','Outcome','Principle','Requirement','Constraint','Meaning','Value'],
    Strategy:      ['Resource','Capability','CourseOfAction','ValueStream'],
    Business:      ['BusinessActor','BusinessRole','BusinessCollaboration','BusinessInterface','BusinessProcess','BusinessFunction','BusinessInteraction','BusinessEvent','BusinessService','BusinessObject','Contract','Representation','Product'],
    Application:   ['ApplicationComponent','ApplicationCollaboration','ApplicationInterface','ApplicationFunction','ApplicationInteraction','ApplicationProcess','ApplicationEvent','ApplicationService','DataObject'],
    Technology:    ['Node','Device','SystemSoftware','TechnologyCollaboration','TechnologyInterface','Path','CommunicationNetwork','TechnologyFunction','TechnologyProcess','TechnologyInteraction','TechnologyEvent','TechnologyService','Artifact'],
    Physical:      ['Equipment','Facility','DistributionNetwork','Material'],
    'Implementation & Migration': ['WorkPackage','ImplementationEvent','Deliverable','ImplementationPlatform','Gap','Plateau'],
    Composite:     ['Grouping','Location'],
  };

  const TYPE_TO_LAYER = {};
  for (const [layer, types] of Object.entries(ELEMENT_TYPES)) {
    for (const t of types) TYPE_TO_LAYER[t] = layer;
  }

  const RELATIONSHIP_TYPES = [
    'AssociationRelationship','AccessRelationship','InfluenceRelationship',
    'RealizationRelationship','ServingRelationship','AssignmentRelationship',
    'AggregationRelationship','CompositionRelationship','FlowRelationship',
    'TriggeringRelationship','SpecializationRelationship',
  ];

  const LAYER_COLORS = {
    Application:   { bg: '#eff6ff', text: '#1d4ed8' },
    Business:      { bg: '#fffbeb', text: '#92400e' },
    Technology:    { bg: '#f0fdf4', text: '#166534' },
    Motivation:    { bg: '#f5f3ff', text: '#6d28d9' },
    Strategy:      { bg: '#f0fdfa', text: '#0f766e' },
    Physical:      { bg: '#fff7ed', text: '#9a3412' },
    'Implementation & Migration': { bg: '#f0f9ff', text: '#075985' },
    Composite:     { bg: '#f8fafc', text: '#475569' },
  };

  // ── Elements state ────────────────────────────────────────────────────────
  let elements    = [];
  let elLoading   = true;
  let elError     = null;
  let elSearch    = '';
  let activeLayer = '';   // '' = all

  $: elLayers = [...new Set(elements.map(e => e.layer))].sort();

  $: elFiltered = elements.filter(e => {
    if (activeLayer && e.layer !== activeLayer) return false;
    if (!elSearch) return true;
    const q = elSearch.toLowerCase();
    return e.name.toLowerCase().includes(q) || e.type.toLowerCase().includes(q);
  });

  // ── Relationships state ───────────────────────────────────────────────────
  let relationships  = [];
  let relLoading     = true;
  let relError       = null;
  let relSearch      = '';
  let activeRelType  = '';  // '' = all

  $: elementsById = Object.fromEntries(elements.map(e => [e.source_id, e]));

  $: relTypes = [...new Set(relationships.map(r => r.type))].sort();

  $: relFiltered = relationships.filter(r => {
    if (activeRelType && r.type !== activeRelType) return false;
    if (!relSearch) return true;
    const q = relSearch.toLowerCase();
    const src = elementsById[r.source_element]?.name ?? '';
    const tgt = elementsById[r.target_element]?.name ?? '';
    return r.type.toLowerCase().includes(q) || src.toLowerCase().includes(q) || tgt.toLowerCase().includes(q);
  });

  // ── Load data ─────────────────────────────────────────────────────────────
  async function loadElements() {
    elLoading = true; elError = null;
    try {
      const data = await api.get('/workspaces/' + wsId + '/elements?page=1&limit=500');
      elements = data.items ?? [];
    } catch (e) { elError = e.message; }
    finally { elLoading = false; }
  }

  async function loadRelationships() {
    relLoading = true; relError = null;
    try {
      const data = await api.get('/workspaces/' + wsId + '/relationships?page=1&limit=500');
      relationships = data.items ?? [];
    } catch (e) { relError = e.message; }
    finally { relLoading = false; }
  }

  onMount(() => {
    loadElements();
    loadRelationships();
  });

  // ── Slide-over state ──────────────────────────────────────────────────────
  let panelOpen    = false;
  let panelMode    = 'element'; // 'element' | 'relationship'
  let saving       = false;

  // Element form
  let elForm = { id: null, type: '', name: '', documentation: '', version: 0 };

  // Relationship form
  let relForm = { id: null, type: '', source_element: '', target_element: '', name: '', documentation: '', version: 0 };
  let srcSearch = '';
  let tgtSearch = '';
  let srcOpen   = false;
  let tgtOpen   = false;

  $: srcResults = srcSearch.length > 0
    ? elements.filter(e => e.name.toLowerCase().includes(srcSearch.toLowerCase())).slice(0, 8)
    : [];
  $: tgtResults = tgtSearch.length > 0
    ? elements.filter(e => e.name.toLowerCase().includes(tgtSearch.toLowerCase())).slice(0, 8)
    : [];

  $: selectedType = elForm.type ? TYPE_TO_LAYER[elForm.type] : null;

  function openNewElement() {
    elForm = { id: null, type: '', name: '', documentation: '', version: 0 };
    panelMode = 'element'; panelOpen = true;
  }

  function openEditElement(e) {
    elForm = { id: e.id, type: e.type, name: e.name, documentation: e.documentation ?? '', version: e.version ?? 0 };
    panelMode = 'element'; panelOpen = true;
  }

  function openNewRelationship() {
    relForm = { id: null, type: '', source_element: '', target_element: '', name: '', documentation: '', version: 0 };
    srcSearch = ''; tgtSearch = ''; srcOpen = false; tgtOpen = false;
    panelMode = 'relationship'; panelOpen = true;
  }

  function openEditRelationship(r) {
    const src = elementsById[r.source_element];
    const tgt = elementsById[r.target_element];
    const normalizedType = r.type.endsWith('Relationship') ? r.type : r.type + 'Relationship';
    relForm = { id: r.id, type: normalizedType, source_element: r.source_element, target_element: r.target_element, name: r.name ?? '', documentation: r.documentation ?? '', version: r.version ?? 0 };
    srcSearch = src?.name ?? ''; tgtSearch = tgt?.name ?? '';
    srcOpen = false; tgtOpen = false;
    panelMode = 'relationship'; panelOpen = true;
  }

  function closePanel() { panelOpen = false; }

  // ── Save element ──────────────────────────────────────────────────────────
  async function saveElement() {
    if (!elForm.type || !elForm.name.trim()) { toast.error('Type and name are required'); return; }
    saving = true;
    try {
      if (elForm.id) {
        const updated = await api.put('/workspaces/' + wsId + '/elements/' + elForm.id, {
          type: elForm.type, name: elForm.name.trim(),
          documentation: elForm.documentation, version: elForm.version,
        });
        elements = elements.map(e => e.id === updated.id ? updated : e);
        toast.success('Element updated');
      } else {
        const created = await api.post('/workspaces/' + wsId + '/elements', {
          type: elForm.type, name: elForm.name.trim(),
          documentation: elForm.documentation,
          source_id: 'id-' + crypto.randomUUID().slice(0, 8),
        });
        elements = [...elements, created];
        toast.success('Element created');
      }
      closePanel();
    } catch (e) { toast.error(e.message); }
    finally { saving = false; }
  }

  // ── Save relationship ─────────────────────────────────────────────────────
  async function saveRelationship() {
    if (!relForm.type || !relForm.source_element || !relForm.target_element) {
      toast.error('Type, source and target are required'); return;
    }
    saving = true;
    try {
      if (relForm.id) {
        const updated = await api.put('/workspaces/' + wsId + '/relationships/' + relForm.id, {
          type: relForm.type, source_element: relForm.source_element,
          target_element: relForm.target_element, name: relForm.name,
          documentation: relForm.documentation, version: relForm.version,
        });
        relationships = relationships.map(r => r.id === updated.id ? updated : r);
        toast.success('Relationship updated');
      } else {
        const created = await api.post('/workspaces/' + wsId + '/relationships', {
          type: relForm.type, source_element: relForm.source_element,
          target_element: relForm.target_element, name: relForm.name,
          documentation: relForm.documentation,
          source_id: 'id-' + crypto.randomUUID().slice(0, 8),
        });
        relationships = [...relationships, created];
        toast.success('Relationship created');
      }
      closePanel();
    } catch (e) { toast.error(e.message); }
    finally { saving = false; }
  }

  // ── Delete ────────────────────────────────────────────────────────────────
  let deleteTarget = null; // { kind: 'element'|'relationship', id, name }

  async function confirmDelete() {
    if (!deleteTarget) return;
    try {
      if (deleteTarget.kind === 'element') {
        await api.delete('/workspaces/' + wsId + '/elements/' + deleteTarget.id);
        elements = elements.filter(e => e.id !== deleteTarget.id);
        toast.success('Element deleted');
      } else {
        await api.delete('/workspaces/' + wsId + '/relationships/' + deleteTarget.id);
        relationships = relationships.filter(r => r.id !== deleteTarget.id);
        toast.success('Relationship deleted');
      }
    } catch (e) { toast.error(e.message); }
    finally { deleteTarget = null; }
  }

  // ── Helpers ───────────────────────────────────────────────────────────────
  function shortType(t) {
    return t.replace(/Relationship$/, '').replace(/([A-Z])/g, ' $1').trim() || t;
  }

  function layerColor(layer) {
    return LAYER_COLORS[layer] ?? LAYER_COLORS['Composite'];
  }

  function relConnectorSvg(rawType) {
    // Normalize: both 'Composition' and 'CompositionRelationship' are stored in the DB
    const type = rawType.endsWith('Relationship') ? rawType : rawType + 'Relationship';

    const solid  = (x1, x2) => `<line x1="${x1}" y1="8" x2="${x2}" y2="8" stroke="currentColor" stroke-width="1.5"/>`;
    const dashed = (x1, x2) => `<line x1="${x1}" y1="8" x2="${x2}" y2="8" stroke="currentColor" stroke-width="1.5" stroke-dasharray="4,3"/>`;
    // open chevron arrowhead
    const openArr  = `<polyline points="66,4 74,8 66,12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/>`;
    // filled triangle arrowhead
    const filledArr= `<polygon points="66,4 76,8 66,12" fill="currentColor" stroke="none"/>`;
    // hollow triangle (Realization / Specialization)
    const hollowTri= `<polygon points="66,4 76,8 66,12" fill="var(--color-background,#fff)" stroke="currentColor" stroke-width="1.5"/>`;
    // diamonds at source-side (x=2–14)
    const hollowDia= `<polygon points="2,8 8,3 14,8 8,13" fill="var(--color-background,#fff)" stroke="currentColor" stroke-width="1.5"/>`;
    const filledDia= `<polygon points="2,8 8,3 14,8 8,13" fill="currentColor" stroke="currentColor" stroke-width="1"/>`;
    // circle for Assignment
    const dot      = `<circle cx="6" cy="8" r="4" fill="currentColor" stroke="currentColor" stroke-width="1"/>`;
    switch (type) {
      case 'AssociationRelationship':    return solid(2, 70)  + openArr;
      case 'AccessRelationship':         return dashed(2, 70) + openArr;
      case 'InfluenceRelationship':      return dashed(2, 70) + openArr;
      case 'RealizationRelationship':    return dashed(2, 64) + hollowTri;
      case 'ServingRelationship':        return solid(2, 70)  + openArr;
      case 'AssignmentRelationship':     return dot + solid(10, 70) + filledArr;
      case 'AggregationRelationship':    return hollowDia + solid(14, 76);
      case 'CompositionRelationship':    return filledDia + solid(14, 76);
      case 'FlowRelationship':           return solid(2, 70)  + filledArr;
      case 'TriggeringRelationship':     return solid(2, 70)  + filledArr;
      case 'SpecializationRelationship': return solid(2, 64)  + hollowTri;
      default:                           return solid(2, 70)  + openArr;
    }
  }
</script>

<!-- ── Layout ── -->
<div class="content flex flex-col h-full min-h-0">

  <!-- Tabs -->
  <div class="flex items-center gap-1 mb-4 border-b border-border">
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors -mb-px border-b-2 {activeTab === 'elements' ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'}"
      onclick={() => activeTab = 'elements'}
    >Elements</button>
    <button
      class="px-4 py-2 text-[13px] font-medium transition-colors -mb-px border-b-2 {activeTab === 'relationships' ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'}"
      onclick={() => activeTab = 'relationships'}
    >Relationships</button>
  </div>

  <!-- ── Elements tab ── -->
  {#if activeTab === 'elements'}
    {#if elLoading}
      <div class="flex items-center gap-2 text-muted-foreground py-6">
        <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
        Loading…
      </div>
    {:else if elError}
      <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">Error: {elError}</div>
    {:else}
      <div class="flex gap-4 min-h-0 flex-1">

        <!-- Layer sidebar -->
        <div class="w-44 flex-shrink-0">
          <div class="text-[10px] font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">Layer</div>
          <button
            class="w-full text-left px-2 py-1.5 rounded text-[13px] transition-colors mb-0.5 {activeLayer === '' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted'}"
            onclick={() => activeLayer = ''}
          >All <span class="text-[11px] ml-1 opacity-60">({elements.length})</span></button>
          {#each elLayers as layer}
            {@const c = layerColor(layer)}
            {@const count = elements.filter(e => e.layer === layer).length}
            <button
              class="w-full text-left px-2 py-1.5 rounded text-[13px] transition-colors mb-0.5 {activeLayer === layer ? 'font-medium' : 'text-muted-foreground hover:bg-muted'}"
              style={activeLayer === layer ? `background:${c.bg}; color:${c.text}` : ''}
              onclick={() => activeLayer = activeLayer === layer ? '' : layer}
            >{layer} <span class="text-[11px] ml-1 opacity-60">({count})</span></button>
          {/each}
        </div>

        <!-- Main panel -->
        <div class="flex-1 min-w-0 flex flex-col">
          <!-- Toolbar -->
          <div class="flex items-center gap-3 mb-3">
            <input type="search" bind:value={elSearch} placeholder="Search name or type…"
              class="bg-background border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary flex-1 max-w-xs" />
            <span class="text-[12px] text-muted-foreground ml-auto">{elFiltered.length} of {elements.length}</span>
            <button onclick={openNewElement}
              class="bg-primary text-primary-foreground px-3 py-1.5 rounded-md text-[13px] font-medium hover:bg-primary/90 transition-colors flex items-center gap-1.5">
              + Add Element
            </button>
          </div>

          {#if elFiltered.length === 0}
            <div class="text-center py-16 text-muted-foreground">
              <div class="text-[36px] mb-3">📭</div>
              <p class="text-[14px]">{elements.length === 0 ? 'No elements — import a model or add one.' : 'No results match your filters.'}</p>
            </div>
          {:else}
            <div class="overflow-x-auto border border-border rounded-lg flex-1">
              <table class="w-full text-[13px]">
                <thead>
                  <tr class="border-b border-border bg-muted">
                    <th class="w-8 px-2 py-2.5"></th>
                    <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold">Type</th>
                    <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold">Name</th>
                    <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold">Documentation</th>
                    <th class="px-3 py-2.5 w-20"></th>
                  </tr>
                </thead>
                <tbody>
                  {#each elFiltered as el}
                    {@const c = layerColor(el.layer)}
                    {@const icon = ICONS[el.type]}
                    <tr class="border-b border-border hover:bg-muted/40 transition-colors">
                      <td class="px-2 py-2 text-center">
                        {#if icon}
                          <svg viewBox="0 0 16 16" width="16" height="16" style="stroke:{c.text};fill:none;" class="inline-block">
                            {@html icon}
                          </svg>
                        {/if}
                      </td>
                      <td class="px-3 py-2 whitespace-nowrap">
                        <span class="text-[11px] px-1.5 py-0.5 rounded font-medium" style="background:{c.bg};color:{c.text}">{el.type}</span>
                      </td>
                      <td class="px-3 py-2 font-medium text-foreground">{el.name}</td>
                      <td class="px-3 py-2 text-muted-foreground max-w-xs truncate" title={el.documentation}>{el.documentation || '—'}</td>
                      <td class="px-3 py-2 whitespace-nowrap">
                        <div class="flex items-center gap-1 justify-end">
                          <button onclick={() => openEditElement(el)}
                            class="p-1 rounded hover:bg-muted text-muted-foreground hover:text-foreground transition-colors" title="Edit">✎</button>
                          <button onclick={() => deleteTarget = { kind: 'element', id: el.id, name: el.name }}
                            class="p-1 rounded hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors" title="Delete">⌫</button>
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </div>
      </div>
    {/if}
  {/if}

  <!-- ── Relationships tab ── -->
  {#if activeTab === 'relationships'}
    {#if relLoading || elLoading}
      <div class="flex items-center gap-2 text-muted-foreground py-6">
        <div class="size-4 rounded-full border-2 border-border border-t-primary animate-spin flex-shrink-0"></div>
        Loading…
      </div>
    {:else if relError}
      <div class="text-sm text-destructive bg-destructive/10 border border-destructive/30 rounded-md px-3 py-2">Error: {relError}</div>
    {:else}
      <div class="flex gap-4 min-h-0 flex-1">

        <!-- Type sidebar -->
        <div class="w-44 flex-shrink-0">
          <div class="text-[10px] font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">Type</div>
          <button
            class="w-full text-left px-2 py-1.5 rounded text-[13px] transition-colors mb-0.5 {activeRelType === '' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted'}"
            onclick={() => activeRelType = ''}
          >All <span class="text-[11px] ml-1 opacity-60">({relationships.length})</span></button>
          {#each relTypes as t}
            {@const count = relationships.filter(r => r.type === t).length}
            <button
              class="w-full text-left px-2 py-1.5 rounded text-[13px] transition-colors mb-0.5 {activeRelType === t ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted'}"
              onclick={() => activeRelType = activeRelType === t ? '' : t}
            >{shortType(t)} <span class="text-[11px] ml-1 opacity-60">({count})</span></button>
          {/each}
        </div>

        <!-- Main panel -->
        <div class="flex-1 min-w-0 flex flex-col">
          <div class="flex items-center gap-3 mb-3">
            <input type="search" bind:value={relSearch} placeholder="Search type, source, or target…"
              class="bg-background border border-border rounded-md px-3 py-1.5 text-[13px] text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary flex-1 max-w-xs" />
            <span class="text-[12px] text-muted-foreground ml-auto">{relFiltered.length} of {relationships.length}</span>
            <button onclick={openNewRelationship}
              class="bg-primary text-primary-foreground px-3 py-1.5 rounded-md text-[13px] font-medium hover:bg-primary/90 transition-colors">
              + Add Relationship
            </button>
          </div>

          {#if relFiltered.length === 0}
            <div class="text-center py-16 text-muted-foreground">
              <div class="text-[36px] mb-3">📭</div>
              <p class="text-[14px]">{relationships.length === 0 ? 'No relationships yet.' : 'No results match your filters.'}</p>
            </div>
          {:else}
            <div class="overflow-x-auto border border-border rounded-lg flex-1">
              <table class="w-full text-[13px]">
                <thead>
                  <tr class="border-b border-border bg-muted">
                    <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold">Source</th>
                    <th class="px-4 py-2.5"></th>
                    <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold">Target</th>
                    <th class="text-left px-3 py-2.5 text-muted-foreground font-semibold">Name</th>
                    <th class="px-3 py-2.5 w-16"></th>
                  </tr>
                </thead>
                <tbody>
                  {#each relFiltered as rel}
                    {@const src = elementsById[rel.source_element]}
                    {@const tgt = elementsById[rel.target_element]}
                    {@const srcC = layerColor(src?.layer ?? 'Composite')}
                    {@const tgtC = layerColor(tgt?.layer ?? 'Composite')}
                    <tr class="border-b border-border hover:bg-muted/40 transition-colors">
                      <!-- Source element -->
                      <td class="px-3 py-2 max-w-[180px]">
                        <div class="flex items-center gap-1.5 min-w-0">
                          {#if src && ICONS[src.type]}
                            <svg viewBox="0 0 16 16" width="14" height="14" style="stroke:{srcC.text};fill:none;flex-shrink:0">
                              {@html ICONS[src.type]}
                            </svg>
                          {/if}
                          <span class="font-medium text-foreground truncate" title={src?.name}>{src?.name ?? rel.source_element}</span>
                        </div>
                        {#if src}
                          <div class="text-[10px] mt-0.5 ml-[18px]" style="color:{srcC.text}">{src.type}</div>
                        {/if}
                      </td>
                      <!-- Relationship connector -->
                      <td class="px-2 py-2 whitespace-nowrap">
                        <div class="flex flex-col items-center gap-0.5">
                          <svg viewBox="0 0 80 16" width="80" height="16" class="text-muted-foreground flex-shrink-0">
                            {@html relConnectorSvg(rel.type)}
                          </svg>
                          <span class="text-[9px] text-muted-foreground leading-none">{shortType(rel.type)}</span>
                        </div>
                      </td>
                      <!-- Target element -->
                      <td class="px-3 py-2 max-w-[180px]">
                        <div class="flex items-center gap-1.5 min-w-0">
                          {#if tgt && ICONS[tgt.type]}
                            <svg viewBox="0 0 16 16" width="14" height="14" style="stroke:{tgtC.text};fill:none;flex-shrink:0">
                              {@html ICONS[tgt.type]}
                            </svg>
                          {/if}
                          <span class="font-medium text-foreground truncate" title={tgt?.name}>{tgt?.name ?? rel.target_element}</span>
                        </div>
                        {#if tgt}
                          <div class="text-[10px] mt-0.5 ml-[18px]" style="color:{tgtC.text}">{tgt.type}</div>
                        {/if}
                      </td>
                      <td class="px-3 py-2 text-muted-foreground">{rel.name || '—'}</td>
                      <td class="px-3 py-2 whitespace-nowrap">
                        <div class="flex items-center gap-1 justify-end">
                          <button onclick={() => openEditRelationship(rel)}
                            class="p-1 rounded hover:bg-muted text-muted-foreground hover:text-foreground transition-colors" title="Edit">✎</button>
                          <button onclick={() => deleteTarget = { kind: 'relationship', id: rel.id, name: rel.type }}
                            class="p-1 rounded hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors" title="Delete">⌫</button>
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </div>
      </div>
    {/if}
  {/if}
</div>

<!-- ── Slide-over panel ── -->
{#if panelOpen}
  <!-- Backdrop -->
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/20 z-40" onclick={closePanel}></div>

  <!-- Panel -->
  <div class="fixed top-0 right-0 h-full w-[400px] bg-background border-l border-border shadow-xl z-50 flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between px-5 py-4 border-b border-border">
      <h2 class="text-[15px] font-semibold">
        {#if panelMode === 'element'}
          {elForm.id ? 'Edit Element' : 'New Element'}
        {:else}
          {relForm.id ? 'Edit Relationship' : 'New Relationship'}
        {/if}
      </h2>
      <button onclick={closePanel} class="text-muted-foreground hover:text-foreground text-lg leading-none p-1">✕</button>
    </div>

    <!-- Form -->
    <div class="flex-1 overflow-y-auto px-5 py-5 space-y-5">

      {#if panelMode === 'element'}
        <!-- Type selector -->
        <div>
          <label for="el-type" class="block text-[12px] font-semibold text-muted-foreground mb-1.5 uppercase tracking-wide">Type *</label>
          <select id="el-type" bind:value={elForm.type}
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary">
            <option value="">Select a type…</option>
            {#each Object.entries(ELEMENT_TYPES) as [layer, types]}
              <optgroup label={layer}>
                {#each types as t}
                  <option value={t}>{t}</option>
                {/each}
              </optgroup>
            {/each}
          </select>

          <!-- SVG preview card -->
          {#if elForm.type && ICONS[elForm.type]}
            {@const c = layerColor(TYPE_TO_LAYER[elForm.type])}
            <div class="mt-2.5 flex items-center gap-3 px-3 py-2.5 rounded-lg border border-border"
                 style="background:{c.bg}">
              <svg viewBox="0 0 16 16" width="28" height="28" style="stroke:{c.text};fill:none;flex-shrink:0">
                {@html ICONS[elForm.type]}
              </svg>
              <div>
                <div class="text-[13px] font-semibold" style="color:{c.text}">{elForm.type}</div>
                <div class="text-[11px] opacity-70" style="color:{c.text}">{TYPE_TO_LAYER[elForm.type]} layer</div>
              </div>
            </div>
          {/if}
        </div>

        <!-- Name -->
        <div>
          <label for="el-name" class="block text-[12px] font-semibold text-muted-foreground mb-1.5 uppercase tracking-wide">Name *</label>
          <input id="el-name" type="text" bind:value={elForm.name} placeholder="e.g. CRM System"
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
        </div>

        <!-- Documentation -->
        <div>
          <label for="el-doc" class="block text-[12px] font-semibold text-muted-foreground mb-1.5 uppercase tracking-wide">Documentation</label>
          <textarea id="el-doc" bind:value={elForm.documentation} rows="4" placeholder="Optional description…"
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary resize-none"></textarea>
        </div>

      {:else}
        <!-- Relationship type -->
        <div>
          <label for="rel-type" class="block text-[12px] font-semibold text-muted-foreground mb-1.5 uppercase tracking-wide">Type *</label>
          <select id="rel-type" bind:value={relForm.type}
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary">
            <option value="">Select a type…</option>
            {#each RELATIONSHIP_TYPES as t}
              <option value={t}>{shortType(t)}</option>
            {/each}
          </select>
        </div>

        <!-- Source element search -->
        <div class="relative">
          <label for="rel-src" class="block text-[12px] font-semibold text-muted-foreground mb-1.5 uppercase tracking-wide">Source Element *</label>
          <input id="rel-src" type="text" bind:value={srcSearch}
            onfocus={() => srcOpen = true}
            oninput={() => { srcOpen = true; relForm.source_element = ''; }}
            placeholder="Search element…"
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
          {#if srcOpen && srcResults.length > 0}
            <div class="absolute left-0 right-0 top-full mt-1 bg-background border border-border rounded-md shadow-lg z-10 max-h-48 overflow-y-auto">
              {#each srcResults as e}
                {@const c = layerColor(e.layer)}
                <button class="w-full text-left px-3 py-2 text-[13px] hover:bg-muted flex items-center gap-2"
                  onmousedown={() => { relForm.source_element = e.source_id; srcSearch = e.name; srcOpen = false; }}>
                  <span class="text-[10px] px-1.5 py-0.5 rounded flex-shrink-0" style="background:{c.bg};color:{c.text}">{e.type}</span>
                  <span class="truncate">{e.name}</span>
                </button>
              {/each}
            </div>
          {/if}
          {#if relForm.source_element}
            <div class="mt-1 text-[11px] text-muted-foreground">✓ {srcSearch}</div>
          {/if}
        </div>

        <!-- Target element search -->
        <div class="relative">
          <label for="rel-tgt" class="block text-[12px] font-semibold text-muted-foreground mb-1.5 uppercase tracking-wide">Target Element *</label>
          <input id="rel-tgt" type="text" bind:value={tgtSearch}
            onfocus={() => tgtOpen = true}
            oninput={() => { tgtOpen = true; relForm.target_element = ''; }}
            placeholder="Search element…"
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
          {#if tgtOpen && tgtResults.length > 0}
            <div class="absolute left-0 right-0 top-full mt-1 bg-background border border-border rounded-md shadow-lg z-10 max-h-48 overflow-y-auto">
              {#each tgtResults as e}
                {@const c = layerColor(e.layer)}
                <button class="w-full text-left px-3 py-2 text-[13px] hover:bg-muted flex items-center gap-2"
                  onmousedown={() => { relForm.target_element = e.source_id; tgtSearch = e.name; tgtOpen = false; }}>
                  <span class="text-[10px] px-1.5 py-0.5 rounded flex-shrink-0" style="background:{c.bg};color:{c.text}">{e.type}</span>
                  <span class="truncate">{e.name}</span>
                </button>
              {/each}
            </div>
          {/if}
          {#if relForm.target_element}
            <div class="mt-1 text-[11px] text-muted-foreground">✓ {tgtSearch}</div>
          {/if}
        </div>

        <!-- Name -->
        <div>
          <label for="rel-name" class="block text-[12px] font-semibold text-muted-foreground mb-1.5 uppercase tracking-wide">Name <span class="normal-case font-normal">(optional)</span></label>
          <input id="rel-name" type="text" bind:value={relForm.name} placeholder="e.g. provides data to"
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
        </div>

        <!-- Documentation -->
        <div>
          <label for="rel-doc" class="block text-[12px] font-semibold text-muted-foreground mb-1.5 uppercase tracking-wide">Documentation</label>
          <textarea id="rel-doc" bind:value={relForm.documentation} rows="3" placeholder="Optional description…"
            class="w-full bg-background border border-border rounded-md px-3 py-2 text-[13px] text-foreground focus:outline-none focus:ring-1 focus:ring-primary resize-none"></textarea>
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <div class="px-5 py-4 border-t border-border flex items-center justify-end gap-2">
      <button onclick={closePanel} disabled={saving}
        class="px-4 py-2 rounded-md text-[13px] border border-border hover:bg-muted transition-colors">
        Cancel
      </button>
      <button onclick={panelMode === 'element' ? saveElement : saveRelationship} disabled={saving}
        class="px-4 py-2 rounded-md text-[13px] font-medium bg-primary text-primary-foreground hover:bg-primary/90 transition-colors flex items-center gap-1.5 disabled:opacity-60">
        {#if saving}
          <span class="size-3.5 rounded-full border-2 border-white/40 border-t-white animate-spin"></span>
        {/if}
        {panelMode === 'element' ? (elForm.id ? 'Save Changes' : 'Create Element') : (relForm.id ? 'Save Changes' : 'Create Relationship')}
      </button>
    </div>
  </div>
{/if}

<!-- ── Delete confirmation dialog ── -->
<Dialog.Root open={!!deleteTarget} onOpenChange={(o) => { if (!o) deleteTarget = null; }}>
  <Dialog.Content class="max-w-sm">
    <Dialog.Header>
      <Dialog.Title>Delete {deleteTarget?.kind}?</Dialog.Title>
      <Dialog.Description>
        <strong>{deleteTarget?.name}</strong> will be permanently removed. This cannot be undone.
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer class="gap-2">
      <Button variant="outline" onclick={() => deleteTarget = null}>Cancel</Button>
      <Button variant="destructive" onclick={confirmDelete}>Delete</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
