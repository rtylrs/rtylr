#include "ops/blurhash.hpp"
#include "ops/band_util.hpp"
#include "common.hpp"
#include <glib.h>
#include <cmath>
#include <vector>
#include <array>
#include <algorithm>

namespace voxire {
namespace ops {

namespace {

const char* kChars =
    "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz#$%*+,-.:;=?@[]^_{|}~";

std::string base83(int value, int length) {
    std::string out(length, '0');
    for (int i = length - 1; i >= 0; --i) {
        const int digit = value % 83;
        value /= 83;
        out[i] = kChars[digit];
    }
    return out;
}

double sRGBToLinear(int v) {
    const double x = v / 255.0;
    return x <= 0.04045 ? x / 12.92 : std::pow((x + 0.055) / 1.055, 2.4);
}

// The sRGB transfer function is pure: build the 256-entry table once so the
// hot DCT loop never calls std::pow.
struct SRGBToLinearTable {
    std::array<double, 256> v{};
    SRGBToLinearTable() {
        for (int i = 0; i < 256; ++i) v[i] = sRGBToLinear(i);
    }
};
const SRGBToLinearTable kSRGBToLinear;

int linearTosRGB(double v) {
    const double x = v < 0 ? 0 : (v > 1 ? 1 : v);
    const double r = x <= 0.0031308 ? x * 12.92 : 1.055 * std::pow(x, 1.0 / 2.4) - 0.055;
    return static_cast<int>(r * 255 + 0.5);
}

double signPow(double val, double exp) {
    const double v = std::pow(std::fabs(val), exp);
    return val < 0 ? -v : v;
}

int quantiseAC(double v, double maxVal) {
    int q = static_cast<int>(std::floor(signPow(v / maxVal, 0.5) * 9 + 9.5));
    if (q < 0) q = 0;
    if (q > 18) q = 18;
    return q;
}

} // namespace

std::string blurHash(vips::VImage img, int xComponents, int yComponents) {
    if (xComponents < 1 || xComponents > 9 || yComponents < 1 || yComponents > 9) {
        throw ProcessingError("blurhash components must be between 1 and 9");
    }

    // Downscale to keep the transform cheap; drop alpha, force 8-bit sRGB.
    Split s = splitAlpha(img);
    vips::VImage small = s.color.colourspace(VIPS_INTERPRETATION_sRGB)
        .thumbnail_image(64, vips::VImage::option()->set("height", 64))
        .cast(VIPS_FORMAT_UCHAR);

    const int w = small.width();
    const int h = small.height();
    const int bands = small.bands();

    size_t size = 0;
    auto* data = static_cast<unsigned char*>(small.write_to_memory(&size));
    if (data == nullptr) {
        throw ProcessingError("blurhash: failed to read pixel data");
    }

    std::vector<std::array<double, 3>> factors(
        static_cast<size_t>(xComponents) * yComponents, {0, 0, 0});

    // Precompute the separable cosine bases: cos terms depend only on
    // (component, coordinate), never on the other axis or the pixel colour.
    // Evaluating them inside the pixel loops costs two std::cos calls per
    // component per pixel; hoisted here they cost (xc*w + yc*h) in total.
    std::vector<std::vector<double>> cosX(static_cast<size_t>(xComponents),
        std::vector<double>(w));
    std::vector<std::vector<double>> cosY(static_cast<size_t>(yComponents),
        std::vector<double>(h));
    for (int cx = 0; cx < xComponents; ++cx)
        for (int x = 0; x < w; ++x)
            cosX[cx][x] = std::cos(M_PI * cx * x / w);
    for (int cy = 0; cy < yComponents; ++cy)
        for (int y = 0; y < h; ++y)
            cosY[cy][y] = std::cos(M_PI * cy * y / h);

    for (int y = 0; y < h; ++y) {
        const unsigned char* row = data + static_cast<size_t>(y) * w * bands;
        for (int x = 0; x < w; ++x, row += bands) {
            // Linearise once per pixel; every factor reuses these values.
            const double lr = kSRGBToLinear.v[row[0]];
            const double lg = kSRGBToLinear.v[bands > 1 ? row[1] : row[0]];
            const double lb = kSRGBToLinear.v[bands > 2 ? row[2] : row[0]];
            for (int cy = 0; cy < yComponents; ++cy) {
                const double by = cosY[cy][y];
                auto* f = &factors[static_cast<size_t>(cy) * xComponents];
                for (int cx = 0; cx < xComponents; ++cx) {
                    const double basis =
                        ((cx == 0 && cy == 0) ? 1.0 : 2.0) * cosX[cx][x] * by;
                    f[cx][0] += basis * lr;
                    f[cx][1] += basis * lg;
                    f[cx][2] += basis * lb;
                }
            }
        }
    }
    g_free(data);

    // Normalise the accumulated sums (the old per-component scale, applied
    // once now that accumulation is pixel-major).
    const double norm = 1.0 / (static_cast<double>(w) * h);
    for (auto& f : factors) {
        f[0] *= norm;
        f[1] *= norm;
        f[2] *= norm;
    }

    const int acCount = xComponents * yComponents - 1;
    std::string hash;

    // size flag
    hash += base83((xComponents - 1) + (yComponents - 1) * 9, 1);

    // quantised maximum AC value
    double maxVal = 1.0;
    if (acCount > 0) {
        double actualMax = 0.0;
        for (int i = 1; i <= acCount; ++i) {
            for (int c = 0; c < 3; ++c) {
                actualMax = std::max(actualMax, std::fabs(factors[i][c]));
            }
        }
        int quantMax = static_cast<int>(std::floor(actualMax * 166 - 0.5));
        if (quantMax < 0) quantMax = 0;
        if (quantMax > 82) quantMax = 82;
        maxVal = (quantMax + 1) / 166.0;
        hash += base83(quantMax, 1);
    } else {
        hash += base83(0, 1);
    }

    // DC
    const auto& dc = factors[0];
    const int dcVal = (linearTosRGB(dc[0]) << 16) +
                      (linearTosRGB(dc[1]) << 8) + linearTosRGB(dc[2]);
    hash += base83(dcVal, 4);

    // AC
    for (int i = 1; i <= acCount; ++i) {
        const auto& f = factors[i];
        const int acVal = quantiseAC(f[0], maxVal) * 19 * 19 +
                          quantiseAC(f[1], maxVal) * 19 +
                          quantiseAC(f[2], maxVal);
        hash += base83(acVal, 2);
    }
    return hash;
}

} // namespace ops
} // namespace voxire
