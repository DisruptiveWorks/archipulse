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
//   Cylinder   → vertical cylinder (Role, Stakeholder)
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

  // Horizontal cylinder — ellipse on right, half-arc on left, two rails
  BusinessRole: `
    <ellipse cx="12" cy="8" rx="3" ry="5" stroke-width="1.2" fill="none"/>
    <line x1="4" y1="3" x2="12" y2="3" stroke-width="1.2"/>
    <line x1="4" y1="13" x2="12" y2="13" stroke-width="1.2"/>
    <path d="M4,3 a3,5 0 0 0 0,10" stroke-width="1.2" fill="none"/>
  `,

  // Two overlapping circles
  BusinessCollaboration: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Circle on right + horizontal line on left
  BusinessInterface: `
    <circle cx="10.7" cy="8" r="3.3" stroke-width="1.2" fill="none"/>
    <line x1="1.3" y1="8" x2="7.4" y2="8" stroke-width="1.2"/>
  `,

  // Notched right-pointing arrow — ArchiMate process symbol (confirmed correct)
  BusinessProcess: `
    <polygon points="1.5,5 9.5,5 9.5,2.5 14.5,8 9.5,13.5 9.5,11 1.5,11" stroke-width="1.4" fill="none" stroke-linejoin="round"/>
  `,

  // Thick open chevron pointing up — ArchiMate function symbol
  BusinessFunction: `
    <polygon points="8,3 14,6 14,13 8,10 2,13 2,6" 
            stroke="currentColor" 
            stroke-width="1.4" 
            fill="none" 
            stroke-linejoin="round"/>
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

  // Flag shape: notch on left, arc on right
  BusinessEvent: `
    <path d="M1.3,2.7 l2.7,5.3 -2.7,5.3 h9.3 a4,4 0 0 0 0,-10.7 z" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
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

  // Circle on right + horizontal line on left
  ApplicationInterface: `
    <circle cx="10.7" cy="8" r="3.3" stroke-width="1.2" fill="none"/>
    <line x1="1.3" y1="8" x2="7.4" y2="8" stroke-width="1.2"/>
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

  // Flag shape: notch on left, arc on right
  ApplicationEvent: `
    <path d="M1.3,2.7 l2.7,5.3 -2.7,5.3 h9.3 a4,4 0 0 0 0,-10.7 z" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
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

  // Two overlapping circles with depth mask (ArchiMate SystemSoftware)
  SystemSoftware: `
    <defs>
      <mask id="ss-mask">
        <rect width="16" height="16" fill="white"/>
        <circle cx="6.67" cy="9.33" r="4" fill="black"/>
      </mask>
    </defs>
    <circle cx="8.67" cy="7.33" r="4" stroke-width="1.2" fill="none" mask="url(#ss-mask)"/>
    <circle cx="6.67" cy="9.33" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Two overlapping circles
  TechnologyCollaboration: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Circle on right + horizontal line on left
  TechnologyInterface: `
    <circle cx="10.7" cy="8" r="3.3" stroke-width="1.2" fill="none"/>
    <line x1="1.3" y1="8" x2="7.4" y2="8" stroke-width="1.2"/>
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

  // Flag shape: notch on left, arc on right
  TechnologyEvent: `
    <path d="M1.3,2.7 l2.7,5.3 -2.7,5.3 h9.3 a4,4 0 0 0 0,-10.7 z" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
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

  // Two interlocking gears — original paths scaled from 152×144 viewBox
  Equipment: `
    <g transform="scale(0.10526,0.11111)" stroke-width="11" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path d="M77.5,64.5 C80.5,58.6 79.2,55 73.5,53.5 C72.2,53.2 70.8,53.3 69.5,53 C67,52.2 65.5,52 64.9,55.5 C64,59.9 61.4,62.6 56,62.3 C53.9,62.2 52.4,61.9 51.6,60.5 C50.3,58.1 48.5,56.5 47,55.1 C44.3,54.6 42.8,56.4 41.1,57.2 C39.3,58 37.2,59 36,61.3 C36.9,63.5 38.1,66 39,68.5 C40.3,72.1 34.8,78 31,76.9 C28.6,76.2 26,74.9 24.1,77.1 C21.7,79.6 21.8,83.3 21.2,86.5 C20.7,88.7 23.4,89.3 25,89.9 C28,91.1 30.8,91.5 31.2,96 C31.6,100.1 30.7,102.4 27,104 C25.6,104.7 24.5,105.8 23.6,106.4 C25.5,111.8 26.9,113.9 31,117.5 C33.9,116.4 36.7,115.2 39.8,114 C40.1,114.2 40.5,114.7 41,115 C46,117.5 47.5,121.3 45.6,126.5 C45.3,127.2 45.1,128.3 46,129 C49,131.1 52.5,131.7 56,131.9 C58.9,132 59.2,129.1 60,127 C61.5,123.3 67.2,120.4 70.4,122.1 C72.2,123.1 73.2,125.3 74.5,127 C75.2,127.8 75.8,128.7 76.5,129.7 C80.6,128.2 84.3,126.3 87.6,123.5 C87,118.6 81.4,113.9 87.9,108.9 C91.7,106.1 95,109.1 98.6,108.4 C100.4,106.8 100.9,104.4 101.5,102 C102.7,97.2 102.1,95 97.5,94.2 C93.1,93.4 92.8,90.5 92.1,87.5 C91.3,84.6 92.6,82.7 95.5,81.4 C97.2,80.7 98.5,79.2 100,78 C99.9,74.3 97.4,71.5 96.3,68.1 C95.7,66.2 93,64.8 91.1,66.6 C86.2,71 82.5,69.3 79,65 Z"/>
      <path d="M100,75.5 C101,75.3 102,74.9 103,75 C107.8,75.8 110.1,74.2 110.1,69 C110.1,65.8 115,64 118,65.5 C119.7,66.3 120.4,68.4 123.2,68.4 C125.5,67.1 127.4,64.4 129.6,62.5 C129.2,59 126.2,57.9 125.1,55.9 C125.5,51.1 128.4,49.3 132.5,49 C137.4,48.6 135.1,44.9 135.4,42.5 C135.8,40 134.7,37.7 132,38.1 C126.9,38.9 126.3,35.2 124.9,32 C125.6,29.2 129,28.3 129.4,25.7 C127.9,22.4 125.6,20.1 122.3,18.6 C121.1,19.1 120.1,20 119,20.96 C115.1,24.1 110.6,22.5 110.1,17.5 C109.6,12.8 106.5,12.7 103,12.4 C99,12.1 98,14.8 98.1,17.5 C98.3,22.2 94.1,21.3 92.1,23.4 C89.7,21.7 87.6,20.1 85.5,18.5 C82.4,19.8 80.7,22.7 78.7,24.3 C78,27.2 80.1,27.7 81,29 C83.7,32.8 81.6,37.6 77,37.9 C74,38.2 71.7,38.7 72,42.5 C72.2,45.6 71,49.4 76.5,49.1 C79.4,49 81.1,50.9 82.1,54 C83.1,57 81.2,57.8 79.5,59 Z"/>
      <path d="M61.5,74 C56.3,75 51.7,76.5 47.9,80.9 C42.3,87.4 42.3,98 48.6,104.4 C55.2,111 63.6,112.2 72,106.9 C78.5,102.8 81.7,94.9 79.9,88 C78.2,81 71,74.2 64,74.5 C63.3,74.5 62.7,74.2 62,74 Z"/>
      <path d="M103.5,29 C98,29.6 93.7,31.8 91,37 C86.9,44.6 89.3,52 97.5,56.5 C104.4,60.3 113.3,56.9 117.1,50.6 C122.2,42 115.9,28.9 104,29 Z"/>
    </g>
  `,

  // Hexagon with internal facet lines (gem/crystal shape)
  Material: `
    <path d="M4.5,2 L11.5,2 L14.5,8 L11.5,14 L4.5,14 L1.5,8 Z" stroke-width="1.2" fill="none"/>
    <line x1="4" y1="7.5" x2="6" y2="3.5" stroke-width="1.2"/>
    <line x1="12" y1="7.5" x2="10" y2="3.5" stroke-width="1.2"/>
    <line x1="6" y1="12" x2="10" y2="12" stroke-width="1.2"/>
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

  // Horizontal cylinder (same as BusinessRole)
  Stakeholder: `
    <ellipse cx="12" cy="8" rx="3" ry="5" stroke-width="1.2" fill="none"/>
    <line x1="4" y1="3" x2="12" y2="3" stroke-width="1.2"/>
    <line x1="4" y1="13" x2="12" y2="13" stroke-width="1.2"/>
    <path d="M4,3 a3,5 0 0 0 0,10" stroke-width="1.2" fill="none"/>
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

  // Capability map staircase — 3 cells bottom row, 2 middle, 1 top-right
  Capability: `
    <rect x="2" y="10" width="4" height="4" stroke-width="1.2" fill="none"/>
    <rect x="6" y="10" width="4" height="4" stroke-width="1.2" fill="none"/>
    <rect x="10" y="10" width="4" height="4" stroke-width="1.2" fill="none"/>
    <rect x="6" y="6" width="4" height="4" stroke-width="1.2" fill="none"/>
    <rect x="10" y="6" width="4" height="4" stroke-width="1.2" fill="none"/>
    <rect x="10" y="2" width="4" height="4" stroke-width="1.2" fill="none"/>
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

  // Flag shape: notch on left, arc on right
  ImplementationEvent: `
    <path d="M1.3,2.7 l2.7,5.3 -2.7,5.3 h9.3 a4,4 0 0 0 0,-10.7 z" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
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
