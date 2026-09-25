#include "ops/vignette.hpp"
#include "ops/band_util.hpp"
#include "numeric.hpp"
#include "common.hpp"
#include <string>

namespace voxire {
namespace ops {

vips::VImage vignette(vips::VImage img, double strength) {
    strength = clampDouble(strength, 0.0, 1.0);
    if (strength == 0.0) {
        return img;
    }
    const int w = img.width();
    const int h = img.height();
    const int edge = clampInt(static_cast<int>(255.0 * (1.0 - strength)), 0, 255);

    // A radial gradient: white (full brightness) at the center fading to the
    // darkened edge value. Rendered via the SVG loader.
    const std::string svg =
        "<svg xmlns='http://www.w3.org/2000/svg' width='" + std::to_string(w) +
        "' height='" + std::to_string(h) + "'>"
        "<defs><radialGradient id='v' cx='50%' cy='50%' r='75%'>"
        "<stop offset='0%' stop-color='rgb(255,255,255)'/>"
        "<stop offset='100%' stop-color='rgb(" + std::to_string(edge) + "," +
        std::to_string(edge) + "," + std::to_string(edge) + ")'/>"
        "</radialGradient></defs>"
        "<rect width='" + std::to_string(w) + "' height='" + std::to_string(h) +
        "' fill='url(#v)'/></svg>";

    vips::VImage grad = vips::VImage::new_from_buffer(svg.c_str(), svg.size(), "");
    // Normalise the gradient to a 0..1 multiplier and replicate to 3 bands.
    vips::VImage m = grad.extract_band(0).linear(1.0 / 255.0, 0.0);
    vips::VImage m3 = m.bandjoin(m).bandjoin(m);

    Split s = splitAlpha(img);
    vips::VImage out = s.color * m3;
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
