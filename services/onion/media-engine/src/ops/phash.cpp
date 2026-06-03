#include "ops/phash.hpp"

namespace voxire {
namespace ops {

uint64_t averageHash(vips::VImage img) {
    // Reduce to 8x8 grayscale via thumbnail (ignoring aspect ratio).
    vips::VImage small = img.thumbnail_image(8, vips::VImage::option()
        ->set("height", 8)
        ->set("size", VIPS_SIZE_FORCE))
        .colourspace(VIPS_INTERPRETATION_B_W)
        .extract_band(0);

    double values[64];
    double sum = 0.0;
    for (int y = 0; y < 8; ++y) {
        for (int x = 0; x < 8; ++x) {
            std::vector<double> p = small.getpoint(x, y);
            const double v = p.empty() ? 0.0 : p[0];
            values[y * 8 + x] = v;
            sum += v;
        }
    }
    const double mean = sum / 64.0;

    uint64_t hash = 0;
    for (int i = 0; i < 64; ++i) {
        if (values[i] >= mean) {
            hash |= (uint64_t(1) << i);
        }
    }
    return hash;
}

} // namespace ops
} // namespace voxire
