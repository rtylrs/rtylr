#include "ops/phash.hpp"
#include "common.hpp"
#include <glib.h>

namespace voxire {
namespace ops {

uint64_t averageHash(vips::VImage img) {
    // Alpha is discarded below, so peel it off before resizing: thumbnail
    // premultiplies/unpremultiplies alpha bands, which dominated the runtime
    // for no benefit on the hash path.
    if (img.bands() > 3 && img.has_alpha()) {
        img = img.extract_band(0,
            vips::VImage::option()->set("n", img.bands() - 1));
    }

    // Reduce to 8x8 grayscale via thumbnail (ignoring aspect ratio).
    vips::VImage small = img.thumbnail_image(8, vips::VImage::option()
        ->set("height", 8)
        ->set("size", VIPS_SIZE_FORCE))
        .colourspace(VIPS_INTERPRETATION_B_W)
        .extract_band(0)
        .cast(VIPS_FORMAT_DOUBLE);

    // Materialise the whole 64-pixel image once. Reading pixels lazily (e.g.
    // getpoint per pixel) re-runs the pipeline for every point, which breaks
    // sequential-access sources outright and is orders of magnitude slower.
    size_t size = 0;
    double* data = static_cast<double*>(small.write_to_memory(&size));
    if (data == nullptr) {
        throw ProcessingError("phash: failed to read pixel data");
    }

    double sum = 0.0;
    for (int i = 0; i < 64; ++i) {
        sum += data[i];
    }
    const double mean = sum / 64.0;

    uint64_t hash = 0;
    for (int i = 0; i < 64; ++i) {
        if (data[i] >= mean) {
            hash |= (uint64_t(1) << i);
        }
    }
    g_free(data);
    return hash;
}

} // namespace ops
} // namespace voxire
