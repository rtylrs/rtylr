"""Convert a downloaded CUA-S1-FORMS .pt checkpoint into the safetensors +
config.json format the classifier package loads.

Weights are not committed. Point CUA_S1_REF at a trycua/cua checkout and run:

    pip install torch safetensors
    export CUA_S1_REF=/path/to/cua
    python convert_checkpoint.py cua-s1-forms.pt /path/to/output-dir

The .pt wrapper may carry config metadata; only tensor state and the known
config fields are used, and torch.load runs with weights_only=True so the
artifact cannot execute code.
"""

import argparse
import os
import sys

REF = os.environ.get("CUA_S1_REF")
if not REF:
    raise SystemExit("set CUA_S1_REF to a trycua/cua checkout")
sys.path.insert(0, os.path.join(REF, "libs", "cua-s1", "python", "src"))

import torch

from cua_s1.checkpoint import save_checkpoint_files

KNOWN_CONFIG_KEYS = (
    "encoder",
    "width",
    "rank",
    "layers",
    "heads",
    "context_tokens",
    "option_tokens",
)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("input", help="downloaded .pt checkpoint")
    parser.add_argument("output", help="output directory for model.safetensors + config.json")
    args = parser.parse_args()

    payload = torch.load(args.input, map_location="cpu", weights_only=True)
    state = payload["state_dict"] if isinstance(payload, dict) and "state_dict" in payload else payload
    raw_config = payload.get("config", {}) if isinstance(payload, dict) else {}
    config = {key: raw_config[key] for key in KNOWN_CONFIG_KEYS if key in raw_config}
    missing = [key for key in KNOWN_CONFIG_KEYS if key not in config]
    if missing:
        raise SystemExit(f"checkpoint config is missing {missing}")

    save_checkpoint_files(args.output, state, config, {
        "source": "cua-ai/cua-s1-forms (Hugging Face)",
        "license": "MIT (model card)",
    })
    params = sum(tensor.numel() for tensor in state.values())
    print(f"wrote {args.output} ({params} params)")


if __name__ == "__main__":
    main()
