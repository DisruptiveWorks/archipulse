package mcpserver

// aoefFormatGuide is an AOEF generation guide for Claude.
// Exposed as an MCP resource so Claude can read it before generating XML.
const aoefFormatGuide = `# AOEF — ArchiMate Open Exchange Format

ArchiPulse imports and exports ArchiMate models as AOEF XML. This guide covers everything needed to generate a valid file.

## Root element

` + "```xml" + `
<?xml version="1.0" encoding="UTF-8"?>
<model xmlns="http://www.opengroup.org/xsd/archimate/3.0/"
       xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
       xsi:schemaLocation="http://www.opengroup.org/xsd/archimate/3.0/ http://www.opengroup.org/xsd/archimate/3.1/archimate3_Diagram.xsd"
       identifier="id-model-001"
       version="1.0">
  <name>Model Name</name>
  ...
</model>
` + "```" + `

Rules:
- ` + "`identifier`" + ` must be unique across the file; use ` + "`id-<slug>`" + ` convention.
- ` + "`<name>`" + ` is plain text (no lang attribute required, but you may add ` + "`xml:lang=\"en\"`" + `).

---

## Elements

` + "```xml" + `
<elements>
  <element identifier="id-elem-001" type="ApplicationComponent">
    <name>CRM System</name>
    <documentation>Manages customer relationships and interactions.</documentation>
    <properties>
      <property propertyDefinitionRef="pd-lifecycle">
        <value>Production</value>
      </property>
    </properties>
  </element>
</elements>
` + "```" + `

### Valid element types by layer

**Motivation:** Stakeholder, Driver, Assessment, Goal, Outcome, Principle, Requirement, Constraint, Meaning, Value

**Strategy:** Resource, Capability, CourseOfAction, ValueStream

**Business:** BusinessActor, BusinessRole, BusinessCollaboration, BusinessInterface, BusinessProcess, BusinessFunction, BusinessInteraction, BusinessEvent, BusinessService, BusinessObject, Contract, Representation, Product

**Application:** ApplicationComponent, ApplicationCollaboration, ApplicationInterface, ApplicationFunction, ApplicationInteraction, ApplicationProcess, ApplicationEvent, ApplicationService, DataObject

**Technology:** Node, Device, SystemSoftware, TechnologyCollaboration, TechnologyInterface, Path, CommunicationNetwork, TechnologyFunction, TechnologyProcess, TechnologyInteraction, TechnologyEvent, TechnologyService, Artifact

**Physical:** Equipment, Facility, DistributionNetwork, Material

**Implementation & Migration:** WorkPackage, ImplementationEvent, Deliverable, ImplementationPlatform, Gap, Plateau

**Composite:** Grouping, Location

---

## Relationships

` + "```xml" + `
<relationships>
  <relationship identifier="id-rel-001"
                type="ServingRelationship"
                source="id-elem-001"
                target="id-elem-002">
    <name>provides data to</name>
  </relationship>
</relationships>
` + "```" + `

### Valid relationship types

| Type | Use when |
|---|---|
| AssociationRelationship | generic unspecified link |
| AccessRelationship | element accesses a data object (add ` + "`accessType=\"Read\"`" + ` / ` + "`Write`" + ` / ` + "`ReadWrite`" + `) |
| InfluenceRelationship | motivational influence (add ` + "`modifier=\"+\"`" + ` or ` + "`\"-\"`" + `) |
| RealizationRelationship | lower-layer element realizes higher-layer concept |
| ServingRelationship | element provides a service to another |
| AssignmentRelationship | actor/role assigned to behaviour |
| AggregationRelationship | whole–part (parts can exist independently) |
| CompositionRelationship | whole–part (parts cannot exist independently) |
| FlowRelationship | information or material flow between behaviours |
| TriggeringRelationship | one behaviour triggers another |
| SpecializationRelationship | specialization of a concept |

Short names (` + "`Serving`" + `, ` + "`Realization`" + `, etc.) are also accepted.

---

## Property definitions (optional)

Declare once at the model level, then reference by ID in elements.

` + "```xml" + `
<propertyDefinitions>
  <propertyDefinition identifier="pd-lifecycle" type="string">
    <name>Lifecycle Status</name>
  </propertyDefinition>
  <propertyDefinition identifier="pd-owner" type="string">
    <name>Owner</name>
  </propertyDefinition>
</propertyDefinitions>
` + "```" + `

---

## Views (diagrams)

Views are optional. Omit them if you only want to import elements and relationships.

` + "```xml" + `
<views>
  <diagrams>
    <view identifier="id-view-001" viewpoint="Application Usage">
      <name>Application Overview</name>
      <node identifier="id-node-001" elementRef="id-elem-001"
            xsi:type="Element" x="100" y="100" w="120" h="55"/>
      <node identifier="id-node-002" elementRef="id-elem-002"
            xsi:type="Element" x="300" y="100" w="120" h="55"/>
      <connection identifier="id-conn-001"
                  relationshipRef="id-rel-001"
                  source="id-node-001"
                  target="id-node-002"/>
    </view>
  </diagrams>
</views>
` + "```" + `

Rules:
- ` + "`node.elementRef`" + ` must match an ` + "`element.identifier`" + `.
- ` + "`connection.relationshipRef`" + ` must match a ` + "`relationship.identifier`" + `.
- ` + "`connection.source`" + ` / ` + "`target`" + ` must match ` + "`node.identifier`" + ` values (not element IDs).
- Layout (x, y, w, h) is in pixels; use multiples of 10 for clean placement.
- ` + "`xsi:type`" + ` on nodes must be exactly ` + "`Element`" + `.

---

## Minimal complete example

` + "```xml" + `
<?xml version="1.0" encoding="UTF-8"?>
<model xmlns="http://www.opengroup.org/xsd/archimate/3.0/"
       xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
       identifier="id-model-example" version="1.0">
  <name>Example Model</name>

  <propertyDefinitions>
    <propertyDefinition identifier="pd-status" type="string">
      <name>Lifecycle Status</name>
    </propertyDefinition>
  </propertyDefinitions>

  <elements>
    <element identifier="id-cap-001" type="Capability">
      <name>Customer Management</name>
    </element>
    <element identifier="id-app-001" type="ApplicationComponent">
      <name>CRM System</name>
      <properties>
        <property propertyDefinitionRef="pd-status">
          <value>Production</value>
        </property>
      </properties>
    </element>
  </elements>

  <relationships>
    <relationship identifier="id-rel-001"
                  type="RealizationRelationship"
                  source="id-app-001"
                  target="id-cap-001"/>
  </relationships>
</model>
` + "```" + `

---

## Common mistakes to avoid

- Do NOT use duplicate ` + "`identifier`" + ` values anywhere in the file.
- Do NOT reference an ` + "`elementRef`" + ` or ` + "`source`" + `/` + "`target`" + ` that has no matching ` + "`identifier`" + `.
- Do NOT use relationship types not in the list above.
- Do NOT use element types not in the list above.
- ` + "`<documentation>`" + ` is plain text — no HTML or markdown inside it.
`
