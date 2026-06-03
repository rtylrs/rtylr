#include "ops/smartcrop.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

VipsInteresting parseInteresting(const std::string& s) {
    if (s == "auto" || s == "attention") return VIPS_INTERESTING_ATTENTION;
    if (s == "entropy")                   return VIPS_INTERESTING_ENTROPY;
    if (s == "centre" || s == "center")   return VIPS_INTERESTING_CENTRE;
    if (s == "low")                       return VIPS_INTERESTING_LOW;
    if (s == "high")                      return VIPS_INTERESTING_HIGH;
    throw ProcessingError("unknown crop strategy: " + s);
}

vips::VImage smartCrop(vips::VImage img, const Size& target,
                       VipsInteresting strategy) {
    if (target.width <= 0 || target.height <= 0) {
        throw ProcessingError("smartcrop target size must be positive");
    }
    return img.thumbnail_image(target.width, vips::VImage::option()
        ->set("height", target.height)
        ->set("size", VIPS_SIZE_BOTH)
        ->set("crop", strategy));
}

} // namespace ops
} // namespace voxire
