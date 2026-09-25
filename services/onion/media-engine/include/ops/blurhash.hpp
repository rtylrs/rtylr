#pragma once

#include <string>
#include <vips/vips8>

namespace voxire {
namespace ops {

// blurHash encodes a compact BlurHash string (https://blurha.sh) for the image,
// the tiny placeholder shown while the real asset loads. xComponents and
// yComponents (1..9) control the horizontal/vertical detail of the placeholder;
// 4x3 is a good default. The image is downscaled internally first.
std::string blurHash(vips::VImage img, int xComponents, int yComponents);

} // namespace ops
} // namespace voxire
