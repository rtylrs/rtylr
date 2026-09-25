#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// sharpen applies an unsharp-mask sharpen. sigma controls the radius of the
// effect (in pixels); larger values sharpen coarser detail. sigma must be > 0.
vips::VImage sharpen(vips::VImage img, double sigma);

} // namespace ops
} // namespace voxire
