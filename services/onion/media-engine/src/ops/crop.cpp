#include "ops/crop.hpp"
#include "common.hpp"
#include <algorithm>

namespace voxire {
namespace ops {

vips::VImage crop(vips::VImage img, const Box& r) {
    if (r.empty()) {
        throw ProcessingError("crop rectangle must have positive width and height");
    }
    if (r.x < 0 || r.y < 0) {
        throw ProcessingError("crop origin must not be negative");
    }
    if (r.x >= img.width() || r.y >= img.height()) {
        throw ProcessingError("crop origin is outside the image bounds");
    }
    // Clamp the extent so a slightly-too-large rectangle trims to the edge
    // instead of erroring.
    const int w = std::min(r.width, img.width() - r.x);
    const int h = std::min(r.height, img.height() - r.y);
    return img.crop(r.x, r.y, w, h);
}

} // namespace ops
} // namespace voxire
