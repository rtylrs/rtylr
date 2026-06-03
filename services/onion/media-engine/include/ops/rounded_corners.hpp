#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// roundedCorners masks the image with a rounded rectangle of the given corner
// radius (px), producing transparency outside the rounded shape (Cloudinary
// r_<radius>). A radius >= min(w,h)/2 yields an ellipse/circle. radius must be
// > 0.
vips::VImage roundedCorners(vips::VImage img, int radius);

} // namespace ops
} // namespace voxire
