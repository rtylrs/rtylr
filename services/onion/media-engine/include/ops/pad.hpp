#pragma once

#include "../color.hpp"
#include "../gravity.hpp"
#include "../geometry.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// pad places the image on a canvas of the target size, filling the surrounding
// space with `background` and anchoring the image by `gravity` (the Cloudinary
// c_pad mode). If the image is larger than the target in a dimension it is left
// as-is in that axis (no scaling — pair with a resize for c_lpad behavior).
vips::VImage pad(vips::VImage img, const Size& target, Gravity gravity,
                 const Color& background);

} // namespace ops
} // namespace voxire
