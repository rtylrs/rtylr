#include "ops/gamma.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage gamma(vips::VImage img, double exponent) {
    if (exponent <= 0) {
        throw ProcessingError("gamma exponent must be greater than 0");
    }
    return img.gamma(vips::VImage::option()->set("exponent", exponent));
}

} // namespace ops
} // namespace voxire
