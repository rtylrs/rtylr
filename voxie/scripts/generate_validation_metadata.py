#!/usr/bin/env python3
"""Generate frontend validation metadata from voxie proto files."""

from __future__ import annotations

import json
import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
PROTO_ROOT = ROOT / "proto"
OUT_DIR = ROOT / "gen" / "validation"


FIELD_RE = re.compile(
    r"^\s*(?P<label>optional|repeated)?\s*"
    r"(?P<type>[A-Za-z_][A-Za-z0-9_.]*)\s+"
    r"(?P<name>[a-z][a-z0-9_]*)\s*=\s*"
    r"(?P<number>\d+)"
    r"(?:\s*\[(?P<options>.*?)\])?\s*;",
    re.DOTALL,
)
MESSAGE_RE = re.compile(r"^\s*message\s+([A-Za-z][A-Za-z0-9_]*)\s*\{")
PACKAGE_RE = re.compile(r"^\s*package\s+([A-Za-z0-9_.]+)\s*;")


SCALAR_TYPES = {
    "double",
    "float",
    "int32",
    "int64",
    "uint32",
    "uint64",
    "sint32",
    "sint64",
    "fixed32",
    "fixed64",
    "sfixed32",
    "sfixed64",
    "bool",
    "string",
    "bytes",
}


def camel(name: str) -> str:
    parts = name.split("_")
    return parts[0] + "".join(p[:1].upper() + p[1:] for p in parts[1:])


def read_package(proto: Path) -> str:
    for line in proto.read_text().splitlines():
        match = PACKAGE_RE.match(line)
        if match:
            return match.group(1)
    raise RuntimeError(f"missing package in {proto}")


def extract_validation_rules(options: str | None, field_type: str) -> dict[str, object]:
    rules: dict[str, object] = {}
    if not options:
        return rules

    if "(buf.validate.field).required = true" in options:
        rules["required"] = True

    scalar = field_type.split(".")[-1]
    for proto_rule, meta_key in [
        ("min_len", "minLen"),
        ("max_len", "maxLen"),
        ("min_items", "minItems"),
        ("max_items", "maxItems"),
        ("gt", "gt"),
        ("gte", "gte"),
        ("lt", "lt"),
        ("lte", "lte"),
        ("email", "email"),
        ("uuid", "uuid"),
        ("uri", "uri"),
    ]:
        pattern = re.compile(
            rf"\(buf\.validate\.field\)\.(?:{re.escape(scalar)}|repeated)"
            rf"(?:\s*=\s*\{{[^}}]*\b{proto_rule}\s*:\s*([^,\s}}]+)|"
            rf"\.{proto_rule}\s*=\s*([^,\]\s]+))",
            re.DOTALL,
        )
        match = pattern.search(options)
        if not match:
            continue
        raw = next(group for group in match.groups() if group is not None).strip()
        if raw in {"true", "false"}:
            value: object = raw == "true"
        else:
            try:
                value = float(raw) if "." in raw else int(raw)
            except ValueError:
                value = raw.strip('"')
        rules[meta_key] = value

    return rules


def is_field_required(rules: dict[str, object]) -> bool:
    if rules.get("required") is True:
        return True
    min_len = rules.get("minLen")
    if isinstance(min_len, int) and min_len > 0:
        return True
    min_items = rules.get("minItems")
    if isinstance(min_items, int) and min_items > 0:
        return True
    if "gt" in rules:
        return True
    return False


def logical_proto_lines(proto: Path) -> list[str]:
    lines: list[str] = []
    pending: list[str] = []
    for raw_line in proto.read_text().splitlines():
        stripped = raw_line.strip()
        if pending:
            pending.append(raw_line)
            if stripped.endswith(";"):
                lines.append("\n".join(pending))
                pending = []
            continue
        if re.match(r"^\s*(optional|repeated)?\s*[A-Za-z_][A-Za-z0-9_.]*\s+[a-z][a-z0-9_]*\s*=", raw_line):
            if stripped.endswith(";"):
                lines.append(raw_line)
            else:
                pending = [raw_line]
            continue
        lines.append(raw_line)
    if pending:
        lines.append("\n".join(pending))
    return lines


def iter_message_fields(proto: Path) -> tuple[str, dict[str, dict[str, object]]]:
    package = read_package(proto)
    messages: dict[str, dict[str, object]] = {}
    stack: list[tuple[str, int]] = []

    for raw_line in logical_proto_lines(proto):
        line = raw_line.strip()
        msg_match = MESSAGE_RE.match(raw_line)
        if msg_match:
            parent = ".".join(name for name, _ in stack)
            name = msg_match.group(1)
            fq_name = f"{package}.{parent + '.' if parent else ''}{name}"
            stack.append((name, 1))
            messages[fq_name] = {}
            continue

        if not stack:
            continue

        field_match = FIELD_RE.match(raw_line)
        if field_match:
            label = field_match.group("label") or ""
            field_type = field_match.group("type")
            proto_name = field_match.group("name")
            is_repeated = label == "repeated"
            is_optional = label == "optional"
            is_scalar = field_type in SCALAR_TYPES
            rules = extract_validation_rules(field_match.group("options"), field_type)
            field = {
                "name": camel(proto_name),
                "protoName": proto_name,
                "type": field_type,
                "repeated": is_repeated,
                "optional": is_optional,
                "required": is_field_required(rules),
                "scalar": is_scalar,
            }
            field.update(rules)
            messages[f"{package}.{'.'.join(name for name, _ in stack)}"][camel(proto_name)] = field

        open_count = raw_line.count("{")
        close_count = raw_line.count("}")
        if stack:
            name, depth = stack[-1]
            depth += open_count - close_count
            if depth <= 0:
                stack.pop()
            else:
                stack[-1] = (name, depth)

    return package, messages


def main() -> None:
    all_messages: dict[str, dict[str, object]] = {}
    for proto in sorted(PROTO_ROOT.rglob("*.proto")):
        _, messages = iter_message_fields(proto)
        all_messages.update(messages)

    request_messages = {
        name: fields
        for name, fields in all_messages.items()
        if name.split(".")[-1].endswith("Request")
    }

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    json_payload = {
        "version": 1,
        "messages": request_messages,
    }
    (OUT_DIR / "metadata.json").write_text(json.dumps(json_payload, indent=2, sort_keys=True) + "\n")

    ts = (
        "/* eslint-disable */\n"
        "// Code generated by scripts/generate_validation_metadata.py. DO NOT EDIT.\n\n"
        "export type ContractFieldValidation = {\n"
        "  name: string;\n"
        "  protoName: string;\n"
        "  type: string;\n"
        "  repeated: boolean;\n"
        "  optional: boolean;\n"
        "  required: boolean;\n"
        "  scalar: boolean;\n"
        "  minLen?: number;\n"
        "  maxLen?: number;\n"
        "  minItems?: number;\n"
        "  maxItems?: number;\n"
        "  gt?: number;\n"
        "  gte?: number;\n"
        "  lt?: number;\n"
        "  lte?: number;\n"
        "  email?: boolean;\n"
        "  uuid?: boolean;\n"
        "  uri?: boolean;\n"
        "};\n\n"
        "export type ContractValidationMetadata = Record<string, Record<string, ContractFieldValidation>>;\n\n"
        f"export const contractValidationMetadata = {json.dumps(request_messages, indent=2, sort_keys=True)} as const satisfies ContractValidationMetadata;\n\n"
        "export function getContractValidation(messageName: string) {\n"
        "  return contractValidationMetadata[messageName as keyof typeof contractValidationMetadata];\n"
        "}\n"
    )
    (OUT_DIR / "metadata.ts").write_text(ts)


if __name__ == "__main__":
    main()
