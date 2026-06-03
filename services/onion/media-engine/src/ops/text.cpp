#include "ops/text.hpp"
#include "common.hpp"

namespace voxire {
namespace ops {

vips::VImage renderText(const std::string& text, const std::string& font,
                        int maxWidth, const Color& color) {
    if (text.empty()) {
        throw ProcessingError("text overlay content must not be empty");
    }
    // text() yields a single-band anti-aliased coverage mask (0-255).
    auto opt = vips::VImage::option()->set("font", font.c_str())->set("dpi", 72);
    if (maxWidth > 0) {
        opt = opt->set("width", maxWidth);
    }
    vips::VImage mask = vips::VImage::text(text.c_str(), opt);

    // Build a solid color field the size of the mask, then attach the mask as
    // the alpha channel so the text takes the requested color.
    vips::VImage solid = mask.new_from_image(std::vector<double>{
        static_cast<double>(color.r),
        static_cast<double>(color.g),
        static_cast<double>(color.b),
    });
    return solid.bandjoin(mask);
}

} // namespace ops
} // namespace voxire
