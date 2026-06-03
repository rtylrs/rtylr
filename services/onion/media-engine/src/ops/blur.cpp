#include "ops/blur.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage blur(vips::VImage img, double sigma) {
    if (sigma <= 0) {
        throw ProcessingError("blur sigma must be greater than 0");
    }
    return img.gaussblur(sigma);
}

} // namespace ops
} // namespace voxire
