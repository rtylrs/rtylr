#include "transform/apply.hpp"
#include "transform/params.hpp"
#include "transform/effects.hpp"
#include "transform/geometry_step.hpp"
#include "common.hpp"
#include "numeric.hpp"
#include "color.hpp"

#include "ops/rotate.hpp"
#include "ops/opacity.hpp"
#include "ops/rounded_corners.hpp"
#include "ops/flatten.hpp"

#include <map>
#include <string>

namespace voxire {
namespace transform {

static std::string get(const std::map<std::string, std::string>& p,
                       const std::string& key) {
    auto it = p.find(key);
    return it == p.end() ? std::string() : it->second;
}

static vips::VImage applyStep(vips::VImage img,
                              const std::map<std::string, std::string>& p) {
    // Geometry first (w/h/c/g) so later effects act on the final pixels.
    const std::string ws = get(p, "w");
    const std::string hs = get(p, "h");
    if (!ws.empty() || !hs.empty() || !get(p, "c").empty()) {
        const int w = ws.empty() ? 0 : parseIntStrict(ws, "w");
        const int h = hs.empty() ? 0 : parseIntStrict(hs, "h");
        img = applyGeometry(img, w, h, get(p, "c"), get(p, "g"));
    }

    // Rotation: a_ (angle).
    const std::string a = get(p, "a");
    if (!a.empty()) {
        img = ops::rotate(img, parseDoubleStrict(a, "a"), parseColor("transparent"));
    }

    // Background flatten: b_<color>.
    const std::string b = get(p, "b");
    if (!b.empty()) {
        img = ops::flatten(img, parseColor(b));
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

    return img;
}

vips::VImage applyChain(vips::VImage img, const std::string& chain) {
    for (const auto& step : splitSteps(chain)) {
        img = applyStep(img, parseParams(step));
    }
    return img;
}

} // namespace transform
} // namespace voxire
