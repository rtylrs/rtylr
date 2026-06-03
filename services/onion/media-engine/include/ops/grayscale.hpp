#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// grayscale converts the image to the sRGB grayscale (B_W) colorspace and back
// to sRGB so the output stays a normal 3-band (plus optional alpha) image that
// every encoder accepts.
vips::VImage grayscale(vips::VImage img);

} // namespace ops
} // namespace voxire
