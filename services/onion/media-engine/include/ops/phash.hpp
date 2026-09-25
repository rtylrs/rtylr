#pragma once

#include <cstdint>
#include <vips/vips8>

namespace voxire {
namespace ops {

// averageHash computes a 64-bit perceptual "aHash": the image is reduced to
// 8x8 grayscale and each pixel becomes a bit set when it is at or above the
// 64-pixel mean. Hamming distance between two hashes estimates visual
// similarity (useful for dedup / near-duplicate detection).
uint64_t averageHash(vips::VImage img);

} // namespace ops
} // namespace voxire
