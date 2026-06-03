#include "ops/composite.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage composite(vips::VImage base, vips::VImage overlay,
                       Gravity gravity, int margin) {
    if (margin < 0) {
        throw ProcessingError("composite margin must not be negative");
    }
    // Anchor the overlay within the base inset by `margin` on every edge.
    const int innerW = base.width() - 2 * margin;
    const int innerH = base.height() - 2 * margin;
    int x = 0, y = 0;
    offsetFor(gravity, overlay.width(), overlay.height(), innerW, innerH, x, y);
    x += margin;
    y += margin;

    return base.composite2(overlay, VIPS_BLEND_MODE_OVER,
        vips::VImage::option()->set("x", x)->set("y", y));
}

} // namespace ops
} // namespace voxire
