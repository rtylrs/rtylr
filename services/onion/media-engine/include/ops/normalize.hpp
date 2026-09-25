#pragma once

#include <vips/vips8>

namespace voxire {
namespace ops {

// normalize stretches the tonal range so the darkest pixels map to 0 and the
// brightest to 255, improving contrast on flat/low-contrast images (Cloudinary
// e_improve / auto-contrast). A flat image is returned unchanged. Alpha is
// preserved.
vips::VImage normalize(vips::VImage img);

} // namespace ops
} // namespace voxire
