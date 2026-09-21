#include "image_processor.hpp"
#include "encode.hpp"
#include <vips/vips8>
#include <algorithm>
#include <cmath>
#include <stdexcept>

namespace voxire {

void ImageProcessor::process(const ImageOptions& opts) {
    if (opts.inputPath.empty() || opts.outputPath.empty()) {
        throw ProcessingError("input and output paths are required");
    }
    if (opts.width == 0 && opts.height == 0) {
        throw ProcessingError("at least one of width or height must be non-zero");
    }
    if (opts.width < 0 || opts.height < 0) {
        throw ProcessingError("width and height must not be negative");
    }
    if (opts.quality < 1 || opts.quality > 100) {
        throw ProcessingError("quality must be between 1 and 100");
    }

    vips::VImage img = vips::VImage::new_from_file(
        opts.inputPath.c_str(),
        vips::VImage::option()->set("access", VIPS_ACCESS_SEQUENTIAL)
    );

    // Honour the EXIF orientation tag before resizing so output from phone
    // cameras isn't rotated or mirrored.
    img = img.autorot();

    img = resizeImage(img, opts.width, opts.height, opts.fit);
    saveImage(img, opts);
}

vips::VImage ImageProcessor::resizeImage(vips::VImage img, int targetW, int targetH, FitMode fit) {
    int srcW = img.width();
    int srcH = img.height();

    if (targetW == 0) targetW = static_cast<int>(srcW * (static_cast<double>(targetH) / srcH));
    if (targetH == 0) targetH = static_cast<int>(srcH * (static_cast<double>(targetW) / srcW));

    double scaleW = static_cast<double>(targetW) / srcW;
    double scaleH = static_cast<double>(targetH) / srcH;

    VipsSize size = VIPS_SIZE_BOTH;
    double scale  = 1.0;

    switch (fit) {
        case FitMode::Cover:
            scale = std::max(scaleW, scaleH);
            size  = VIPS_SIZE_BOTH;
            break;
        case FitMode::Contain:
            // Fit within the box, scaling up small images to fill it.
            scale = std::min(scaleW, scaleH);
            size  = VIPS_SIZE_BOTH;
            break;
        case FitMode::Inside:
            // Fit within the box but never enlarge beyond the source.
            scale = std::min(scaleW, scaleH);
            size  = VIPS_SIZE_DOWN;
            break;
        case FitMode::Outside:
            scale = std::max(scaleW, scaleH);
            size  = VIPS_SIZE_UP;
            break;
        case FitMode::Fill:
            // Stretch to the exact target dimensions, distorting the aspect
            // ratio. VIPS_SIZE_FORCE drives width and height independently; the
            // previous centre-crop produced cover behaviour, not a fill.
            return img.thumbnail_image(
                targetW,
                vips::VImage::option()
                    ->set("height", targetH)
                    ->set("size", VIPS_SIZE_FORCE)
            );
    }

    // Round the thumbnail box up. vips_thumbnail without "crop" fits the image
    // inside the box with one equalised shrink factor and rounds the resulting
    // axes, so a truncated box lets an axis come out a pixel below the target —
    // 4032x3024 -> cover 750x500 requested a 750x562 box, produced a 749x562
    // thumbnail, and the crop below failed with "bad extract area".
    img = img.thumbnail_image(
        static_cast<int>(std::ceil(srcW * scale)),
        vips::VImage::option()
            ->set("height", static_cast<int>(std::ceil(srcH * scale)))
            ->set("size", size)
    );

    if (fit == FitMode::Cover) {
        // Centre-crop to exact target dimensions, clamped so a future rounding
        // surprise degrades to a slightly smaller image instead of throwing.
        int x = std::max(0, (img.width()  - targetW) / 2);
        int y = std::max(0, (img.height() - targetH) / 2);
        img = img.crop(x, y, std::min(targetW, img.width()), std::min(targetH, img.height()));
    }

    return img;
}

void ImageProcessor::saveImage(vips::VImage img, const ImageOptions& opts) {
    // Delegate to the shared encoder so format-specific tuning (webp effort,
    // avif speed, png level) lives in exactly one place.
    EncodeOptions enc;
    if (opts.format == "webp")        enc.format = OutputFormat::WebP;
    else if (opts.format == "jpeg" ||
             opts.format == "jpg")    enc.format = OutputFormat::JPEG;
    else if (opts.format == "png")    enc.format = OutputFormat::PNG;
    else if (opts.format == "avif")   enc.format = OutputFormat::AVIF;
    else throw ProcessingError("unsupported output format: " + opts.format);
    enc.quality = opts.quality;
    encodeToFile(std::move(img), opts.outputPath, enc);
}

} // namespace voxire
