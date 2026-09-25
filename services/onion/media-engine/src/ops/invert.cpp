#include "ops/invert.hpp"

namespace voxire {
namespace ops {

vips::VImage invert(vips::VImage img) {
    // Split off alpha (if present) so the negative doesn't invert transparency.
    if (img.has_alpha()) {
        const int n = img.bands();
        vips::VImage alpha = img.extract_band(n - 1);
        vips::VImage rgb = img.extract_band(0, vips::VImage::option()->set("n", n - 1));
        return rgb.invert().bandjoin(alpha);
    }
    return img.invert();
}

} // namespace ops
} // namespace voxire
