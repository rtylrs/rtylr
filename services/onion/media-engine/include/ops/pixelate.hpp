#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// pixelate downsamples the image by blockSize and scales it back with
// nearest-neighbour, producing square blocks of uniform color (the Cloudinary
// e_pixelate effect). blockSize must be >= 2.
vips::VImage pixelate(vips::VImage img, int blockSize);

} // namespace ops
} // namespace voxire
