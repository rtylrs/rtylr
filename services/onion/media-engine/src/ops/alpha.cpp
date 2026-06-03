#include "ops/alpha.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage addAlpha(vips::VImage img, int value) {
    if (value < 0 || value > 255) {
        throw ProcessingError("alpha value must be between 0 and 255");
    }
    if (img.has_alpha()) {
        return img;
    }
    return img.bandjoin(value);
}

vips::VImage removeAlpha(vips::VImage img) {
    if (!img.has_alpha()) {
        return img;
    }
    return splitAlpha(img).color;
}

} // namespace ops
} // namespace voxire
