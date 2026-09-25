#pragma once

#include <string>
#include <vips/vips8>

namespace voxire {
namespace transform {

// applyGeometry resizes/crops img per Cloudinary's w/h/c/g semantics:
//   scale  - resize to exactly wxh (may distort)
//   fit    - fit within wxh, keep aspect, no crop (may enlarge)
//   limit  - like fit but never enlarge
//   fill   - cover wxh and crop overflow at the gravity
//   thumb  - like fill but always attention-cropped
//   pad    - fit within wxh then pad to the exact box
// crop defaults to "fit" when empty. gravity "auto" selects attention-based
// cropping for fill/thumb.
vips::VImage applyGeometry(vips::VImage img, int w, int h,
                           const std::string& crop, const std::string& gravity);

} // namespace transform
} // namespace voxire
