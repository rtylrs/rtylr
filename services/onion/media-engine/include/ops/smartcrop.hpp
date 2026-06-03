#pragma once

#include "../geometry.hpp"
#include <string>
#include <vips/vips8>

namespace voxire {
namespace ops {

// parseInteresting maps a strategy name to a libvips crop strategy:
//   auto/attention -> ATTENTION, entropy -> ENTROPY, centre/center -> CENTRE,
//   low -> LOW, high -> HIGH. Throws ProcessingError on anything else.
VipsInteresting parseInteresting(const std::string& s);

// smartCrop resizes to fill the target box and crops the most salient region,
// implementing Cloudinary's g_auto. `strategy` selects the saliency model.
vips::VImage smartCrop(vips::VImage img, const Size& target,
                       VipsInteresting strategy);

} // namespace ops
} // namespace voxire
