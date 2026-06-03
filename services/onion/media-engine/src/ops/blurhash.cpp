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

    for (int cy = 0; cy < yComponents; ++cy) {
        for (int cx = 0; cx < xComponents; ++cx) {
            const double norm = (cx == 0 && cy == 0) ? 1.0 : 2.0;
            double r = 0, g = 0, b = 0;
            for (int y = 0; y < h; ++y) {
                for (int x = 0; x < w; ++x) {
                    const double basis = norm *
                        std::cos(M_PI * cx * x / w) *
                        std::cos(M_PI * cy * y / h);
                    const unsigned char* px = data + (y * w + x) * bands;
                    r += basis * sRGBToLinear(px[0]);
                    g += basis * sRGBToLinear(bands > 1 ? px[1] : px[0]);
                    b += basis * sRGBToLinear(bands > 2 ? px[2] : px[0]);
                }
            }
            const double scale = 1.0 / (w * h);
            factors[cy * xComponents + cx] = {r * scale, g * scale, b * scale};
        }
    }
    g_free(data);

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
