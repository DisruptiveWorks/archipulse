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
    <rect x="1.33" y="4.67" width="13.33" height="6.67" rx="3.33" stroke-width="1.2" fill="none"/>
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
    <rect fill="none" height="10" id="svg_1" rx="1" stroke-width="1.2" width="10" x="3" y="2.5"/>
  <line id="svg_2" x1="3.2" x2="12.8" y1="5" y2="5"/>
  <line id="svg_3" x1="3.2" x2="12.8" y1="10" y2="10"/>
  `,

  // Document with curved bottom + horizontal divider
  Representation: `
    <path d="M1.5,2 H14.5 V11 Q8,15 1.5,11 Z" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
    <line x1="1.5" y1="6.5" x2="14.5" y2="6.5" stroke-width="1.2"/>
  `,

  // Rectangle with small left-side tab (product)
  Product: `
<rect fill="none" height="10" id="svg_1" rx="1" stroke-width="1.2" transform="matrix(1.20118 0 0 1 -2.86063 0)" width="10" x="4" y="3"/>
  <rect fill="none" height="3" id="svg_2" rx="0.5" stroke-width="1.2" transform="matrix(1.4 0 0 1 -1.92 0)" width="4" x="2.8" y="3"/>
  `,

  // ── Application layer ────────────────────────────────────────────────────

  // Rect with two plug tabs on left (gaps in left border where tabs connect, tabs are closed rects)
  ApplicationComponent: `
<path d="m3.5,2l11,0l0,12l-11,0l0,-2.5m0,-3l0,-2m0,-3l0,-1.5" fill="none" id="svg_1" stroke-linecap="round" stroke-width="1.2"/>
  <rect fill="none" height="3" id="svg_2" stroke-width="1.2" width="4.9" x="1" y="3.5"/>
  <rect fill="none" height="3" id="svg_3" stroke-width="1.2" transform="matrix(1.225 0 0 1 -0.225 0)" width="4" x="1" y="8.5"/>
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
    <polygon points="8,3 14,6 14,13 8,10 2,13 2,6" 
            stroke="currentColor" 
            stroke-width="1.4" 
            fill="none" 
            stroke-linejoin="round"/>  `,

  // Two overlapping circles
  ApplicationInteraction: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Oval
  ApplicationService: `
    <rect x="1.33" y="4.67" width="13.33" height="6.67" rx="3.33" stroke-width="1.2" fill="none"/>
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
    <g transform="translate(18.000000,11.000000) rotate(180) scale(0.01)"
fill="currentColor" stroke="none" stroke-width="1.2" >
<path stroke-width="50" d="M296 1269 c-55 -43 -56 -50 -56 -407 0 -326 0 -329 23 -363 30 -44
60 -59 125 -59 l54 0 -101 -107 c-55 -59 -98 -110 -95 -115 8 -13 1302 -10
1309 3 4 5 -38 56 -94 112 l-101 102 66 6 c56 5 70 10 97 38 l32 31 3 345 c2
296 1 349 -13 375 -32 61 -17 60 -648 60 -567 0 -575 0 -601 -21z m1187 -10
c55 -25 57 -38 57 -398 l0 -331 -29 -32 -29 -33 -556 -3 c-306 -1 -570 1 -588
6 -17 4 -41 20 -52 36 -20 26 -21 41 -24 349 -2 292 -1 325 15 357 10 19 29
40 43 47 33 17 1127 19 1163 2z m-151 -819 c2 -5 45 -51 96 -103 50 -52 92
-96 92 -98 0 -2 -278 -4 -619 -4 l-619 0 82 88 c44 48 88 97 96 107 14 18 34
19 442 20 261 0 428 -4 430 -10z"/>
<path d="M1473 1215 c0 -8 4 -12 9 -9 5 3 6 10 3 15 -9 13 -12 11 -12 -6z"/>
<path d="M1473 515 c0 -8 4 -12 9 -9 5 3 6 10 3 15 -9 13 -12 11 -12 -6z"/>
<path d="M443 363 c4 -3 1 -13 -6 -22 -11 -14 -10 -14 5 -2 16 12 16 31 1 31
-4 0 -3 -3 0 -7z"/>
<path d="M266 308 c3 -5 10 -6 15 -3 13 9 11 12 -6 12 -8 0 -12 -4 -9 -9z"/>
<path d="M1550 306 c0 -2 8 -10 18 -17 15 -13 16 -12 3 4 -13 16 -21 21 -21
13z"/>
</g>
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
    <polygon points="8,3 14,6 14,13 8,10 2,13 2,6" 
            stroke="currentColor" 
            stroke-width="1.4" 
            fill="none" 
            stroke-linejoin="round"/>  `,

  // Two overlapping circles
  TechnologyInteraction: `
    <circle cx="5.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
    <circle cx="10.5" cy="8" r="4" stroke-width="1.2" fill="none"/>
  `,

  // Oval
  TechnologyService: `
    <rect x="1.33" y="4.67" width="13.33" height="6.67" rx="3.33" stroke-width="1.2" fill="none"/>
  `,

  // Flag shape: notch on left, arc on right
  TechnologyEvent: `
    <path d="M1.3,2.7 l2.7,5.3 -2.7,5.3 h9.3 a4,4 0 0 0 0,-10.7 z" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
  `,

  // Rect with two plug tabs (same as ApplicationComponent)
  TechnologyComponent: `
  <path d="m3.5,2l11,0l0,12l-11,0l0,-2.5m0,-3l0,-2m0,-3l0,-1.5" fill="none" id="svg_1" stroke-linecap="round" stroke-width="1.2"/>
  <rect fill="none" height="3" id="svg_2" stroke-width="1.2" width="4.9" x="1" y="3.5"/>
  <rect fill="none" height="3" id="svg_3" stroke-width="1.2" transform="matrix(1.225 0 0 1 -0.225 0)" width="4" x="1" y="8.5"/>
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
    <path d="M1.33,3.33 h3.33 v6.67 l2.67,-2 v2 l2.67,-2 v2 l2.67,-2 v6.67 H1.33 Z" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
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
  <g transform="scale(0.14)">
    <path fill="none" opacity="1.000000" stroke="none" 
	d="
M88.000000,97.000000 
	C58.696236,97.000000 29.892473,97.000000 1.044355,97.000000 
	C1.044355,65.062439 1.044355,33.124802 1.044355,1.093582 
	C50.220196,1.093582 99.440628,1.093582 148.830536,1.093582 
	C148.830536,32.999290 148.830536,64.999557 148.830536,97.000000 
	C128.791946,97.000000 108.645973,97.000000 88.000000,97.000000 
M107.293739,82.954536 
	C117.338760,74.992989 127.393929,67.044205 137.424667,59.064686 
	C142.744156,54.832977 142.761169,51.757488 137.612396,47.523304 
	C132.598343,43.399925 127.604118,39.252319 122.615051,35.098717 
	C117.195633,30.586843 111.792191,26.055796 106.251396,21.424114 
	C100.797386,28.161650 107.328186,30.759243 110.143738,34.580807 
	C86.231644,34.580807 62.890877,34.580807 39.249580,34.580807 
	C49.035793,24.324099 49.035793,24.324099 45.365459,19.820595 
	C44.330620,20.548933 43.217861,21.220655 42.231354,22.043695 
	C32.512936,30.151728 22.810070,38.278435 13.114287,46.413540 
	C7.933849,50.760113 8.044102,53.733402 13.411368,57.885590 
	C22.364418,64.811790 31.234295,71.847481 40.279800,78.650101 
	C41.671982,79.697083 43.814335,79.746544 45.609749,80.257332 
	C45.291096,77.813034 45.428810,75.174614 44.499531,72.990067 
	C43.837486,71.433723 41.740028,70.487991 40.022877,69.052124 
	C64.369675,69.052124 88.073586,69.052124 112.850624,69.052124 
	C109.363403,71.902588 106.554420,73.612747 104.601723,76.018051 
	C103.459503,77.425026 103.446915,79.943344 103.462868,81.957451 
	C103.466782,82.451004 105.574478,82.927879 107.293739,82.954536 
z"/>
<path fill="currentColor" opacity="1.000000" stroke="none" 
	d="
M107.004349,83.183472 
	C105.574478,82.927879 103.466782,82.451004 103.462868,81.957451 
	C103.446915,79.943344 103.459503,77.425026 104.601723,76.018051 
	C106.554420,73.612747 109.363403,71.902588 112.850624,69.052124 
	C88.073586,69.052124 64.369675,69.052124 40.022877,69.052124 
	C41.740028,70.487991 43.837486,71.433723 44.499531,72.990067 
	C45.428810,75.174614 45.291096,77.813034 45.609749,80.257332 
	C43.814335,79.746544 41.671982,79.697083 40.279800,78.650101 
	C31.234295,71.847481 22.364418,64.811790 13.411368,57.885590 
	C8.044102,53.733402 7.933849,50.760113 13.114287,46.413540 
	C22.810070,38.278435 32.512936,30.151728 42.231354,22.043695 
	C43.217861,21.220655 44.330620,20.548933 45.365459,19.820595 
	C49.035793,24.324099 49.035793,24.324099 39.249580,34.580807 
	C62.890877,34.580807 86.231644,34.580807 110.143738,34.580807 
	C107.328186,30.759243 100.797386,28.161650 106.251396,21.424114 
	C111.792191,26.055796 117.195633,30.586843 122.615051,35.098717 
	C127.604118,39.252319 132.598343,43.399925 137.612396,47.523304 
	C142.761169,51.757488 142.744156,54.832977 137.424667,59.064686 
	C127.393929,67.044205 117.338760,74.992989 107.004349,83.183472 
M47.500885,59.938896 
	C69.465500,60.079117 91.437729,59.932652 113.390587,60.515575 
	C121.040428,60.718700 127.443222,59.411987 132.441772,53.108883 
	C126.606422,47.571815 120.952538,42.654816 111.278397,42.976295 
	C87.342888,43.771698 63.351456,43.666817 39.407387,43.012665 
	C30.605988,42.772213 24.063232,45.376759 18.335131,52.093292 
	C23.521008,56.195118 28.163488,61.203781 36.032803,59.969959 
	C39.440434,59.435680 43.007046,59.915379 47.500885,59.938896 
z"/>
<path fill="none" opacity="1.000000" stroke="none" 
	d="
M47.001163,59.936699 
	C43.007046,59.915379 39.440434,59.435680 36.032803,59.969959 
	C28.163488,61.203781 23.521008,56.195118 18.335131,52.093292 
	C24.063232,45.376759 30.605988,42.772213 39.407387,43.012665 
	C63.351456,43.666817 87.342888,43.771698 111.278397,42.976295 
	C120.952538,42.654816 126.606422,47.571815 132.441772,53.108883 
	C127.443222,59.411987 121.040428,60.718700 113.390587,60.515575 
	C91.437729,59.932652 69.465500,60.079117 47.001163,59.936699 
z"/>
</g>
  `,

  // ── Motivation layer ─────────────────────────────────────────────────────

  // Horizontal cylinder (same as BusinessRole)
  Stakeholder: `
    <ellipse cx="12" cy="8" rx="3" ry="5" stroke-width="1.2" fill="none"/>
    <line x1="4" y1="3" x2="12" y2="3" stroke-width="1.2"/>
    <line x1="4" y1="13" x2="12" y2="13" stroke-width="1.2"/>
    <path d="M4,3 a3,5 0 0 0 0,10" stroke-width="1.2" fill="none"/>
  `,

  // Wheel: outer circle, 8 spokes from center, filled inner circle
  Driver: `
  <circle cx="8" cy="8" fill="none" id="svg_1" r="6.4" stroke-width="1.2" transform="matrix(1 0 0 1 0 0)"/>
  <g id="svg_2" stroke-width="0.9">
   <line id="svg_3" x1="8" x2="8" y1="8" y2="0"/>
   <line id="svg_4" x1="8" x2="13.66" y1="8" y2="2.34"/>
   <line id="svg_5" x1="8" x2="16" y1="8" y2="8"/>
   <line id="svg_6" x1="8" x2="13.66" y1="8" y2="13.66"/>
   <line id="svg_7" x1="8" x2="8" y1="8" y2="16"/>
   <line id="svg_8" x1="8" x2="2.34" y1="8" y2="13.66"/>
   <line id="svg_9" x1="8" x2="0" y1="8" y2="8"/>
   <line id="svg_10" x1="8" x2="2.34" y1="8" y2="2.34"/>
  </g>
  <circle cx="8" cy="8" fill="currentColor" id="svg_11" r="1"/>
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

  // Circle overlapping a right-pointing arrow
  WorkPackage: `
  <g transform="scale(0.14)">
<path fill="none" opacity="1.000000" stroke="none" d="M95.000000,133.000000 
    C63.360981,133.000000 32.221962,133.000000 1.041472,133.000000 
    C1.041472,89.066711 1.041472,45.133377 1.041472,1.100020 
    C50.888790,1.100020 100.777847,1.100020 150.833450,1.100020 
    C150.833450,44.999489 150.833450,88.999687 150.833450,133.000000 
    C132.467087,133.000000 113.983543,133.000000 95.000000,133.000000 
    M82.305595,89.319229 
    C92.270493,82.570717 98.009743,73.055206 99.547798,61.290340 
    C103.547508,30.695927 75.434105,7.241077 47.534653,17.965097 
    C30.863958,24.372999 22.448303,37.480202 20.373682,54.898693 
    C17.542580,78.668587 36.135883,101.021332 60.008804,102.052231 
    C72.315102,102.583649 84.661293,102.191360 96.989464,102.215469 
    C98.591843,102.218597 100.194237,102.215897 102.297043,102.215897 
    C102.297043,107.843094 102.297043,112.858215 102.297043,118.436432 
    C114.336105,111.116486 125.737007,104.184555 137.609055,96.966164 
    C131.202820,93.067192 125.544090,89.605362 119.867432,86.173195 
    C114.114594,82.694969 108.344017,79.246063 101.809395,75.320679 
    C101.809395,81.404037 101.809395,86.473030 101.809395,91.553551 
    C95.218872,91.553551 89.083412,91.553551 82.672607,91.159752 
    C82.445435,90.742325 82.218269,90.324898 82.305595,89.319229 
    Z"/>

  <path fill="currentColor" opacity="1.000000" stroke="none" d="M82.947960,91.553551 
    C89.083412,91.553551 95.218872,91.553551 101.809395,91.553551 
    C101.809395,86.473030 101.809395,81.404037 101.809395,75.320679 
    C108.344017,79.246063 114.114594,82.694969 119.867432,86.173195 
    C125.544090,89.605362 131.202820,93.067192 137.609055,96.966164 
    C125.737007,104.184555 114.336105,111.116486 102.297043,118.436432 
    C102.297043,112.858215 102.297043,107.843094 102.297043,102.215897 
    C100.194237,102.215897 98.591843,102.218597 96.989464,102.215469 
    C84.661293,102.191360 72.315102,102.583649 60.008804,102.052231 
    C36.135883,101.021332 17.542580,78.668587 20.373682,54.898693 
    C22.448303,37.480202 30.863958,24.372999 47.534653,17.965097 
    C75.434105,7.241077 103.547508,30.695927 99.547798,61.290340 
    C98.009743,73.055206 92.270493,82.570717 81.768593,89.694626 
    C78.249870,92.277779 75.884628,91.156281 74.029663,88.527939 
    C72.077904,85.762444 73.432129,83.455757 75.530441,81.508636 
    C76.866493,80.268837 78.504944,79.352028 79.823700,78.096954 
    C88.158783,70.164345 91.725700,60.567841 89.094856,49.202957 
    C86.448898,37.772770 79.645607,29.850309 68.041428,26.909723 
    C51.310947,22.670095 33.991848,34.401680 31.029915,51.705307 
    C27.712395,71.086296 39.542923,89.341919 57.049171,91.021553 
    C65.620209,91.843903 74.311241,91.415741 82.947960,91.553551 
    Z"/>

  <path fill="none" opacity="1.000000" stroke="none" d="M82.810287,91.356651 
    C74.311241,91.415741 65.620209,91.843903 57.049171,91.021553 
    C39.542923,89.341919 27.712395,71.086296 31.029915,51.705307 
    C33.991848,34.401680 51.310947,22.670095 68.041428,26.909723 
    C79.645607,29.850309 86.448898,37.772770 89.094856,49.202957 
    C91.725700,60.567841 88.158783,70.164345 79.823700,78.096954 
    C78.504944,79.352028 76.866493,80.268837 75.530441,81.508636 
    C73.432129,83.455757 72.077904,85.762444 74.029663,88.527939 
    C75.884628,91.156281 78.249870,92.277779 81.611343,89.988747 
    C82.218269,90.324898 82.445435,90.742325 82.810287,91.356651 
    Z"/>
    </g>
  `,

  // Flag shape: notch on left, arc on right
  ImplementationEvent: `
    <path d="M1.3,2.7 l2.7,5.3 -2.7,5.3 h9.3 a4,4 0 0 0 0,-10.7 z" stroke-width="1.2" fill="none" stroke-linejoin="round"/>
  `,

  // Rect with S-wave bottom edge (left dips down, middle rises, right mid)
  Deliverable: `
    <path d="m1.5,2l13,0l0,9.5c-1.5,-1.5 -3.5,-1.5 -5.5,-1c-2,0.5 -3.5,4 -5,4c-1.5,0 -2.5,-1.5 -2.5,-2.5l0,-10z" fill="none" id="svg_1" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.2" transform="matrix(1 0 0 0.622222 0 0.755556)"/>
  `,

  // Three horizontal lines (plateau = stable period)
  Plateau: `
<line id="svg_1" stroke-linecap="round" stroke-width="2" transform="matrix(0.825 0 0 1 2.45 0)" x1="2" x2="14" y1="5" y2="5"/>
  <line id="svg_2" stroke-linecap="round" stroke-width="2" transform="matrix(0.900901 0 0 1 1.2982 0)" x1="2" x2="13.1" y1="8" y2="8"/>
  <line id="svg_3" stroke-linecap="round" stroke-width="2" transform="matrix(0.83277 0 0 1 0.487841 0)" x1="2" x2="14" y1="11" y2="11"/>
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

  // ── Grouping / Group ─────────────────────────────────────────────────────
  // Grouping: small tab on top-left + body rect (ArchiMate element)
  Grouping: `
    <rect x="1.5" y="4" width="13" height="10.5" rx="0.5" stroke-width="1.2" fill="none"/>
    <rect x="1.5" y="2" width="5" height="2.5" rx="0.5" stroke-width="1.2" fill="none"/>
  `,

  // Group: same shape but dashed (ArchiMate diagram grouping construct)
  Group: `
    <rect x="1.5" y="4" width="13" height="10.5" rx="0.5" stroke-width="1.2" fill="none" stroke-dasharray="2 1"/>
    <rect x="1.5" y="2" width="5" height="2.5" rx="0.5" stroke-width="1.2" fill="none" stroke-dasharray="2 1"/>
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

  // Grouping — semi-transparent with solid border
  Grouping: { fill: 'rgba(249,250,251,0.5)', stroke: '#9CA3AF', text: '#374151', dashed: false },

  // Group — transparent with dashed border (diagram-level grouping construct)
  Group: { fill: 'rgba(249,250,251,0.3)', stroke: '#9CA3AF', text: '#374151', dashed: true },
};

export const DEFAULT_COLOR = { fill: '#F9FAFB', stroke: '#6B7280', text: '#374151' };

export function getColor(elementType) {
  return LAYER_COLORS[elementType] || DEFAULT_COLOR;
}
