"""Regenerate the demo fixtures from the cua-s1 reference implementation.

Requires a checkout of trycua/cua (for libs/cua-s1, MIT) plus torch and
safetensors in the environment:

    export CUA_S1_REF=/path/to/cua          # repo root containing libs/cua-s1
    pip install torch safetensors
    python testdata/gen_fixtures.py internal/ai/classifier/testdata/forms

Default mode writes a real cua-s1 format checkpoint (model.safetensors +
config.json, with format markers and state signatures) plus golden.json
with reference logits for the Go parity test. The weights are a random demo
initialization (torch.manual_seed(7)) at reduced dims so the fixture stays
small; the Go code path is identical at full width.

With --checkpoint DIR the script scores the demo examples using that
checkpoint instead (for example a converted CUA-S1-FORMS release) and
writes golden.json into DIR, which the env-gated real-weights parity test
consumes:

    python testdata/gen_fixtures.py ignored --checkpoint /path/to/real-forms
    REAL_CHECKPOINT=/path/to/real-forms go test ./internal/ai/classifier/
"""

import argparse
import json
import os
import sys

REF = os.environ.get("CUA_S1_REF")
if not REF:
    raise SystemExit("set CUA_S1_REF to a trycua/cua checkout")
sys.path.insert(0, os.path.join(REF, "libs", "cua-s1", "python", "src"))

import torch

from cua_s1.checkpoint import save_checkpoint_files
from cua_s1.model import ChoiceExample, load_checkpoint, make_system, trainable_state

CONFIG = {
    "encoder": "tinyx",
    "width": 16,
    "rank": 4,
    "context_tokens": 32,
    "option_tokens": 16,
    "layers": 2,
    "heads": 2,
    "dropout": 0.0,
}

EXAMPLES = [
    ("TASK fill the form from the document, then submit\n"
     "FORM Northwind Clinic - New Patient Registration\n"
     "ELEMENT Edit \"Phone number\" value=\"\"",
     ["fill Tel: (503) 555-0142", "fill DOB: 03/14/1987", "check", "click", "skip"],
     0),
    ("TASK fill the form from the document, then submit\n"
     "FORM عيادة الشفاء - تسجيل مريض جديد\n"
     'ELEMENT Edit "الاسم الكامل" value=""',
     ["fill الاسم: محمد أحمد", "fill 名称：王小明", "fill Name: José García 🧑‍⚕️", "skip"],
     0),
    ("TASK fill the form from the document, then submit\n"
     "FORM " + "Very Long Department Name " * 10 + "\n"
     'ELEMENT Edit "Emergency contact phone" value=""',
     ["fill Phone: +1 633 208 1384", "fill Emergency: +1 633 208 9999", "skip"],
     1),
    ("TASK fill the form from the document, then submit\n"
     "FORM Demo\n"
     'ELEMENT Button "Submit Form" value=""',
     ["click", "skip"],
     0),
    ("TASK fill the form from the document, then submit\n"
     "FORM Demo\n"
     'ELEMENT Edit "Legal name" value="Sofia Larsson"',
     ["fill Legal name: Sofia Larsson", "fill Legal name: Liam Fischer", "skip"],
     2),
    ("TASK fill the form from the document, then submit\n"
     "FORM Demo\n"
     'ELEMENT Checkbox "Has insurance" value="unchecked"',
     ["fill Insurance: yes", "check", "skip"],
     1),
]


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("out", help="fixture directory (default mode)")
    parser.add_argument("--checkpoint", help="converted checkpoint dir to score with")
    args = parser.parse_args()

    if args.checkpoint:
        model, collator, config = load_checkpoint(args.checkpoint, "cpu")
        out = args.checkpoint
    else:
        torch.manual_seed(7)
        model, collator = make_system(CONFIG, "cpu")
        config = CONFIG
        out = args.out
    model.eval()

    golden_examples = []
    with torch.no_grad():
        for context, options, label in EXAMPLES:
            ex = ChoiceExample(context=context, options=tuple(options), label=label)
            batch = collator([ex])
            result = model(batch)
            logits = (result["logits"] if isinstance(result, dict) else result)[0].tolist()
            pred = max(range(len(logits)), key=logits.__getitem__)
            golden_examples.append(
                {"context": context, "options": options, "label": label,
                 "logits": logits, "reference_argmax": pred}
            )

    if not args.checkpoint:
        save_checkpoint_files(out, trainable_state(model), config, {"demo": "form-mapping-go"})
    with open(f"{out}/golden.json", "w", encoding="utf-8") as fh:
        json.dump({"config": config, "examples": golden_examples}, fh, indent=2,
                  sort_keys=True, ensure_ascii=False)
        fh.write("\n")

    print("params:", sum(p.numel() for p in model.parameters()))
    print("examples:", len(golden_examples))


if __name__ == "__main__":
    main()
