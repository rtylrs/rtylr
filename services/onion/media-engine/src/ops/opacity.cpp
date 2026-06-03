#include "ops/opacity.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage opacity(vips::VImage img, double factor) {
    if (factor < 0.0 || factor > 1.0) {
        throw ProcessingError("opacity factor must be between 0 and 1");
    }
    if (!img.has_alpha()) {
        img = img.bandjoin(255);
    }
    Split s = splitAlpha(img);
    vips::VImage scaledAlpha = s.alpha.linear(factor, 0.0);
    return s.color.bandjoin(scaledAlpha);
}

} // namespace ops
} // namespace voxire
