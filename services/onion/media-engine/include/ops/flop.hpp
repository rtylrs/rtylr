#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// flopHorizontal mirrors the image left-to-right (about the vertical axis).
vips::VImage flopHorizontal(vips::VImage img);

} // namespace ops
} // namespace voxire
