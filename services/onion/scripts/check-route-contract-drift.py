#!/usr/bin/env python3
"""Check the hand-served Flow and HR routes against their .onk contracts."""

from __future__ import annotations

import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def route_key(method: str, path: str) -> tuple[str, str]:
    return method.upper(), path


def join_route(base_path: str, route_path: str) -> str:
    return f"{base_path.rstrip('/')}/{route_path.lstrip('/')}" or "/"


def parse_contract_routes(path: Path, service_name: str) -> set[tuple[str, str]]:
    text = path.read_text(encoding="utf-8")
    service_match = re.search(
        rf"service\s+{re.escape(service_name)}\s*\{{(?P<body>.*?)\n\}}",
        text,
        re.DOTALL,
    )
    if service_match is None:
        raise ValueError(f"service {service_name} not found in {path}")

    body = service_match.group("body")
    base_match = re.search(r'base_path:\s*"([^"]+)"', body)
    base_path = base_match.group(1) if base_match else ""
    route_pattern = re.compile(
        r"^\s*\w+\([^)]*\)\s*->.*?"
        r"@(?P<method>get|post|put|patch|delete)\(\"(?P<path>[^\"]+)\"\)",
        re.MULTILINE,
    )
    return {
        route_key(match.group("method"), join_route(base_path, match.group("path")))
        for match in route_pattern.finditer(body)
    }


def parse_handler_routes(path: Path) -> dict[str, list[str]]:
    text = path.read_text(encoding="utf-8")
    definitions: dict[str, list[str]] = {}
    one_line_pattern = re.compile(
        r"func\s+(?P<name>\w+Routes)\(\)\s+\[\]string\s*\{"
        r"(?P<body>.*)\}\s*\}",
        re.MULTILINE,
    )
    for match in one_line_pattern.finditer(text):
        definitions[match.group("name")] = re.findall(
            r'"([^"\n]+)"', match.group("body")
        )

    multi_line_pattern = re.compile(
        r"func\s+(?P<name>\w+Routes)\(\)\s+\[\]string\s*\{"
        r"[ \t]*\n(?P<body>.*?)(?:\n\s*\})",
        re.DOTALL,
    )
    for match in multi_line_pattern.finditer(text):
        definitions[match.group("name")] = re.findall(
            r'"([^"\n]+)"', match.group("body")
        )
    return definitions


def parse_registered_hr_routes() -> set[tuple[str, str]]:
    routes_path = ROOT / "internal/modules/hr_time/internal/api/routes.go"
    text = routes_path.read_text(encoding="utf-8")
    handlers = {}
    for handler_path in routes_path.parent.glob("*.go"):
        if handler_path.name.endswith("_test.go"):
            continue
        handlers.update(parse_handler_routes(handler_path))

    registration_pattern = re.compile(
        r"serverHttp\.Register(?P<method>GET|POST|PUT|PATCH|DELETE)"
        r"(?:WithValidation)?\[[^\]]+\]\(\s*ctx,\s*router,\s*"
        r"(?:handlers\.|api\.)?(?P<handler>\w+)Routes\(\)",
        re.DOTALL,
    )
    registered = set()
    for match in registration_pattern.finditer(text):
        route_name = f"{match.group('handler')}Routes"
        paths = handlers.get(route_name)
        if paths is None:
            raise ValueError(f"route function {route_name} is not defined")
        registered.update(
            route_key(match.group("method"), path) for path in paths
        )
    if not registered:
        raise ValueError(f"no HR routes found in {routes_path}")
    return registered


def parse_registered_flow_escape_hatches() -> set[tuple[str, str]]:
    module_path = ROOT / "internal/modules/workspaces/module.go"
    text = module_path.read_text(encoding="utf-8")
    pattern = re.compile(
        r"\brouter\.(?P<method>Get|Post|Put|Patch|Delete)\(\s*"
        r'"(?P<path>/flow[^"\n]*)"'
    )
    return {
        route_key(match.group("method"), match.group("path"))
        for match in pattern.finditer(text)
    }


def find_stale_route_literals() -> list[str]:
    stale = (
        "/flow/presence/today",
        "/flow/insights",
        "/flow/team/workload",
        "/flow/roadmap",
        "/hr/time-entries/import-flow",
        "/hr/time-entries/export-billable",
    )
    findings = []
    for path in (ROOT / "internal/modules").rglob("*.go"):
        if path.name.endswith("_test.go"):
            continue
        text = path.read_text(encoding="utf-8")
        for route in stale:
            if re.search(rf"{re.escape(route)}(?=[\"'`]|$|\}})", text):
                findings.append(f"{route} in {path.relative_to(ROOT)}")
    return findings


def format_routes(routes: set[tuple[str, str]]) -> str:
    return ", ".join(f"{method} {path}" for method, path in sorted(routes))


def main() -> int:
    failures = []

    try:
        registered_hr = parse_registered_hr_routes()
        declared_hr = parse_contract_routes(
            ROOT / "voxie/api/hr/compat/v1/service.onk", "CompatibilityService"
        )
        missing_hr = registered_hr - declared_hr
        if missing_hr:
            failures.append(
                "HR routes missing from hr.compat.v1: " + format_routes(missing_hr)
            )

        registered_flow = parse_registered_flow_escape_hatches()
        expected_flow = {
            route_key("GET", "/flow/ws"),
            route_key("POST", "/flow/ws-ticket"),
        }
        if registered_flow != expected_flow:
            failures.append(
                "Flow raw routes changed. Expected "
                f"{format_routes(expected_flow)}; found {format_routes(registered_flow)}"
            )

        declared_realtime = parse_contract_routes(
            ROOT / "voxie/api/flow/realtime/v1/service.onk", "RealtimeService"
        )
        if route_key("POST", "/flow/ws-ticket") not in declared_realtime:
            failures.append("POST /flow/ws-ticket is missing from flow.realtime.v1")
        if route_key("GET", "/flow/ws") in declared_realtime:
            failures.append("GET /flow/ws must remain a documented transport escape hatch")

        stale_routes = find_stale_route_literals()
        if stale_routes:
            failures.append("stale route literals found: " + "; ".join(stale_routes))
    except (OSError, ValueError) as error:
        failures.append(str(error))

    if failures:
        for failure in failures:
            print(f"route contract check failed: {failure}", file=sys.stderr)
        return 1

    print(
        "route contract check passed: "
        f"{len(parse_registered_hr_routes())} HR routes and "
        "2 documented Flow transport routes"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
