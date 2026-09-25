#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// trim auto-crops uniform border pixels (e.g. whitespace) around the content,
// like the Cloudinary e_trim effect. `threshold` is how different a pixel must
// be from the detected border color to count as content (0-255). When nothing
// is trimmable the image is returned unchanged.
vips::VImage trim(vips::VImage img, double threshold);

} // namespace ops
} // namespace voxire
