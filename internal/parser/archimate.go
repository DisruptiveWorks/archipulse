package parser

// ValidElementTypes contains all valid ArchiMate 3.2 element type names.
var ValidElementTypes = map[string]struct{}{
	// --- Motivation layer ---
	"Stakeholder": {}, "Driver": {}, "Assessment": {}, "Goal": {},
	"Outcome": {}, "Principle": {}, "Requirement": {}, "Constraint": {},
	"Meaning": {}, "Value": {},

	// --- Strategy layer ---
	"Resource": {}, "Capability": {}, "CourseOfAction": {}, "ValueStream": {},

	// --- Business layer ---
	"BusinessActor": {}, "BusinessRole": {}, "BusinessCollaboration": {},
	"BusinessInterface": {}, "BusinessProcess": {}, "BusinessFunction": {},
	"BusinessInteraction": {}, "BusinessEvent": {}, "BusinessService": {},
	"BusinessObject": {}, "Contract": {}, "Representation": {},
	"Product": {},

	// --- Application layer ---
	"ApplicationComponent": {}, "ApplicationCollaboration": {},
	"ApplicationInterface": {}, "ApplicationFunction": {},
	"ApplicationInteraction": {}, "ApplicationProcess": {},
	"ApplicationEvent": {}, "ApplicationService": {}, "DataObject": {},

	// --- Technology layer ---
	"Node": {}, "Device": {}, "SystemSoftware": {},
	"TechnologyCollaboration": {}, "TechnologyInterface": {},
	"Path": {}, "CommunicationNetwork": {}, "TechnologyFunction": {},
	"TechnologyProcess": {}, "TechnologyInteraction": {},
	"TechnologyEvent": {}, "TechnologyService": {}, "Artifact": {},

	// --- Physical layer ---
	"Equipment": {}, "Facility": {}, "DistributionNetwork": {}, "Material": {},

	// --- Implementation & Migration layer ---
	"WorkPackage": {}, "ImplementationEvent": {}, "Deliverable": {},
	"ImplementationPlatform": {}, "Gap": {}, "Plateau": {},

	// --- Composite elements ---
	"Grouping": {}, "Location": {},

	// --- Junction (used in views) ---
	"Junction": {}, "OrJunction": {}, "AndJunction": {},
}

// ValidRelationshipTypes contains all valid ArchiMate 3.2 relationship type names.
// Both the full names (canonical) and the short names used in AOEF XML are accepted.
var ValidRelationshipTypes = map[string]struct{}{
	// Canonical (long) names
	"AssociationRelationship":    {},
	"AccessRelationship":         {},
	"InfluenceRelationship":      {},
	"RealizationRelationship":    {},
	"ServingRelationship":        {},
	"AssignmentRelationship":     {},
	"AggregationRelationship":    {},
	"CompositionRelationship":    {},
	"FlowRelationship":           {},
	"TriggeringRelationship":     {},
	"SpecializationRelationship": {},
	// Short names used in AOEF xsi:type attributes
	"Association":    {},
	"Access":         {},
	"Influence":      {},
	"Realization":    {},
	"Serving":        {},
	"Assignment":     {},
	"Aggregation":    {},
	"Composition":    {},
	"Flow":           {},
	"Triggering":     {},
	"Specialization": {},
	// Generic fallback present in some AOEF files
	"Relationship": {},
}

// ElementLayer returns the ArchiMate layer for a given element type.
// Returns empty string for unknown types.
func ElementLayer(elementType string) string {
	switch elementType {
	case "Stakeholder", "Driver", "Assessment", "Goal", "Outcome",
		"Principle", "Requirement", "Constraint", "Meaning", "Value":
		return "Motivation"
	case "Resource", "Capability", "CourseOfAction", "ValueStream":
		return "Strategy"
	case "BusinessActor", "BusinessRole", "BusinessCollaboration",
		"BusinessInterface", "BusinessProcess", "BusinessFunction",
		"BusinessInteraction", "BusinessEvent", "BusinessService",
		"BusinessObject", "Contract", "Representation", "Product":
		return "Business"
	case "ApplicationComponent", "ApplicationCollaboration",
		"ApplicationInterface", "ApplicationFunction",
		"ApplicationInteraction", "ApplicationProcess",
		"ApplicationEvent", "ApplicationService", "DataObject":
		return "Application"
	case "Node", "Device", "SystemSoftware", "TechnologyCollaboration",
		"TechnologyInterface", "Path", "CommunicationNetwork",
		"TechnologyFunction", "TechnologyProcess", "TechnologyInteraction",
		"TechnologyEvent", "TechnologyService", "Artifact":
		return "Technology"
	case "Equipment", "Facility", "DistributionNetwork", "Material":
		return "Physical"
	case "WorkPackage", "ImplementationEvent", "Deliverable",
		"ImplementationPlatform", "Gap", "Plateau":
		return "ImplementationMigration"
	case "Grouping", "Location", "Junction", "OrJunction", "AndJunction":
		return "Composite"
	default:
		return ""
	}
}
