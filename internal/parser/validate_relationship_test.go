package parser

import (
	"testing"
)

func TestValidateRelationship(t *testing.T) {
	tests := []struct {
		name      string
		srcType   string
		tgtType   string
		relType   string
		wantValid bool
		wantRule  string // expected violation rule when wantValid=false
	}{
		// --- Assignment ---
		{name: "assignment actor→process", srcType: "BusinessActor", tgtType: "BusinessProcess", relType: "AssignmentRelationship", wantValid: true},
		{name: "assignment component→function", srcType: "ApplicationComponent", tgtType: "ApplicationFunction", relType: "AssignmentRelationship", wantValid: true},
		{name: "assignment node→service", srcType: "Node", tgtType: "TechnologyService", relType: "AssignmentRelationship", wantValid: true},
		{name: "assignment actor↔role same layer", srcType: "BusinessActor", tgtType: "BusinessRole", relType: "AssignmentRelationship", wantValid: true},
		{name: "assignment component→component (illegal)", srcType: "ApplicationComponent", tgtType: "ApplicationComponent", relType: "AssignmentRelationship", wantValid: false, wantRule: "assignment-active-to-active"},
		{name: "assignment component→node cross-layer (illegal)", srcType: "ApplicationComponent", tgtType: "Node", relType: "AssignmentRelationship", wantValid: false, wantRule: "assignment-active-to-active"},
		{name: "assignment process→function (illegal src)", srcType: "BusinessProcess", tgtType: "BusinessFunction", relType: "AssignmentRelationship", wantValid: false, wantRule: "assignment-source"},
		// short-form type should also work
		{name: "assignment short form", srcType: "BusinessRole", tgtType: "BusinessProcess", relType: "Assignment", wantValid: true},

		// --- Access ---
		{name: "access function→dataobject", srcType: "ApplicationFunction", tgtType: "DataObject", relType: "AccessRelationship", wantValid: true},
		{name: "access component→artifact", srcType: "ApplicationComponent", tgtType: "Artifact", relType: "AccessRelationship", wantValid: true},
		{name: "access component→component (illegal)", srcType: "ApplicationComponent", tgtType: "ApplicationComponent", relType: "AccessRelationship", wantValid: false, wantRule: "access-target"},
		{name: "access dataobject→function (illegal src)", srcType: "DataObject", tgtType: "ApplicationFunction", relType: "AccessRelationship", wantValid: false, wantRule: "access-source"},

		// --- Serving ---
		{name: "serving app→business", srcType: "ApplicationService", tgtType: "BusinessProcess", relType: "ServingRelationship", wantValid: true},
		{name: "serving component→component", srcType: "ApplicationComponent", tgtType: "ApplicationComponent", relType: "ServingRelationship", wantValid: true},
		{name: "serving dataobject→function (illegal)", srcType: "DataObject", tgtType: "ApplicationFunction", relType: "ServingRelationship", wantValid: false, wantRule: "serving-aspects"},

		// --- Flow ---
		{name: "flow process→process", srcType: "BusinessProcess", tgtType: "BusinessProcess", relType: "FlowRelationship", wantValid: true},
		{name: "flow process→dataobject", srcType: "BusinessProcess", tgtType: "BusinessObject", relType: "FlowRelationship", wantValid: true},
		{name: "flow component→process (illegal)", srcType: "ApplicationComponent", tgtType: "ApplicationProcess", relType: "FlowRelationship", wantValid: false, wantRule: "flow-aspects"},

		// --- Triggering ---
		{name: "triggering process→event", srcType: "BusinessProcess", tgtType: "BusinessEvent", relType: "TriggeringRelationship", wantValid: true},
		{name: "triggering component→process (illegal)", srcType: "ApplicationComponent", tgtType: "ApplicationProcess", relType: "TriggeringRelationship", wantValid: false, wantRule: "triggering-aspects"},

		// --- Realization ---
		{name: "realization appservice→bizservice", srcType: "ApplicationService", tgtType: "BusinessService", relType: "RealizationRelationship", wantValid: true},
		{name: "realization behavior→goal", srcType: "BusinessProcess", tgtType: "Goal", relType: "RealizationRelationship", wantValid: true},
		{name: "realization goal→process (illegal src)", srcType: "Goal", tgtType: "BusinessProcess", relType: "RealizationRelationship", wantValid: false, wantRule: "realization-source"},

		// --- Influence ---
		{name: "influence driver→goal", srcType: "Driver", tgtType: "Goal", relType: "InfluenceRelationship", wantValid: true},
		{name: "influence goal→process", srcType: "Goal", tgtType: "BusinessProcess", relType: "InfluenceRelationship", wantValid: true},
		{name: "influence component→function (illegal)", srcType: "ApplicationComponent", tgtType: "ApplicationFunction", relType: "InfluenceRelationship", wantValid: false, wantRule: "influence-motivation"},

		// --- Specialization ---
		{name: "specialization same aspect", srcType: "ApplicationComponent", tgtType: "ApplicationComponent", relType: "SpecializationRelationship", wantValid: true},
		{name: "specialization different aspects (illegal)", srcType: "ApplicationComponent", tgtType: "ApplicationFunction", relType: "SpecializationRelationship", wantValid: false, wantRule: "specialization-aspect"},

		// --- Association / Composition / Aggregation: always valid ---
		{name: "association any→any", srcType: "ApplicationComponent", tgtType: "Goal", relType: "AssociationRelationship", wantValid: true},
		{name: "composition component→function", srcType: "ApplicationComponent", tgtType: "ApplicationFunction", relType: "CompositionRelationship", wantValid: true},
		{name: "aggregation node→artifact", srcType: "Node", tgtType: "Artifact", relType: "AggregationRelationship", wantValid: true},

		// --- Unknown element types: should not produce violations ---
		{name: "unknown source type", srcType: "CustomWidget", tgtType: "BusinessProcess", relType: "AssignmentRelationship", wantValid: true},
		{name: "unknown target type", srcType: "BusinessActor", tgtType: "CustomWidget", relType: "AssignmentRelationship", wantValid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := ValidateRelationship(tt.srcType, tt.tgtType, tt.relType)
			isValid := len(violations) == 0
			if isValid != tt.wantValid {
				t.Errorf("ValidateRelationship(%q, %q, %q) valid=%v, want %v (violations: %v)",
					tt.srcType, tt.tgtType, tt.relType, isValid, tt.wantValid, violations)
				return
			}
			if !tt.wantValid && tt.wantRule != "" {
				if violations[0].Rule != tt.wantRule {
					t.Errorf("expected rule %q, got %q (%s)", tt.wantRule, violations[0].Rule, violations[0].Description)
				}
			}
		})
	}
}
