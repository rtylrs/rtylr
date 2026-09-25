#include "ops/duotone.hpp"
#include "ops/band_util.hpp"

namespace voxire {
namespace ops {

vips::VImage duotone(vips::VImage img, const Color& shadow, const Color& highlight) {
    Split s = splitAlpha(img);

    // Luminance in 0..1, replicated to three bands; its complement weights the
    // shadow color, itself weights the highlight color.
    vips::VImage lum = s.color.colourspace(VIPS_INTERPRETATION_B_W)
                              .extract_band(0)
                              .linear(1.0 / 255.0, 0.0);
    vips::VImage lum3 = lum.bandjoin(lum).bandjoin(lum);
    vips::VImage inv3 = lum3.linear(-1.0, 1.0); // 1 - lum

    vips::VImage shadowField = s.color.new_from_image(std::vector<double>{
        static_cast<double>(shadow.r), static_cast<double>(shadow.g),
        static_cast<double>(shadow.b)});
    vips::VImage highlightField = s.color.new_from_image(std::vector<double>{
        static_cast<double>(highlight.r), static_cast<double>(highlight.g),
        static_cast<double>(highlight.b)});

    vips::VImage out = (shadowField * inv3) + (highlightField * lum3);
    return joinAlpha(out, s);
}

} // namespace ops
} // namespace voxire
