#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// blur applies a Gaussian blur with the given sigma (standard deviation, in
// pixels). Larger sigma means a stronger blur. sigma must be > 0.
vips::VImage blur(vips::VImage img, double sigma);

} // namespace ops
} // namespace voxire
