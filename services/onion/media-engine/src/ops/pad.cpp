#include "ops/pad.hpp"
#include "common.hpp"
#include <algorithm>

namespace voxire {
namespace ops {

vips::VImage pad(vips::VImage img, const Size& target, Gravity gravity,
                 const Color& background) {
    if (target.width <= 0 || target.height <= 0) {
        throw ProcessingError("pad target size must be positive");
    }
    // The canvas is at least as large as the image in each axis.
    const int outW = std::max(target.width, img.width());
    const int outH = std::max(target.height, img.height());

    int x = 0, y = 0;
    offsetFor(gravity, img.width(), img.height(), outW, outH, x, y);

    if (background.a < 255 && !img.has_alpha()) {
        img = img.bandjoin(255);
    }
    return img.embed(x, y, outW, outH,
        vips::VImage::option()
            ->set("extend", VIPS_EXTEND_BACKGROUND)
            ->set("background", background.bands()));
}

} // namespace ops
} // namespace voxire
