#!/usr/bin/env python3
"""Make a compat contract declare exactly the legacy routes onion serves.

Reads the route table written by `ROUTE_DUMP=<file> go test ./cmd/api -run
TestDumpLegacyRoutes`, then rewrites <compat>/service.onk so that:
- methods whose route has no legacy handler are removed,
- every route under the given prefixes that no contract declares yet is
  added, with request types chosen from its verb and path parameters. Raw
  handlers and JSON routes that run inside route middlewares are added as
  @http methods, which hand the request to that handler unchanged,
- existing methods keep their names and request types.

Usage: scripts/sync_compat_contract.py <route-table> <compat-pkg-dir> <prefix>...
"""
import collections
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parents[1]

# Routes the global request body limit exempts: multipart CV uploads whose
# handlers cap the body themselves (5 MiB plus form overhead).
LARGE_BODY_ROUTES = {"/careers/apply", "/voxhire/applications", "/voxhire/careers/apply"}
API = ROOT.parent.parent / "voxie" / "api"


def norm(path: str) -> str:
    return re.sub(r"\{[^}]+\}", "{}", path.rstrip("/") or "/")


def contract_routes(exclude: pathlib.Path) -> set:
    routes = set()
    for f in API.glob("**/service.onk"):
        if f == exclude:
            continue
        s = f.read_text()
        package = re.search(r"package ([\w.]+)", s).group(1)
        for svc in re.finditer(r"service (\w+) \{(.*?)\n\}", s, re.S):
            body = svc.group(2)
            base = re.search(r'base_path:\s*"([^"]*)"', body)
            base = base.group(1) if base else "/" + package.replace(".", "/")
            for m in re.finditer(r'\w+\(\w+\)\s*->\s*\w+\s*@(get|post|put|patch|delete)\("([^"]*)"\)', body):
                full = re.sub(r"/+", "/", base.rstrip("/") + m.group(2)) if m.group(2) else base
                routes.add((m.group(1).upper(), norm(full)))
    return routes


def method_name(verb: str, path: str, base: str, taken: set) -> str:
    rel = path[len(base):] if path.startswith(base) else path
    words = [w for w in re.split(r"[/\-_]", rel) if w and not w.startswith("{")]
    has_param = "{" in rel
    prefix = {"GET": "get" if has_param else "list", "POST": "create", "PUT": "update", "PATCH": "patch", "DELETE": "delete"}[verb]
    if verb == "POST" and words and words[-1] in {"approve", "reject", "submit", "complete", "publish", "export", "accept", "verify", "trigger", "sync", "role", "upsert"}:
        prefix = words.pop()
    name = prefix + "".join(w[:1].upper() + w[1:] for w in words) or prefix + "Root"
    candidate, n = name, 2
    while candidate in taken:
        candidate, n = f"{name}{n}", n + 1
    taken.add(candidate)
    return candidate


def request_type(verb: str, path: str) -> tuple[str, str]:
    params = re.findall(r"\{(\w+)\}", path)
    body_verb = verb in {"POST", "PUT", "PATCH"}
    if not params:
        return ("BodyRequest", ' @body("body")') if body_verb else ("EmptyRequest", "")
    name = "Path" + "".join(p.title().replace("_", "") for p in params) + ("BodyRequest" if body_verb else "Request")
    return name, (' @body("body")' if body_verb else "")


def main() -> None:
    table_file, pkg_dir, *prefixes = sys.argv[1:]
    pkg = API / pkg_dir
    service_file, models_file = pkg / "service.onk", pkg / "models.onk"
    service, models = service_file.read_text(), models_file.read_text()

    served, kinds = {}, {}
    for line in open(table_file):
        kind, route = line.rstrip("\n").split("\t")
        verb, path = route.split(" ", 1)
        served[(verb, norm(path))] = path
        kinds[(verb, norm(path))] = kind

    other = contract_routes(service_file)
    svc_match = re.search(r"(service \w+ \{)(.*?)(\n\})", service, re.S)
    body = svc_match.group(2)
    base = re.search(r'base_path:\s*"([^"]*)"', body).group(1)

    kept, declared, removed = [], set(), []
    taken = set()
    for line in body.split("\n"):
        m = re.match(r'\s*(\w+)\(\w+\)\s*->\s*\w+\s*@(get|post|put|patch|delete)\("([^"]*)"\)', line)
        if not m:
            kept.append(line)
            continue
        full = re.sub(r"/+", "/", base.rstrip("/") + m.group(3))
        key = (m.group(2).upper(), norm(full))
        if key in served:
            kept.append(line)
            declared.add(key)
            taken.add(m.group(1))
        else:
            removed.append(m.group(1))

    added, needed = [], collections.OrderedDict()
    for key in sorted(served):
        verb, npath = key
        path = served[key]
        if key in declared or key in other or not any(path == p or path.startswith(p.rstrip("/") + "/") for p in prefixes):
            continue
        if not path.startswith(base):
            raise SystemExit(f"{path} is outside base_path {base}")
        params = re.findall(r"\{(\w+)\}", path)
        if kinds[key] == "json":
            req, tags = request_type(verb, path)
            if req not in ("BodyRequest", "EmptyRequest"):
                needed[req] = (params, verb in {"POST", "PUT", "PATCH"})
        else:
            # The handler reads the raw request itself.
            req = "Path" + "".join(p.title().replace("_", "") for p in params) + "Request" if params else "EmptyRequest"
            if params:
                needed[req] = (params, False)
            tags = " @http"
            if path in LARGE_BODY_ROUTES:
                tags += ' @max_body("16MiB")'
        name = method_name(verb, path, base, taken)
        added.append(f'  {name}({req}) -> JsonResponse @{verb.lower()}("{path[len(base):]}"){tags}')

    body = "\n".join(kept).rstrip()
    if added:
        body += "\n\n  // Added from onion's route table.\n" + "\n".join(added)
    service_file.write_text(service[: svc_match.start(2)] + body + service[svc_match.end(2):])

    for msg, (params, with_body) in needed.items():
        if re.search(rf"message {msg}\b", models):
            continue
        fields = "".join(f"  {p}: string\n" for p in params) + ("  body: json\n" if with_body else "")
        models = models.rstrip() + f"\n\nmessage {msg} {{\n{fields}}}\n"
    for msg in ("EmptyRequest", "BodyRequest"):
        if not re.search(rf"message {msg}\b", models):
            models = models.rstrip() + (f"\n\nmessage {msg} {{}}\n" if msg == "EmptyRequest" else "\n\nmessage BodyRequest {\n  body: json\n}\n")
    models_file.write_text(models)
    print(f"removed {len(removed)}: {', '.join(removed)}")
    print(f"added {len(added)}")


if __name__ == "__main__":
    main()
