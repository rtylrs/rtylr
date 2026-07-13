#!/usr/bin/env python3
"""Generate frontend validation metadata from voxie proto files."""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path


DEFAULT_ROOT = Path(__file__).resolve().parents[1]


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
MESSAGE_ONEOF_RE = re.compile(
    r"\(buf\.validate\.message\)\.oneof\s*=\s*\{(?P<body>.*?)\}",
    re.DOTALL,
)


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


def extract_message_oneof_options(line: str) -> list[dict[str, object]]:
    oneofs: list[dict[str, object]] = []
    for match in MESSAGE_ONEOF_RE.finditer(line):
        body = match.group("body")
        proto_fields = re.findall(r'"([a-z][a-z0-9_]*)"', body)
        if not proto_fields:
            continue
        oneofs.append(
            {
                "fields": [camel(field) for field in proto_fields],
                "protoFields": proto_fields,
                "required": re.search(r"\brequired\s*:\s*true", body) is not None,
            }
        )
    return oneofs


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
        if stripped.startswith("option ") or re.match(
            r"^\s*(optional|repeated)?\s*[A-Za-z_][A-Za-z0-9_.]*\s+[a-z][a-z0-9_]*\s*=", raw_line
        ):
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
    message_oneofs: dict[str, list[dict[str, object]]] = {}
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
            message_oneofs[fq_name] = []
            continue

        if not stack:
            continue

        current_message = f"{package}.{'.'.join(name for name, _ in stack)}"
        if "(buf.validate.message).oneof" in raw_line:
            message_oneofs[current_message].extend(extract_message_oneof_options(raw_line))

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
            messages[current_message][camel(proto_name)] = field

        open_count = raw_line.count("{")
        close_count = raw_line.count("}")
        if stack:
            name, depth = stack[-1]
            depth += open_count - close_count
            if depth <= 0:
                stack.pop()
            else:
                stack[-1] = (name, depth)

    for name, oneofs in message_oneofs.items():
        if oneofs:
            messages.setdefault(name, {})["__oneofs"] = oneofs

    return package, messages


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate frontend validation metadata from voxie proto files.")
    # Every service has been migrated off .proto onto api/*.onk (see
    # README.md), so proto/ no longer exists and this always scans zero
    # files - the generated metadata is now permanently empty (but still
    # type-checks; see the comment on contractValidationMetadata's type
    # annotation below). This script has no .onk-aware equivalent yet: it
    # only understands proto's buf.validate syntax. Frontend field validation
    # metadata is a known gap until an onk-decorator-aware generator exists.
    parser.add_argument("--proto-root", type=Path, default=DEFAULT_ROOT / "api")
    parser.add_argument("--out-dir", type=Path, default=DEFAULT_ROOT / "src" / "validation")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    proto_root = args.proto_root
    out_dir = args.out_dir
    all_messages: dict[str, dict[str, object]] = {}
    for proto in sorted(proto_root.rglob("*.proto")):
        _, messages = iter_message_fields(proto)
        all_messages.update(messages)

    request_messages = {
        name: {field_name: field for field_name, field in fields.items() if field_name != "__oneofs"}
        for name, fields in all_messages.items()
        if name.split(".")[-1].endswith("Request")
    }
    request_oneofs = {
        name: fields["__oneofs"]
        for name, fields in all_messages.items()
        if name.split(".")[-1].endswith("Request") and "__oneofs" in fields
    }

    out_dir.mkdir(parents=True, exist_ok=True)
    json_payload = {
        "version": 1,
        "messages": request_messages,
        "oneofs": request_oneofs,
    }
    (out_dir / "metadata.json").write_text(json.dumps(json_payload, indent=2, sort_keys=True) + "\n")

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
        "export type ContractOneofValidation = {\n"
        "  fields: readonly string[];\n"
        "  protoFields: readonly string[];\n"
        "  required: boolean;\n"
        "};\n\n"
        "export type ContractOneofValidationMetadata = Record<string, readonly ContractOneofValidation[]>;\n\n"
        "export type ContractValidationIssue = {\n"
        "  field: string;\n"
        "  protoName: string;\n"
        "  message: string;\n"
        "};\n\n"
        # An explicit type annotation (rather than relying solely on `as const
        # satisfies` inference) is required here: when the underlying dict is
        # empty (e.g. every oneof-validated proto message has since been
        # migrated to .onk, which has no equivalent decorator), TypeScript
        # infers the bare object literal `{}` as its own zero-key type, so
        # `keyof typeof metadata` collapses to `never` and callers indexing
        # into it break. Annotating the declared type keeps `keyof` as
        # `string` regardless of how many keys are actually present.
        f"export const contractValidationMetadata: ContractValidationMetadata = {json.dumps(request_messages, indent=2, sort_keys=True)} as const;\n\n"
        f"export const contractValidationOneofMetadata: ContractOneofValidationMetadata = {json.dumps(request_oneofs, indent=2, sort_keys=True)} as const;\n\n"
        "export function getContractValidation(messageName: string) {\n"
        "  return contractValidationMetadata[messageName as keyof typeof contractValidationMetadata];\n"
        "}\n"
        "\n"
        "export function getContractOneofValidation(messageName: string) {\n"
        "  return contractValidationOneofMetadata[messageName as keyof typeof contractValidationOneofMetadata] ?? [];\n"
        "}\n"
        "\n"
        "export function getRequiredContractFields(messageName: string) {\n"
        "  const fields = getContractValidation(messageName);\n"
        "  if (!fields) return [];\n"
        "  return Object.values(fields).filter((field) => field.required);\n"
        "}\n"
        "\n"
        "function hasValue(value: unknown, field: ContractFieldValidation): boolean {\n"
        "  if (value === null || value === undefined) return false;\n"
        "  if (typeof value === \"string\") return value.trim().length > 0;\n"
        "  if (field.repeated) return Array.isArray(value) && value.length > 0;\n"
        "  return true;\n"
        "}\n"
        "\n"
        "function numberValue(value: unknown): number | null {\n"
        "  if (typeof value === \"number\") return Number.isFinite(value) ? value : null;\n"
        "  if (typeof value === \"string\" && value.trim()) {\n"
        "    const parsed = Number(value);\n"
        "    return Number.isFinite(parsed) ? parsed : null;\n"
        "  }\n"
        "  return null;\n"
        "}\n"
        "\n"
        "export function validateContractPayload(messageName: string, payload: Record<string, unknown>) {\n"
        "  const fields = getContractValidation(messageName);\n"
        "  const oneofs = getContractOneofValidation(messageName);\n"
        "  const issues: ContractValidationIssue[] = [];\n"
        "  if (!fields) return issues;\n"
        "  const fieldMap = fields as Record<string, ContractFieldValidation>;\n"
        "\n"
        "  for (const oneof of oneofs) {\n"
        "    const present = oneof.fields.filter((fieldName) => hasValue(payload[fieldName], fieldMap[fieldName] ?? { repeated: false } as ContractFieldValidation));\n"
        "    if (oneof.required && present.length === 0) {\n"
        "      issues.push({ field: oneof.fields[0] ?? \"\", protoName: oneof.protoFields[0] ?? \"\", message: `one of ${oneof.fields.join(\", \")} is required` });\n"
        "    }\n"
        "    if (present.length > 1) {\n"
        "      issues.push({ field: present[0] ?? \"\", protoName: oneof.protoFields[0] ?? \"\", message: `only one of ${oneof.fields.join(\", \")} can be set` });\n"
        "    }\n"
        "  }\n"
        "\n"
        "  for (const field of Object.values(fields)) {\n"
        "    const value = payload[field.name];\n"
        "    if (field.required && !hasValue(value, field)) {\n"
        "      issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} is required` });\n"
        "      continue;\n"
        "    }\n"
        "    if (!hasValue(value, field)) continue;\n"
        "\n"
        "    if (typeof value === \"string\") {\n"
        "      if (field.minLen !== undefined && value.trim().length < field.minLen) {\n"
        "        issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} must be at least ${field.minLen} characters` });\n"
        "      }\n"
        "      if (field.maxLen !== undefined && value.length > field.maxLen) {\n"
        "        issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} must be at most ${field.maxLen} characters` });\n"
        "      }\n"
        "      if (field.email && !/^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$/.test(value.trim())) {\n"
        "        issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} must be a valid email` });\n"
        "      }\n"
        "    }\n"
        "\n"
        "    if (field.repeated && Array.isArray(value)) {\n"
        "      if (field.minItems !== undefined && value.length < field.minItems) {\n"
        "        issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} must include at least ${field.minItems} item(s)` });\n"
        "      }\n"
        "      if (field.maxItems !== undefined && value.length > field.maxItems) {\n"
        "        issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} must include at most ${field.maxItems} item(s)` });\n"
        "      }\n"
        "    }\n"
        "\n"
        "    const numeric = numberValue(value);\n"
        "    if (numeric !== null) {\n"
        "      if (field.gt !== undefined && numeric <= field.gt) {\n"
        "        issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} must be greater than ${field.gt}` });\n"
        "      }\n"
        "      if (field.gte !== undefined && numeric < field.gte) {\n"
        "        issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} must be at least ${field.gte}` });\n"
        "      }\n"
        "      if (field.lt !== undefined && numeric >= field.lt) {\n"
        "        issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} must be less than ${field.lt}` });\n"
        "      }\n"
        "      if (field.lte !== undefined && numeric > field.lte) {\n"
        "        issues.push({ field: field.name, protoName: field.protoName, message: `${field.name} must be at most ${field.lte}` });\n"
        "      }\n"
        "    }\n"
        "  }\n"
        "\n"
        "  return issues;\n"
        "}\n"
    )
    (out_dir / "metadata.ts").write_text(ts)


if __name__ == "__main__":
    main()
