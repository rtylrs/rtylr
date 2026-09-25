#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// gamma applies gamma correction with the given exponent. exponent > 1
// brightens midtones, < 1 darkens them. exponent must be > 0.
vips::VImage gamma(vips::VImage img, double exponent);

} // namespace ops
} // namespace voxire
