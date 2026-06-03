#include "ops/border.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage border(vips::VImage img, int width, const Color& color) {
    if (width <= 0) {
        throw ProcessingError("border width must be greater than 0");
    }
    // Ensure an alpha band exists when the border color is translucent, so the
    // border and image composite consistently.
    if (color.a < 255 && !img.has_alpha()) {
        img = img.bandjoin(255);
    }
    const int w = img.width() + 2 * width;
    const int h = img.height() + 2 * width;
    return img.embed(width, width, w, h,
        vips::VImage::option()
            ->set("extend", VIPS_EXTEND_BACKGROUND)
            ->set("background", color.bands()));
}

} // namespace ops
} // namespace voxire
