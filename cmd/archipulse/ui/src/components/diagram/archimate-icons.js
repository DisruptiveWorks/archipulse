// ArchiMate 3 element type icons — inline SVG paths (viewBox 0 0 16 16)
// Each entry is an SVG string to embed inside <svg viewBox="0 0 16 16">
//
// Shape vocabulary (per ArchiMate 3 standard):
//   Actor      → stick figure
//   Role       → cylinder (standing)
//   Collaboration → two overlapping circles
//   Interface  → lollipop (circle + stem + T-bar)
//   Process    → right-pointing pentagon (flat left, pointed right)
//   Function   → left-notch chevron (pointed left, flat right)
//   Interaction → two overlapping circles (same as collab — context differentiates)
//   Service    → oval / stadium
//   Event      → trigger pentagon (flat right, notch-left)
//   Object/Data → rectangle with lines
//   Component  → rect with two plug tabs on left
//   Node       → 3D box
//   Cylinder   → vertical cylinder (Role, Stakeholder, SystemSoftware)
//   Grouping   → dashed rect

export const ICONS = {

  // ── Business layer ───────────────────────────────────────────────────────

  // Stick figure
  BusinessActor: `
    <circle cx="8" cy="3.5" r="2" stroke-width="1.2" fill="none"/>
    <line x1="8" y1="5.5" x2="8" y2="10.5" stroke-width="1.2"/>
    <line x1="4.5" y1="7.5" x2="11.5" y2="7.5" stroke-width="1.2"/>
    <line x1="8" y1="10.5" x2="5.5" y2="14" stroke-width="1.2"/>
    <line x1="8" y1="10.5" x2="10.5" y2="14" stroke-width="1.2"/>
  `,

  // Vertical cylinder
  BusinessRole: `
    <ellipse cx="8" cy="4.5" rx="5" ry="1.8" stroke-width="1.2" fill="none"/>
    <line x1="3" y1="4.5" x2="3" y2="12" stroke-width="1.2"/>
    <line x1="13" y1="4.5" x2="13" y2="12" stroke-width="1.2"/>
    <ellipse cx="8" cy="12" rx="5" ry="1.8" stroke-width="1.2" fill="none"/>
  `,

  // Two overlapping circles
  BusinessCollaboration: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Lollipop — circle + stem + T-bar
  BusinessInterface: `
    <circle cx="8" cy="5" r="3" stroke-width="1.2" fill="none"/>
    <line x1="8" y1="8" x2="8" y2="13.5" stroke-width="1.2"/>
    <line x1="5.5" y1="13.5" x2="10.5" y2="13.5" stroke-width="1.2"/>
  `,

  // Notched right-pointing arrow — ArchiMate process symbol (confirmed correct)
  BusinessProcess: `
    <polygon points="1.5,5 9.5,5 9.5,2.5 14.5,8 9.5,13.5 9.5,11 1.5,11" stroke-width="1.4" fill="none" stroke-linejoin="round"/>
  `,

  // Thick open chevron pointing up — ArchiMate function symbol
  BusinessFunction: `
    <polygon points="8,3 14,8 14,13 8,8 2,13 2,8" stroke-width="1.4" fill="none" stroke-linejoin="round"/>
  `,

  // Two overlapping circles (same shape as collaboration)
  BusinessInteraction: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Oval / stadium
  BusinessService: `
    <rect x="2" y="5.5" width="12" height="5" rx="2.5" stroke-width="1.2" fill="none"/>
  `,

  // Trigger: flat right, notch on left (mirror of process)
  BusinessEvent: `
    <polygon points="4,5 14,5 14,11 4,11 1,8" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
  `,

  // Rectangle with header line — standard BusinessObject notation
  BusinessObject: `
    <rect x="2" y="3" width="12" height="10" rx="1" stroke-width="1.2" fill="none"/>
    <line x1="2" y1="6.5" x2="14" y2="6.5" stroke-width="1"/>
  `,

  // Rectangle with bottom tab (contract/document)
  Contract: `
    <rect x="3" y="2.5" width="10" height="10" rx="1" stroke-width="1.2" fill="none"/>
    <line x1="3" y1="9.5" x2="13" y2="9.5" stroke-width="1"/>
    <line x1="5" y1="11.5" x2="11" y2="11.5" stroke-width="1"/>
  `,

  // Rectangle with wavy bottom edge
  Representation: `
    <path d="M3,3 L13,3 L13,11 Q11,13 9,11 Q7,9 5,11 Q4,12 3,11 Z" stroke-width="1.2" fill="none"/>
  `,

  // Rectangle with small left-side tab (product)
  Product: `
    <rect x="4" y="3" width="10" height="10" rx="1" stroke-width="1.2" fill="none"/>
    <rect x="2" y="5" width="4" height="3" rx="0.5" stroke-width="1.2" fill="none"/>
  `,

  // ── Application layer ────────────────────────────────────────────────────

  // Rect with two plug tabs on left
  ApplicationComponent: `
    <rect x="4" y="2" width="10" height="12" rx="1" stroke-width="1.2" fill="none"/>
    <rect x="1" y="4.5" width="4" height="2.5" rx="0.5" stroke-width="1.2" fill="none"/>
    <rect x="1" y="9" width="4" height="2.5" rx="0.5" stroke-width="1.2" fill="none"/>
  `,

  // Two overlapping circles
  ApplicationCollaboration: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Lollipop
  ApplicationInterface: `
    <circle cx="8" cy="5" r="3" stroke-width="1.2" fill="none"/>
    <line x1="8" y1="8" x2="8" y2="13.5" stroke-width="1.2"/>
    <line x1="5.5" y1="13.5" x2="10.5" y2="13.5" stroke-width="1.2"/>
  `,

  // Hollow right-pointing arrow
  ApplicationProcess: `
    <polyline points="2,5.5 11,5.5 14.5,8 11,10.5 2,10.5 2,5.5" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
  `,

  // Thick open chevron pointing up — same as BusinessFunction
  ApplicationFunction: `
    <polygon points="8,3 14,8 14,13 8,8 2,13 2,8" stroke-width="1.4" fill="none" stroke-linejoin="round"/>
  `,

  // Two overlapping circles
  ApplicationInteraction: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Oval
  ApplicationService: `
    <rect x="2" y="5.5" width="12" height="5" rx="2.5" stroke-width="1.2" fill="none"/>
  `,

  // Trigger
  ApplicationEvent: `
    <polygon points="4,5 14,5 14,11 4,11 1,8" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
  `,

  // Rectangle with header line (mirrors BusinessObject)
  DataObject: `
    <rect x="2" y="3" width="12" height="10" rx="1" stroke-width="1.2" fill="none"/>
    <line x1="2" y1="6.5" x2="14" y2="6.5" stroke-width="1"/>
  `,

  // ── Technology layer ─────────────────────────────────────────────────────

  // 3D box (isometric)
  Node: `
    <rect x="2" y="5" width="10" height="8" rx="1" stroke-width="1.2" fill="none"/>
    <polyline points="2,5 5,2 14,2 14,10 12,13" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
    <line x1="12" y1="5" x2="14" y2="2" stroke-width="1.2"/>
    <line x1="12" y1="5" x2="12" y2="13" stroke-width="1"/>
  `,

  // Device: 3D flat box
  Device: `
    <rect x="2" y="6" width="10" height="7" rx="1" stroke-width="1.2" fill="none"/>
    <polyline points="2,6 4,3 13,3 13,10 12,13" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
    <line x1="12" y1="6" x2="13" y2="3" stroke-width="1.2"/>
  `,

  // Vertical cylinder
  SystemSoftware: `
    <ellipse cx="8" cy="4.5" rx="5" ry="1.8" stroke-width="1.2" fill="none"/>
    <line x1="3" y1="4.5" x2="3" y2="12" stroke-width="1.2"/>
    <line x1="13" y1="4.5" x2="13" y2="12" stroke-width="1.2"/>
    <ellipse cx="8" cy="12" rx="5" ry="1.8" stroke-width="1.2" fill="none"/>
  `,

  // Two overlapping circles
  TechnologyCollaboration: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Lollipop
  TechnologyInterface: `
    <circle cx="8" cy="5" r="3" stroke-width="1.2" fill="none"/>
    <line x1="8" y1="8" x2="8" y2="13.5" stroke-width="1.2"/>
    <line x1="5.5" y1="13.5" x2="10.5" y2="13.5" stroke-width="1.2"/>
  `,

  // Notched right-pointing arrow — same as BusinessProcess
  TechnologyProcess: `
    <polygon points="1.5,5 9.5,5 9.5,2.5 14.5,8 9.5,13.5 9.5,11 1.5,11" stroke-width="1.4" fill="none" stroke-linejoin="round"/>
  `,

  // Thick open chevron pointing up — same as BusinessFunction
  TechnologyFunction: `
    <polygon points="8,3 14,8 14,13 8,8 2,13 2,8" stroke-width="1.4" fill="none" stroke-linejoin="round"/>
  `,

  // Two overlapping circles
  TechnologyInteraction: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Oval
  TechnologyService: `
    <rect x="2" y="5.5" width="12" height="5" rx="2.5" stroke-width="1.2" fill="none"/>
  `,

  // Trigger
  TechnologyEvent: `
    <polygon points="4,5 14,5 14,11 4,11 1,8" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
  `,

  // Rect with two plug tabs (same as ApplicationComponent)
  TechnologyComponent: `
    <rect x="4" y="2" width="10" height="12" rx="1" stroke-width="1.2" fill="none"/>
    <rect x="1" y="4.5" width="4" height="2.5" rx="0.5" stroke-width="1.2" fill="none"/>
    <rect x="1" y="9" width="4" height="2.5" rx="0.5" stroke-width="1.2" fill="none"/>
  `,

  // Dog-eared document
  Artifact: `
    <path d="M3,2 L11,2 L13,4 L13,14 L3,14 Z" stroke-width="1.2" fill="none"/>
    <polyline points="11,2 11,4 13,4" stroke-width="1.2" fill="none"/>
    <line x1="5.5" y1="7" x2="10.5" y2="7" stroke-width="1"/>
    <line x1="5.5" y1="9.5" x2="10.5" y2="9.5" stroke-width="1"/>
  `,

  // Factory outline
  Facility: `
    <polyline points="2,13 2,7 5,7 5,10 8,7 8,10 11,7 11,10 14,7 14,13 2,13" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
  `,

  // Gear
  Equipment: `
    <circle cx="8" cy="8" r="2.5" stroke-width="1.2" fill="none"/>
    <circle cx="8" cy="8" r="4.5" stroke-width="0" fill="none"/>
    <line x1="8" y1="1.5" x2="8" y2="3.5" stroke-width="2"/>
    <line x1="8" y1="12.5" x2="8" y2="14.5" stroke-width="2"/>
    <line x1="1.5" y1="8" x2="3.5" y2="8" stroke-width="2"/>
    <line x1="12.5" y1="8" x2="14.5" y2="8" stroke-width="2"/>
    <line x1="3.5" y1="3.5" x2="4.9" y2="4.9" stroke-width="2"/>
    <line x1="11.1" y1="11.1" x2="12.5" y2="12.5" stroke-width="2"/>
    <line x1="12.5" y1="3.5" x2="11.1" y2="4.9" stroke-width="2"/>
    <line x1="4.9" y1="11.1" x2="3.5" y2="12.5" stroke-width="2"/>
  `,

  // Hexagon outline
  Material: `
    <polygon points="8,1.5 13.5,4.5 13.5,11.5 8,14.5 2.5,11.5 2.5,4.5" stroke-width="1.2" fill="none"/>
  `,

  // Network: dots connected
  CommunicationNetwork: `
    <circle cx="8" cy="8" r="1.5" fill="currentColor"/>
    <circle cx="3" cy="5" r="1.5" fill="currentColor"/>
    <circle cx="13" cy="5" r="1.5" fill="currentColor"/>
    <circle cx="3" cy="11" r="1.5" fill="currentColor"/>
    <circle cx="13" cy="11" r="1.5" fill="currentColor"/>
    <line x1="8" y1="8" x2="3" y2="5" stroke-width="1.2"/>
    <line x1="8" y1="8" x2="13" y2="5" stroke-width="1.2"/>
    <line x1="8" y1="8" x2="3" y2="11" stroke-width="1.2"/>
    <line x1="8" y1="8" x2="13" y2="11" stroke-width="1.2"/>
  `,

  // Bidirectional arrow
  Path: `
    <line x1="1" y1="8" x2="15" y2="8" stroke-width="1.2"/>
    <polyline points="4,5.5 1,8 4,10.5" stroke-width="1.2" fill="none"/>
    <polyline points="12,5.5 15,8 12,10.5" stroke-width="1.2" fill="none"/>
  `,

  // Double bidirectional arrow
  DistributionNetwork: `
    <line x1="1" y1="6" x2="15" y2="6" stroke-width="1.2"/>
    <polyline points="4,3.5 1,6 4,8.5" stroke-width="1.2" fill="none"/>
    <polyline points="12,3.5 15,6 12,8.5" stroke-width="1.2" fill="none"/>
    <line x1="1" y1="10" x2="15" y2="10" stroke-width="1.2"/>
    <polyline points="4,7.5 1,10 4,12.5" stroke-width="1.2" fill="none"/>
    <polyline points="12,7.5 15,10 12,12.5" stroke-width="1.2" fill="none"/>
  `,

  // ── Motivation layer ─────────────────────────────────────────────────────

  // Cylinder (same as role)
  Stakeholder: `
    <ellipse cx="8" cy="4.5" rx="5" ry="1.8" stroke-width="1.2" fill="none"/>
    <line x1="3" y1="4.5" x2="3" y2="12" stroke-width="1.2"/>
    <line x1="13" y1="4.5" x2="13" y2="12" stroke-width="1.2"/>
    <ellipse cx="8" cy="12" rx="5" ry="1.8" stroke-width="1.2" fill="none"/>
  `,

  // Compass wheel
  Driver: `
    <circle cx="8" cy="8" r="6" stroke-width="1.2" fill="none"/>
    <circle cx="8" cy="8" r="1.5" stroke-width="1" fill="none"/>
    <line x1="8" y1="2" x2="8" y2="6.5" stroke-width="1.5"/>
    <line x1="8" y1="9.5" x2="8" y2="14" stroke-width="1.5"/>
    <line x1="2" y1="8" x2="6.5" y2="8" stroke-width="1.5"/>
    <line x1="9.5" y1="8" x2="14" y2="8" stroke-width="1.5"/>
  `,

  // Magnifying glass
  Assessment: `
    <circle cx="6.5" cy="6.5" r="4" stroke-width="1.2" fill="none"/>
    <line x1="9.5" y1="9.5" x2="14" y2="14" stroke-width="1.5"/>
  `,

  // Bullseye (target)
  Goal: `
    <circle cx="8" cy="8" r="6" stroke-width="1.2" fill="none"/>
    <circle cx="8" cy="8" r="3.5" stroke-width="1.2" fill="none"/>
    <circle cx="8" cy="8" r="1.2" fill="currentColor"/>
  `,

  // Bullseye with outer ring filled (outcome = achieved goal)
  Outcome: `
    <circle cx="8" cy="8" r="6" stroke-width="1.2" fill="none"/>
    <circle cx="8" cy="8" r="3.5" stroke-width="2.5" fill="none"/>
    <circle cx="8" cy="8" r="1.2" fill="currentColor"/>
  `,

  // Exclamation mark
  Principle: `
    <line x1="8" y1="2.5" x2="8" y2="9.5" stroke-width="2" stroke-linecap="round"/>
    <circle cx="8" cy="12.5" r="1.3" fill="currentColor"/>
  `,

  // Parallelogram (slanted rectangle)
  Requirement: `
    <polygon points="4,3.5 14,3.5 12,12.5 2,12.5" stroke-width="1.2" fill="none"/>
  `,

  // Parallelogram (more slanted)
  Constraint: `
    <polygon points="5,3.5 14,3.5 11,12.5 2,12.5" stroke-width="1.2" fill="none"/>
  `,

  // Oval
  Value: `
    <ellipse cx="8" cy="8" rx="6.5" ry="4" stroke-width="1.2" fill="none"/>
  `,

  // Cloud / speech bubble
  Meaning: `
    <path d="M4,11 Q1,11 1,8 Q1,5.5 3.5,5 Q3.5,2 6.5,2 Q8.5,2 9.5,3.5 Q10.5,2 12,2 Q14.5,2 14.5,5 Q15,5.5 15,8 Q15,11 12,11 L6,13 Z" stroke-width="1.2" fill="none"/>
  `,

  // ── Strategy layer ───────────────────────────────────────────────────────

  // Three vertical bars (resource/bar chart)
  Resource: `
    <rect x="2" y="7" width="3" height="6" rx="0.5" stroke-width="1.2" fill="none"/>
    <rect x="6.5" y="4" width="3" height="9" rx="0.5" stroke-width="1.2" fill="none"/>
    <rect x="11" y="5.5" width="3" height="7.5" rx="0.5" stroke-width="1.2" fill="none"/>
  `,

  // Stacked rectangles
  Capability: `
    <rect x="2" y="9.5" width="12" height="3.5" rx="0.5" stroke-width="1.2" fill="none"/>
    <rect x="3.5" y="6" width="9" height="3.5" rx="0.5" stroke-width="1.2" fill="none"/>
    <rect x="5" y="2.5" width="6" height="3.5" rx="0.5" stroke-width="1.2" fill="none"/>
  `,

  // Chevron (same as Function — ValueStream reuses)
  ValueStream: `
    <polygon points="5,5 13,5 13,11 5,11 2,8" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
  `,

  // Bold arrow with target circle at tip
  CourseOfAction: `
    <line x1="1" y1="8" x2="12" y2="8" stroke-width="2" stroke-linecap="round"/>
    <polyline points="9,5 13,8 9,11" stroke-width="1.8" fill="none" stroke-linejoin="round"/>
  `,

  // ── Implementation & Migration layer ─────────────────────────────────────

  // Clipboard
  WorkPackage: `
    <rect x="3" y="4" width="10" height="11" rx="1" stroke-width="1.2" fill="none"/>
    <path d="M6,4 L6,2.5 Q6,1.5 8,1.5 Q10,1.5 10,2.5 L10,4" stroke-width="1.2" fill="none"/>
    <line x1="5.5" y1="7.5" x2="10.5" y2="7.5" stroke-width="1"/>
    <line x1="5.5" y1="10" x2="9" y2="10" stroke-width="1"/>
  `,

  // Trigger (same shape as event)
  ImplementationEvent: `
    <polygon points="4,5 14,5 14,11 4,11 1,8" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
  `,

  // Wave / scroll shape
  Deliverable: `
    <path d="M2,4 Q2,2 4,2 L12,2 Q14,2 14,4 L14,12 Q14,14 12,14 Q10,14 10,12 Q10,10 8,10 L2,10 Z" stroke-width="1.2" fill="none"/>
  `,

  // Three horizontal lines (plateau = stable period)
  Plateau: `
    <line x1="2" y1="5" x2="14" y2="5" stroke-width="2" stroke-linecap="round"/>
    <line x1="2" y1="8" x2="14" y2="8" stroke-width="2" stroke-linecap="round"/>
    <line x1="2" y1="11" x2="14" y2="11" stroke-width="2" stroke-linecap="round"/>
  `,

  // Circle with horizontal line (gap / difference)
  Gap: `
    <circle cx="8" cy="8" r="6" stroke-width="1.2" fill="none"/>
    <line x1="2" y1="8" x2="14" y2="8" stroke-width="1.5"/>
  `,

  // ── Physical / Composite ─────────────────────────────────────────────────

  // Map pin
  Location: `
    <circle cx="8" cy="6" r="4" stroke-width="1.2" fill="none"/>
    <path d="M4.5,9 Q5,13 8,15 Q11,13 11.5,9" stroke-width="1.2" fill="none"/>
  `,

  // ── Grouping ─────────────────────────────────────────────────────────────
  Grouping: `
    <rect x="1.5" y="1.5" width="13" height="13" rx="1.5" stroke-width="1.2" fill="none" stroke-dasharray="3 2"/>
  `,
};

// Fallback for unknown types
export const DEFAULT_ICON = `
  <rect x="3" y="3" width="10" height="10" rx="1.5" stroke-width="1.2" fill="none"/>
`;

export function getIcon(elementType) {
  return ICONS[elementType] || DEFAULT_ICON;
}

// ── Layer color palette ───────────────────────────────────────────────────
// Hues follow the ArchiMate 3 layer convention; saturation is toned down
// from the standard to be easier on the eye (per project preference).
//
//  Business          → amber
//  Application       → blue
//  Technology        → green  (user prefers our green over standard cyan)
//  Motivation        → violet
//  Strategy          → gold   (distinct from business amber)
//  Implementation    → rose
//  Physical/Location → pink/fuchsia

const BUSINESS  = { fill: '#FFFBEB', stroke: '#D97706', text: '#78350F' };
const APP       = { fill: '#EFF6FF', stroke: '#2563EB', text: '#1E3A8A' };
const TECH      = { fill: '#F0FDF4', stroke: '#16A34A', text: '#14532D' };
const MOTIV     = { fill: '#FAF5FF', stroke: '#7C3AED', text: '#3B0764' };
const STRATEGY  = { fill: '#FEFCE8', stroke: '#B45309', text: '#78350F' };
const IMPL      = { fill: '#FFF1F2', stroke: '#BE123C', text: '#881337' };
const PHYSICAL  = { fill: '#FDF4FF', stroke: '#A21CAF', text: '#701A75' };

export const LAYER_COLORS = {
  // Business
  BusinessActor:         BUSINESS,
  BusinessRole:          BUSINESS,
  BusinessCollaboration: BUSINESS,
  BusinessInterface:     BUSINESS,
  BusinessProcess:       BUSINESS,
  BusinessFunction:      BUSINESS,
  BusinessInteraction:   BUSINESS,
  BusinessService:       BUSINESS,
  BusinessEvent:         BUSINESS,
  BusinessObject:        BUSINESS,
  Contract:              BUSINESS,
  Representation:        BUSINESS,
  Product:               BUSINESS,

  // Application
  ApplicationComponent:     APP,
  ApplicationCollaboration: APP,
  ApplicationInterface:     APP,
  ApplicationProcess:       APP,
  ApplicationFunction:      APP,
  ApplicationInteraction:   APP,
  ApplicationService:       APP,
  ApplicationEvent:         APP,
  DataObject:               APP,

  // Technology
  Node:                   TECH,
  Device:                 TECH,
  SystemSoftware:         TECH,
  TechnologyCollaboration:TECH,
  TechnologyInterface:    TECH,
  TechnologyProcess:      TECH,
  TechnologyFunction:     TECH,
  TechnologyInteraction:  TECH,
  TechnologyService:      TECH,
  TechnologyEvent:        TECH,
  Artifact:               TECH,
  Facility:               TECH,
  Equipment:              TECH,
  Material:               TECH,
  CommunicationNetwork:   TECH,
  Path:                   TECH,
  DistributionNetwork:    TECH,
  TechnologyComponent:    TECH,

  // Motivation
  Stakeholder:  MOTIV,
  Driver:       MOTIV,
  Assessment:   MOTIV,
  Goal:         MOTIV,
  Outcome:      MOTIV,
  Principle:    MOTIV,
  Requirement:  MOTIV,
  Constraint:   MOTIV,
  Value:        MOTIV,
  Meaning:      MOTIV,

  // Strategy
  Resource:       STRATEGY,
  Capability:     STRATEGY,
  ValueStream:    STRATEGY,
  CourseOfAction: STRATEGY,

  // Implementation & Migration
  WorkPackage:         IMPL,
  ImplementationEvent: IMPL,
  Deliverable:         IMPL,
  Plateau:             IMPL,
  Gap:                 IMPL,

  // Physical / Composite
  Location: PHYSICAL,

  // Grouping — transparent with dashed border
  Grouping: { fill: 'rgba(249,250,251,0.5)', stroke: '#9CA3AF', text: '#374151', dashed: true },
};

export const DEFAULT_COLOR = { fill: '#F9FAFB', stroke: '#6B7280', text: '#374151' };

export function getColor(elementType) {
  return LAYER_COLORS[elementType] || DEFAULT_COLOR;
}
