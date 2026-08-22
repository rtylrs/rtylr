#include "transform/geometry_step.hpp"
#include "ops/smartcrop.hpp"
#include "ops/pad.hpp"
#include "geometry.hpp"
#include "gravity.hpp"
#include "color.hpp"

#include <cstdint>

namespace voxire {
namespace transform {

vips::VImage applyGeometry(vips::VImage img, int w, int h,
                           const std::string& crop, const std::string& gravity) {
	if (w <= 0 && h <= 0) {
		return img;
	}
	if (w > 8000 || h > 8000 || (w > 0 && h > 0 && static_cast<int64_t>(w) * h > 24'000'000)) {
		throw ProcessingError("requested geometry exceeds safety limits");
	}

    if (crop == "scale") {
        const int tw = w > 0 ? w : img.width();
        const int th = h > 0 ? h : img.height();
        return img.thumbnail_image(tw, vips::VImage::option()
            ->set("height", th)->set("size", VIPS_SIZE_FORCE));
    }

    if ((crop == "fill" || crop == "thumb") && w > 0 && h > 0) {
        VipsInteresting strat = (crop == "thumb" || gravity == "auto" || gravity.empty())
            ? VIPS_INTERESTING_ATTENTION : VIPS_INTERESTING_CENTRE;
        return ops::smartCrop(img, Size{w, h}, strat);
    }

    // fit / limit / pad / default: contain within the box keeping aspect.
    const int tw = w > 0 ? w : 100000;
    const int th = h > 0 ? h : 100000;
    const VipsSize sz = (crop == "limit") ? VIPS_SIZE_DOWN : VIPS_SIZE_BOTH;
    vips::VImage out = img.thumbnail_image(tw, vips::VImage::option()
        ->set("height", th)->set("size", sz));

    if (crop == "pad" && w > 0 && h > 0) {
        const Gravity g = gravity.empty() ? Gravity::Center : parseGravity(gravity);
        out = ops::pad(out, Size{w, h}, g, parseColor("transparent"));
    }
    return out;
}

} // namespace transform
} // namespace voxire
