#include "transform/apply.hpp"
#include "transform/params.hpp"
#include "transform/effects.hpp"
#include "transform/geometry_step.hpp"
#include "common.hpp"
#include "numeric.hpp"
#include "color.hpp"

#include "ops/border.hpp"
#include "ops/crop.hpp"
#include "ops/flip.hpp"
#include "ops/flop.hpp"
#include "ops/rotate.hpp"
#include "ops/opacity.hpp"
#include "ops/rounded_corners.hpp"
#include "ops/flatten.hpp"
#include "ops/scale.hpp"
#include "ops/composite.hpp"
#include "ops/text.hpp"

#include <algorithm>
#include <cstdint>
#include <map>
#include <string>

namespace voxire {
namespace transform {

static std::string get(const std::map<std::string, std::string>& p,
                       const std::string& key) {
    auto it = p.find(key);
    return it == p.end() ? std::string() : it->second;
}

static int getIntOr(const std::map<std::string, std::string>& p,
                    const std::string& key, int fallback) {
    const std::string value = get(p, key);
    return value.empty() ? fallback : parseIntStrict(value, key);
}

static vips::VImage applyCrop(vips::VImage img,
                              const std::map<std::string, std::string>& p,
                              int w, int h) {
    if (w <= 0 || h <= 0) {
        throw ProcessingError("crop requires positive w and h");
    }
    int x = getIntOr(p, "x", -1);
    int y = getIntOr(p, "y", -1);
    if (x < 0) x = std::max(0, (img.width() - w) / 2);
    if (y < 0) y = std::max(0, (img.height() - h) / 2);
    return ops::crop(img, Box{x, y, w, h});
}

static void parseBorder(const std::string& raw, int& width, Color& color) {
    std::string w = raw;
    std::string c = "black";
    const auto sep = raw.find(':');
    if (sep != std::string::npos) {
        w = raw.substr(0, sep);
        c = raw.substr(sep + 1);
    }
    width = parseIntStrict(w, "bo");
    color = parseColor(c.empty() ? "black" : c);
}

static std::string underscoresToSpaces(std::string s) {
    std::replace(s.begin(), s.end(), '_', ' ');
    return s;
}

static vips::VImage applyStep(vips::VImage img,
                              const std::map<std::string, std::string>& p) {
    // Scale/DPR can run before explicit geometry.
    const std::string z = get(p, "z");
    if (!z.empty()) {
        img = ops::scaleBy(img, parseDoubleStrict(z, "z"));
    }

    // Geometry first (w/h/c/g) so later effects act on the final pixels.
    const std::string ws = get(p, "w");
    const std::string hs = get(p, "h");
    if (!ws.empty() || !hs.empty() || !get(p, "c").empty()) {
        const int w = ws.empty() ? 0 : parseIntStrict(ws, "w");
        const int h = hs.empty() ? 0 : parseIntStrict(hs, "h");
        const std::string crop = get(p, "c");
        if (crop == "crop" || crop == "extract") {
            img = applyCrop(img, p, w, h);
        } else {
            img = applyGeometry(img, w, h, crop, get(p, "g"));
        }
    }

    // Rotation: a_ (angle).
    const std::string a = get(p, "a");
    if (!a.empty()) {
        const std::string bg = get(p, "b");
        img = ops::rotate(img, parseDoubleStrict(a, "a"),
                          bg.empty() ? parseColor("transparent") : parseColor(bg));
    }

    // Background flatten: b_<color>.
    const std::string b = get(p, "b");
    if (!b.empty()) {
        img = ops::flatten(img, parseColor(b));
    }

    // Border: bo_<width>[:<color>] grows the canvas with a solid border.
    const std::string bo = get(p, "bo");
    if (!bo.empty()) {
        int width = 0;
        Color color;
        parseBorder(bo, width, color);
        img = ops::border(img, width, color);
    }

    // Effect: e_<name>[:value].
    const std::string e = get(p, "e");
    if (!e.empty()) {
        const auto colon = e.find(':');
        const std::string name = colon == std::string::npos ? e : e.substr(0, colon);
        const std::string val = colon == std::string::npos ? "" : e.substr(colon + 1);
        img = applyEffect(img, name, val);
    }

    // Rounded corners: r_<radius> or r_max (circle/ellipse).
    const std::string r = get(p, "r");
    if (!r.empty()) {
        int radius;
        if (r == "max") {
            radius = (img.width() < img.height() ? img.width() : img.height()) / 2;
        } else {
            radius = parseIntStrict(r, "r");
        }
        if (radius > 0) img = ops::roundedCorners(img, radius);
    }

    // Opacity: o_<0..100>.
    const std::string o = get(p, "o");
    if (!o.empty()) {
        img = ops::opacity(img, clampDouble(parseDoubleStrict(o, "o") / 100.0, 0.0, 1.0));
    }

    // Text overlay. Keep this intentionally simple and deterministic:
    // txt_<text>,font_<pango-font>,tc_<color>,tw_<max-width>,tg_<gravity>,tm_<margin>.
    const std::string txt = get(p, "txt");
    if (!txt.empty()) {
        const std::string font = get(p, "font").empty()
            ? "sans 48"
            : underscoresToSpaces(get(p, "font"));
        const Color color = get(p, "tc").empty() ? parseColor("white") : parseColor(get(p, "tc"));
        const int maxWidth = getIntOr(p, "tw", 0);
        std::string gravityName = get(p, "tg").empty() ? get(p, "g") : get(p, "tg");
        if (gravityName.empty() || gravityName == "auto") gravityName = "center";
        const Gravity gravity = parseGravity(gravityName);
        const int margin = getIntOr(p, "tm", 24);
        vips::VImage overlay = ops::renderText(underscoresToSpaces(txt), font, maxWidth, color);
        const std::string textOpacity = get(p, "to");
        if (!textOpacity.empty()) {
            overlay = ops::opacity(overlay, clampDouble(parseDoubleStrict(textOpacity, "to") / 100.0, 0.0, 1.0));
        }
        img = ops::composite(img, overlay, gravity, margin);
    }

    // Flags: fl_flip, fl_flop, or fl_flip.flop for explicit mirroring.
    const std::string fl = get(p, "fl");
    if (!fl.empty()) {
        if (fl.find("flip") != std::string::npos) img = ops::flipVertical(img);
        if (fl.find("flop") != std::string::npos) img = ops::flopHorizontal(img);
    }

    const std::string fx = get(p, "fx");
    if (fx == "flip") {
        img = ops::flipVertical(img);
    } else if (fx == "flop") {
        img = ops::flopHorizontal(img);
    } else if (!fx.empty()) {
        throw ProcessingError("unknown transform fx: " + fx);
    }

    return img;
}

vips::VImage applyChain(vips::VImage img, const std::string& chain) {
    for (const auto& step : splitSteps(chain)) {
        img = applyStep(img, parseParams(step));
    }
    return img;
}

std::optional<vips::VImage> tryApplyChainFromFile(const std::string& path,
                                                  const std::string& chain) {
    const auto steps = splitSteps(chain);
    if (steps.empty()) return std::nullopt;

    // Only the first step can be folded into the decode. If it carries any
    // directive beyond plain geometry we fall back to the generic pipeline.
    const auto p = parseParams(steps[0]);
    for (const auto& [key, val] : p) {
        if (key != "w" && key != "h" && key != "c" && key != "g") {
            return std::nullopt;
        }
    }

    const auto strW = p.find("w");
    const auto strH = p.find("h");
    const int w = strW != p.end() && !strW->second.empty()
        ? parseIntStrict(strW->second, "w") : 0;
    const int h = strH != p.end() && !strH->second.empty()
        ? parseIntStrict(strH->second, "h") : 0;
    if (w <= 0 && h <= 0) return std::nullopt;
	if (w > 8000 || h > 8000 || (w > 0 && h > 0 && static_cast<int64_t>(w) * h > 24'000'000)) {
		throw ProcessingError("requested geometry exceeds safety limits");
	}

    const std::string crop = p.count("c") ? p.at("c") : "";
    const std::string gravity = p.count("g") ? p.at("g") : "";

    // Mirror applyGeometry's strategies exactly; anything else (extract with
    // offsets, pad, fit modes needing post-steps) falls back.
    //
    // vips_thumbnail crops to the target aspect whenever both dimensions are
    // given, so aspect-preserving fits ("", "limit") are only safe when the
    // caller supplied a single dimension.
    VipsSize size = VIPS_SIZE_BOTH;
    VipsInteresting strat = VIPS_INTERESTING_CENTRE;
    bool singleDim = false;
    if ((w > 0) != (h > 0)) {
        // vips_thumbnail takes the target width as its primary argument, so
        // height-only fits stay on the generic path.
        if (w <= 0) return std::nullopt;
        singleDim = true;
        size = (crop == "limit") ? VIPS_SIZE_DOWN : VIPS_SIZE_BOTH;
    } else if (crop.empty() || crop == "limit") {
        return std::nullopt; // both dims + aspect-preserving fit
    } else if (crop == "scale") {
        size = VIPS_SIZE_FORCE;
    } else if (crop == "fill" || crop == "thumb") {
        strat = (crop == "thumb" || gravity == "auto" || gravity.empty())
            ? VIPS_INTERESTING_ATTENTION : VIPS_INTERESTING_CENTRE;
    } else {
        return std::nullopt;
    }

    // File-level thumbnail lets loaders shrink on decode (JPEG decoders
    // reduce before decompression), which is dramatically cheaper than
    // decoding full-size pixels and resampling afterwards. It also honours
    // EXIF orientation by default, matching the autorot() in the slow path.
    vips::VImage img = singleDim
        ? vips::VImage::thumbnail(path.c_str(), w > 0 ? w : h,
              vips::VImage::option()->set("size", size))
        : vips::VImage::thumbnail(path.c_str(), w,
              vips::VImage::option()
                  ->set("height", h)
                  ->set("size", size)
                  ->set("crop", strat));

    for (size_t i = 1; i < steps.size(); ++i) {
        img = applyStep(img, parseParams(steps[i]));
    }
    return img;
}

} // namespace transform
} // namespace voxire
