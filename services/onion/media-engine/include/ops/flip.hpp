#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// flipVertical mirrors the image top-to-bottom (about the horizontal axis).
vips::VImage flipVertical(vips::VImage img);

} // namespace ops
} // namespace voxire
