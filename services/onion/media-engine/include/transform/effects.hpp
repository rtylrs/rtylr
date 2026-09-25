#pragma once

#include <string>
#include <vips/vips8>

namespace voxire {
namespace transform {

// applyEffect dispatches a Cloudinary-style e_<name>[:value] effect to the
// matching op. `name` is the effect (blur, sharpen, grayscale, sepia, negate,
// pixelate, brightness, contrast, saturation, posterize, blackwhite,
// oil_paint), `value` is the optional argument string (may be empty for a
// sensible default). Throws ProcessingError on an unknown effect.
vips::VImage applyEffect(vips::VImage img, const std::string& name,
                         const std::string& value);

} // namespace transform
} // namespace voxire
