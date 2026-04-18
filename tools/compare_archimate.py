#!/usr/bin/env python3
"""
Semantic ArchiMate XML comparator.
Compares two ArchiMate 3.0 Open Exchange Format files by ArchiMate concepts,
not as raw text. Reports elements, relationships, views, and property definitions
that are added, removed, or changed between file A and file B.

Usage:
    python3 tools/compare_archimate.py <file_a.xml> <file_b.xml>
"""

import sys
import xml.etree.ElementTree as ET
from collections import defaultdict

NS = "http://www.opengroup.org/xsd/archimate/3.0/"
XSI = "http://www.w3.org/2001/XMLSchema-instance"

# ANSI colors
RED    = "\033[91m"
GREEN  = "\033[92m"
YELLOW = "\033[93m"
CYAN   = "\033[96m"
BOLD   = "\033[1m"
RESET  = "\033[0m"

def tag(local):
    return f"{{{NS}}}{local}"

def xsi_type(el):
    return el.get(f"{{{XSI}}}type", "")

def get_names(el):
    """Return dict of {lang: name_text} from <name xml:lang="..."> children."""
    names = {}
    for child in el:
        if child.tag == tag("name"):
            lang = child.get("{http://www.w3.org/XML/1998/namespace}lang", "")
            names[lang] = (child.text or "").strip()
    return names

def get_properties(el):
    """Return dict of {propertyDefinitionRef: value} from <properties><property> children."""
    props = {}
    for props_el in el:
        if props_el.tag == tag("properties"):
            for prop in props_el:
                if prop.tag == tag("property"):
                    ref = prop.get("propertyDefinitionRef", "")
                    val_el = prop.find(tag("value"))
                    val = (val_el.text or "").strip() if val_el is not None else ""
                    props[ref] = val
    return props

def parse_property_definitions(root):
    result = {}
    for defs in root.iter(tag("propertyDefinitions")):
        for pd in defs:
            if pd.tag == tag("propertyDefinition"):
                pid = pd.get("identifier", "")
                dtype = pd.get("type", "")
                names = get_names(pd)
                result[pid] = {"type": dtype, "names": names}
    return result

def parse_elements(root):
    result = {}
    for els in root.iter(tag("elements")):
        for el in els:
            if el.tag == tag("element"):
                eid = el.get("identifier", "")
                result[eid] = {
                    "type": xsi_type(el),
                    "names": get_names(el),
                    "properties": get_properties(el),
                    "documentation": _get_documentation(el),
                }
    return result

def _get_documentation(el):
    for child in el:
        if child.tag == tag("documentation"):
            lang = child.get("{http://www.w3.org/XML/1998/namespace}lang", "")
            return {lang: (child.text or "").strip()}
    return {}

def parse_relationships(root):
    result = {}
    for rels in root.iter(tag("relationships")):
        for rel in rels:
            if rel.tag == tag("relationship"):
                rid = rel.get("identifier", "")
                result[rid] = {
                    "type": xsi_type(rel),
                    "source": rel.get("source", ""),
                    "target": rel.get("target", ""),
                    "accessType": rel.get("accessType", ""),
                    "names": get_names(rel),
                    "properties": get_properties(rel),
                }
    return result

def parse_views(root):
    result = {}
    for views in root.iter(tag("views")):
        for diagrams in views:
            if diagrams.tag == tag("diagrams"):
                for view in diagrams:
                    if view.tag == tag("view"):
                        vid = view.get("identifier", "")
                        result[vid] = {
                            "names": get_names(view),
                            "viewpoint": view.get("viewpoint", ""),
                            "nodes": _parse_nodes(view),
                            "connections": _parse_connections(view),
                        }
    return result

def _parse_nodes(view):
    nodes = {}
    for node in view.iter(tag("node")):
        nid = node.get("identifier", "")
        nodes[nid] = {
            "elementRef": node.get("elementRef", ""),
            "type": xsi_type(node),
            "label": get_names(node),
        }
    return nodes

def _parse_connections(view):
    conns = {}
    for conn in view.iter(tag("connection")):
        cid = conn.get("identifier", "")
        conns[cid] = {
            "relationshipRef": conn.get("relationshipRef", ""),
            "source": conn.get("source", ""),
            "target": conn.get("target", ""),
            "type": xsi_type(conn),
        }
    return conns

def parse_file(path):
    tree = ET.parse(path)
    root = tree.getroot()
    model_id = root.get("identifier", "")
    model_names = get_names(root)
    return {
        "model_id": model_id,
        "model_names": model_names,
        "property_definitions": parse_property_definitions(root),
        "elements": parse_elements(root),
        "relationships": parse_relationships(root),
        "views": parse_views(root),
    }

# --- Reporting helpers ---

def section(title):
    print(f"\n{BOLD}{CYAN}{'='*60}{RESET}")
    print(f"{BOLD}{CYAN}  {title}{RESET}")
    print(f"{BOLD}{CYAN}{'='*60}{RESET}")

def subsection(title):
    print(f"\n{BOLD}{title}{RESET}")
    print(f"  {'-'*40}")

def added(msg):
    print(f"  {GREEN}+ {msg}{RESET}")

def removed(msg):
    print(f"  {RED}- {msg}{RESET}")

def changed(msg):
    print(f"  {YELLOW}~ {msg}{RESET}")

def fmt_names(names):
    if not names:
        return "(no name)"
    return " | ".join(f"[{k}] {v}" for k, v in sorted(names.items()))

# --- Comparators ---

def compare_model_meta(a, b):
    section("Model metadata")
    if a["model_id"] != b["model_id"]:
        removed(f"ID: {a['model_id']}")
        added(f"ID: {b['model_id']}")
    else:
        print(f"  Model ID: {a['model_id']} (same)")

    if a["model_names"] != b["model_names"]:
        changed(f"Name: {fmt_names(a['model_names'])} → {fmt_names(b['model_names'])}")
    else:
        print(f"  Model name: {fmt_names(a['model_names'])} (same)")


def compare_dicts_simple(label, a_dict, b_dict, describe_fn):
    """Generic comparison for flat dicts of {id -> info}."""
    only_a = set(a_dict) - set(b_dict)
    only_b = set(b_dict) - set(a_dict)
    both = set(a_dict) & set(b_dict)

    diffs = [(k, a_dict[k], b_dict[k]) for k in both if a_dict[k] != b_dict[k]]

    print(f"\n  Only in A ({len(only_a)}):  Only in B ({len(only_b)}):  Changed ({len(diffs)}):")
    for k in sorted(only_a):
        removed(f"[{k}] {describe_fn(a_dict[k])}")
    for k in sorted(only_b):
        added(f"[{k}] {describe_fn(b_dict[k])}")
    for k, va, vb in sorted(diffs):
        changed(f"[{k}]")
        diff_fields(va, vb, indent=6)

    if not only_a and not only_b and not diffs:
        print(f"  (identical)")
    return len(only_a), len(only_b), len(diffs)

def diff_fields(va, vb, indent=6):
    pad = " " * indent
    for field in set(list(va.keys()) + list(vb.keys())):
        fa = va.get(field)
        fb = vb.get(field)
        if fa != fb:
            print(f"{pad}{YELLOW}{field}:{RESET}")
            print(f"{pad}  {RED}- {fa}{RESET}")
            print(f"{pad}  {GREEN}+ {fb}{RESET}")

def compare_property_definitions(a, b):
    section("Property Definitions")
    def desc(pd):
        return f"type={pd['type']} name={fmt_names(pd['names'])}"
    compare_dicts_simple("PropertyDefinition", a["property_definitions"], b["property_definitions"], desc)

def compare_elements(a, b):
    section("Elements")
    def desc(el):
        return f"{el['type']} — {fmt_names(el['names'])}"
    only_a_n, only_b_n, diff_n = compare_dicts_simple("Element", a["elements"], b["elements"], desc)

    # Summary by type for removed/added
    if only_a_n or only_b_n:
        subsection("Added/Removed breakdown by type")
        only_a = set(a["elements"]) - set(b["elements"])
        only_b = set(b["elements"]) - set(a["elements"])
        by_type = defaultdict(lambda: {"removed": [], "added": []})
        for k in only_a:
            by_type[a["elements"][k]["type"]]["removed"].append(k)
        for k in only_b:
            by_type[b["elements"][k]["type"]]["added"].append(k)
        for t in sorted(by_type):
            r = len(by_type[t]["removed"])
            ad = len(by_type[t]["added"])
            parts = []
            if r: parts.append(f"{RED}-{r}{RESET}")
            if ad: parts.append(f"{GREEN}+{ad}{RESET}")
            print(f"    {t}: {' '.join(parts)}")

def compare_relationships(a, b):
    section("Relationships")
    def desc(rel):
        return f"{rel['type']} {rel['source']} → {rel['target']}"
    compare_dicts_simple("Relationship", a["relationships"], b["relationships"], desc)

def compare_views(a, b):
    section("Views / Diagrams")
    only_a = set(a["views"]) - set(b["views"])
    only_b = set(b["views"]) - set(a["views"])
    both = set(a["views"]) & set(b["views"])

    print(f"\n  Only in A: {len(only_a)}  |  Only in B: {len(only_b)}  |  In both: {len(both)}")

    for k in sorted(only_a):
        v = a["views"][k]
        removed(f"[{k}] {fmt_names(v['names'])} ({len(v['nodes'])} nodes, {len(v['connections'])} connections)")

    for k in sorted(only_b):
        v = b["views"][k]
        added(f"[{k}] {fmt_names(v['names'])} ({len(v['nodes'])} nodes, {len(v['connections'])} connections)")

    changed_views = 0
    for k in sorted(both):
        va = a["views"][k]
        vb = b["views"][k]
        name_a = fmt_names(va["names"])
        name_b = fmt_names(vb["names"])

        view_diffs = []
        if va["names"] != vb["names"]:
            view_diffs.append(f"name: {name_a!r} → {name_b!r}")
        if va["viewpoint"] != vb["viewpoint"]:
            view_diffs.append(f"viewpoint: {va['viewpoint']!r} → {vb['viewpoint']!r}")

        nodes_only_a = set(va["nodes"]) - set(vb["nodes"])
        nodes_only_b = set(vb["nodes"]) - set(va["nodes"])
        nodes_changed = [nk for nk in set(va["nodes"]) & set(vb["nodes"]) if va["nodes"][nk] != vb["nodes"][nk]]

        conns_only_a = set(va["connections"]) - set(vb["connections"])
        conns_only_b = set(vb["connections"]) - set(va["connections"])

        if nodes_only_a or nodes_only_b or nodes_changed or conns_only_a or conns_only_b or view_diffs:
            changed_views += 1
            changed(f"[{k}] {name_a}")
            for d in view_diffs:
                print(f"      {YELLOW}{d}{RESET}")
            if nodes_only_a:
                print(f"      {RED}nodes removed: {len(nodes_only_a)}{RESET}")
            if nodes_only_b:
                print(f"      {GREEN}nodes added:   {len(nodes_only_b)}{RESET}")
            if nodes_changed:
                print(f"      {YELLOW}nodes changed: {len(nodes_changed)}{RESET}")
            if conns_only_a:
                print(f"      {RED}connections removed: {len(conns_only_a)}{RESET}")
            if conns_only_b:
                print(f"      {GREEN}connections added:   {len(conns_only_b)}{RESET}")

    if not only_a and not only_b and changed_views == 0:
        print("  (identical)")

def print_summary(a, b):
    section("Summary")
    for label, key in [("Property definitions", "property_definitions"), ("Elements", "elements"),
                       ("Relationships", "relationships"), ("Views", "views")]:
        na = len(a[key])
        nb = len(b[key])
        only_a = len(set(a[key]) - set(b[key]))
        only_b = len(set(b[key]) - set(a[key]))
        changed_n = sum(1 for k in set(a[key]) & set(b[key]) if a[key][k] != b[key][k])
        status = "SAME" if (only_a == 0 and only_b == 0 and changed_n == 0) else "DIFFERS"
        color = GREEN if status == "SAME" else YELLOW
        print(f"  {color}{label:<25}{RESET}  A={na}  B={nb}  {RED}-{only_a}{RESET}  {GREEN}+{only_b}{RESET}  {YELLOW}~{changed_n}{RESET}  [{color}{status}{RESET}]")

def main():
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <file_a.xml> <file_b.xml>")
        sys.exit(1)

    path_a, path_b = sys.argv[1], sys.argv[2]
    print(f"\n{BOLD}Comparing:{RESET}")
    print(f"  A: {path_a}")
    print(f"  B: {path_b}")

    try:
        a = parse_file(path_a)
    except Exception as e:
        print(f"{RED}Error parsing A: {e}{RESET}")
        sys.exit(1)
    try:
        b = parse_file(path_b)
    except Exception as e:
        print(f"{RED}Error parsing B: {e}{RESET}")
        sys.exit(1)

    compare_model_meta(a, b)
    print_summary(a, b)
    compare_property_definitions(a, b)
    compare_elements(a, b)
    compare_relationships(a, b)
    compare_views(a, b)

if __name__ == "__main__":
    main()
