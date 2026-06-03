#include "ops/flatten.hpp"

namespace voxire {
namespace ops {

vips::VImage flatten(vips::VImage img, const Color& background) {
    if (!img.has_alpha()) {
        return img;
    }
    return img.flatten(vips::VImage::option()->set(
        "background", std::vector<double>{
            static_cast<double>(background.r),
            static_cast<double>(background.g),
            static_cast<double>(background.b),
        }));
}

} // namespace ops
} // namespace voxire
