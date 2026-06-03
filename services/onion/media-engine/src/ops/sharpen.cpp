#include "ops/sharpen.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage sharpen(vips::VImage img, double sigma) {
    if (sigma <= 0) {
        throw ProcessingError("sharpen sigma must be greater than 0");
    }
    return img.sharpen(vips::VImage::option()->set("sigma", sigma));
}

} // namespace ops
} // namespace voxire
