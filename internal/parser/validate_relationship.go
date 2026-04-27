package parser

import (
	"fmt"
	"strings"
)

// Aspect is the ArchiMate 3.1 structural aspect of an element.
type Aspect int

const (
	AspectActiveStructure Aspect = iota
	AspectBehavior
	AspectPassiveStructure
	AspectMotivation
	AspectComposite
	AspectUnknown
)

// elementAspects maps every ArchiMate 3.1 element type to its aspect.
var elementAspects = map[string]Aspect{
	// --- Active Structure ---
	"BusinessActor": AspectActiveStructure, "BusinessRole": AspectActiveStructure,
	"BusinessCollaboration": AspectActiveStructure, "BusinessInterface": AspectActiveStructure,
	"ApplicationComponent": AspectActiveStructure, "ApplicationCollaboration": AspectActiveStructure,
	"ApplicationInterface": AspectActiveStructure,
	"Node":                 AspectActiveStructure, "Device": AspectActiveStructure,
	"SystemSoftware": AspectActiveStructure, "TechnologyCollaboration": AspectActiveStructure,
	"TechnologyInterface": AspectActiveStructure, "Path": AspectActiveStructure,
	"CommunicationNetwork": AspectActiveStructure,
	"Resource":             AspectActiveStructure, // Strategy
	"Equipment":            AspectActiveStructure, "Facility": AspectActiveStructure,
	"DistributionNetwork":    AspectActiveStructure, // Physical
	"ImplementationPlatform": AspectActiveStructure,

	// --- Behavior ---
	"BusinessProcess": AspectBehavior, "BusinessFunction": AspectBehavior,
	"BusinessInteraction": AspectBehavior, "BusinessEvent": AspectBehavior,
	"BusinessService":    AspectBehavior,
	"ApplicationProcess": AspectBehavior, "ApplicationFunction": AspectBehavior,
	"ApplicationInteraction": AspectBehavior, "ApplicationEvent": AspectBehavior,
	"ApplicationService": AspectBehavior,
	"TechnologyProcess":  AspectBehavior, "TechnologyFunction": AspectBehavior,
	"TechnologyInteraction": AspectBehavior, "TechnologyEvent": AspectBehavior,
	"TechnologyService": AspectBehavior,
	"Capability":        AspectBehavior, "CourseOfAction": AspectBehavior,
	"ValueStream": AspectBehavior, // Strategy
	"WorkPackage": AspectBehavior, "ImplementationEvent": AspectBehavior,

	// --- Passive Structure ---
	"BusinessObject": AspectPassiveStructure, "Contract": AspectPassiveStructure,
	"Representation": AspectPassiveStructure,
	"DataObject":     AspectPassiveStructure, "Artifact": AspectPassiveStructure,
	"Material":    AspectPassiveStructure, // Physical
	"Deliverable": AspectPassiveStructure, "Gap": AspectPassiveStructure,

	// --- Motivation ---
	"Stakeholder": AspectMotivation, "Driver": AspectMotivation,
	"Assessment": AspectMotivation, "Goal": AspectMotivation,
	"Outcome": AspectMotivation, "Principle": AspectMotivation,
	"Requirement": AspectMotivation, "Constraint": AspectMotivation,
	"Meaning": AspectMotivation, "Value": AspectMotivation,

	// --- Composite ---
	"Grouping": AspectComposite, "Location": AspectComposite,
	"Plateau": AspectComposite, "Product": AspectComposite,
}

// ElementAspect returns the ArchiMate aspect for a given element type.
// Returns AspectUnknown for unrecognised types.
func ElementAspect(elementType string) Aspect {
	if a, ok := elementAspects[elementType]; ok {
		return a
	}
	return AspectUnknown
}

// RelationshipViolation describes a single ArchiMate structural rule violation.
// Violations are non-fatal warnings; callers decide whether to block or surface them.
type RelationshipViolation struct {
	Rule        string `json:"rule"`
	Description string `json:"description"`
}

// ValidateRelationship checks the ArchiMate 3.1 structural rules for a relationship
// between two element types. It returns violations (non-fatal warnings); an empty
// slice means the combination is valid according to the metamodel.
//
// Unknown element types are treated as valid so partial or custom models are not blocked.
func ValidateRelationship(sourceType, targetType, relType string) []RelationshipViolation {
	// Normalise to long form (handles both "Influence" and "InfluenceRelationship").
	if relType != "" && !strings.HasSuffix(relType, "Relationship") {
		relType += "Relationship"
	}

	srcAspect := ElementAspect(sourceType)
	tgtAspect := ElementAspect(targetType)

	// Unknown element types: skip validation to avoid blocking custom models.
	if srcAspect == AspectUnknown || tgtAspect == AspectUnknown {
		return nil
	}

	switch relType {
	case "AssignmentRelationship":
		return validateAssignment(sourceType, targetType, srcAspect, tgtAspect)
	case "AccessRelationship":
		return validateAccess(srcAspect, tgtAspect)
	case "FlowRelationship":
		return validateFlow(srcAspect, tgtAspect)
	case "TriggeringRelationship":
		return validateTriggering(srcAspect, tgtAspect)
	case "ServingRelationship":
		return validateServing(srcAspect, tgtAspect)
	case "RealizationRelationship":
		return validateRealization(srcAspect, tgtAspect)
	case "InfluenceRelationship":
		return validateInfluence(srcAspect, tgtAspect)
	case "SpecializationRelationship":
		return validateSpecialization(sourceType, targetType, srcAspect, tgtAspect)
	// Association, Composition, Aggregation: no structural restrictions.
	case "AssociationRelationship", "CompositionRelationship", "AggregationRelationship":
		return nil
	}
	return nil
}

// --- per-type helpers -------------------------------------------------------

func viol(rule, format string, args ...any) []RelationshipViolation {
	return []RelationshipViolation{{Rule: rule, Description: fmt.Sprintf(format, args...)}}
}

// Assignment: ActiveStructure → Behavior (primary) or ActiveStructure → ActiveStructure
// within the same layer (e.g. BusinessActor ↔ BusinessRole).
func validateAssignment(srcType, tgtType string, srcAspect, tgtAspect Aspect) []RelationshipViolation {
	if srcAspect != AspectActiveStructure {
		return viol("assignment-source",
			"Assignment source '%s' must be an active structure element (e.g. Actor, Component, Node)", srcType)
	}
	if tgtAspect == AspectBehavior {
		return nil
	}
	if tgtAspect == AspectActiveStructure {
		// Only valid in the Business layer where Actor/Role are distinct concepts
		// (e.g. BusinessActor ↔ BusinessRole). Application and Technology layers
		// have no equivalent Role element — use Composition/Aggregation instead.
		if ElementLayer(srcType) == "Business" && ElementLayer(tgtType) == "Business" {
			return nil
		}
		return viol("assignment-active-to-active",
			"Assignment between active structure elements '%s' and '%s' is only valid for business actor/role pairs", srcType, tgtType)
	}
	return viol("assignment-target",
		"Assignment target '%s' must be a behavior element (Process, Function, Service, etc.)", tgtType)
}

// Access: (Behavior | ActiveStructure) → PassiveStructure
func validateAccess(srcAspect, tgtAspect Aspect) []RelationshipViolation {
	if srcAspect != AspectBehavior && srcAspect != AspectActiveStructure {
		return viol("access-source",
			"Access relationship source must be a behavior or active structure element")
	}
	if tgtAspect != AspectPassiveStructure {
		return viol("access-target",
			"Access relationship target must be a passive structure element (BusinessObject, DataObject, Artifact, etc.)")
	}
	return nil
}

// Flow: (Behavior | PassiveStructure) → (Behavior | PassiveStructure)
func validateFlow(srcAspect, tgtAspect Aspect) []RelationshipViolation {
	ok := func(a Aspect) bool { return a == AspectBehavior || a == AspectPassiveStructure }
	if !ok(srcAspect) || !ok(tgtAspect) {
		return viol("flow-aspects",
			"Flow relationship requires behavior or passive structure elements on both ends")
	}
	return nil
}

// Triggering: Behavior → Behavior
func validateTriggering(srcAspect, tgtAspect Aspect) []RelationshipViolation {
	if srcAspect != AspectBehavior || tgtAspect != AspectBehavior {
		return viol("triggering-aspects",
			"Triggering relationship requires behavior elements on both ends (Process, Function, Event, etc.)")
	}
	return nil
}

// Serving: (ActiveStructure | Behavior) → (ActiveStructure | Behavior)
func validateServing(srcAspect, tgtAspect Aspect) []RelationshipViolation {
	structural := func(a Aspect) bool { return a == AspectActiveStructure || a == AspectBehavior }
	if !structural(srcAspect) || !structural(tgtAspect) {
		return viol("serving-aspects",
			"Serving relationship requires active structure or behavior elements on both ends")
	}
	return nil
}

// Realization: concrete element realizes an abstraction.
// Source: Behavior, ActiveStructure, or PassiveStructure.
// Target: Behavior, Motivation, ActiveStructure, or PassiveStructure (not Composite).
func validateRealization(srcAspect, tgtAspect Aspect) []RelationshipViolation {
	if srcAspect == AspectMotivation || srcAspect == AspectComposite {
		return viol("realization-source",
			"Realization source must be a concrete element (behavior, active structure, or passive structure), not motivational or composite")
	}
	if tgtAspect == AspectComposite {
		return viol("realization-target",
			"Realization target must not be a composite element")
	}
	return nil
}

// Influence: at least one side must be a motivation element.
func validateInfluence(srcAspect, tgtAspect Aspect) []RelationshipViolation {
	if srcAspect != AspectMotivation && tgtAspect != AspectMotivation {
		return viol("influence-motivation",
			"Influence relationship should involve at least one motivation element (Goal, Driver, Requirement, Constraint, etc.)")
	}
	return nil
}

// Specialization: source and target must share the same aspect.
func validateSpecialization(srcType, tgtType string, srcAspect, tgtAspect Aspect) []RelationshipViolation {
	if srcAspect != tgtAspect {
		return viol("specialization-aspect",
			"Specialization requires source and target to be of the same aspect ('%s' and '%s' differ)", srcType, tgtType)
	}
	return nil
}
