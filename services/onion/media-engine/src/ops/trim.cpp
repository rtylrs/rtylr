#include "ops/trim.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage trim(vips::VImage img, double threshold) {
    if (threshold < 0) {
        throw ProcessingError("trim threshold must not be negative");
    }
    // find_trim returns the content bounding box (left via the return value,
    // the rest through out-params), using the top-left pixel as the border
    // reference color.
    int top = 0, w = 0, h = 0;
    const int left = img.find_trim(&top, &w, &h,
        vips::VImage::option()->set("threshold", threshold));

    if (w <= 0 || h <= 0) {
        return img; // nothing detected as trimmable
    }
    return img.crop(left, top, w, h);
}

} // namespace ops
} // namespace voxire
