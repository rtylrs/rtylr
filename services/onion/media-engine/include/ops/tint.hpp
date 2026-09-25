#pragma once

#include "../color.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// tint maps the image's luminance onto a single target color (a duotone toward
// `color`), then blends that with the original by `amount` (0 = unchanged,
// 1 = fully tinted). amount must be in 0..1. Alpha is preserved.
vips::VImage tint(vips::VImage img, const Color& color, double amount);

} // namespace ops
} // namespace voxire
