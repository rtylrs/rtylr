#pragma once

#include "../color.hpp"
#include <vips/vips8>

namespace voxire {
namespace ops {

// duotone maps image luminance onto a two-color gradient: dark pixels take the
// `shadow` color, bright pixels the `highlight` color, midtones blend between
// them. A classic editorial/Spotify-style effect. Alpha is preserved.
vips::VImage duotone(vips::VImage img, const Color& shadow, const Color& highlight);

} // namespace ops
} // namespace voxire
