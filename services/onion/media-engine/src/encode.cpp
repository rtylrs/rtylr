#include "encode.hpp"

namespace voxire {

OutputFormat parseOutputFormat(const std::string& s) {
    if (s == "auto")          return OutputFormat::Auto;
    if (s == "webp")          return OutputFormat::WebP;
    if (s == "jpeg" || s == "jpg") return OutputFormat::JPEG;
    if (s == "png")           return OutputFormat::PNG;
    if (s == "avif")          return OutputFormat::AVIF;
    if (s == "gif")           return OutputFormat::GIF;
    if (s == "tiff" || s == "tif") return OutputFormat::TIFF;
    throw ProcessingError("unsupported output format: " + s);
}

std::string extensionFor(OutputFormat f) {
    switch (f) {
        case OutputFormat::Auto: return ".webp";
        case OutputFormat::WebP: return ".webp";
        case OutputFormat::JPEG: return ".jpg";
        case OutputFormat::PNG:  return ".png";
        case OutputFormat::AVIF: return ".avif";
        case OutputFormat::GIF:  return ".gif";
        case OutputFormat::TIFF: return ".tiff";
    }
    return ".webp";
}

OutputFormat resolveAutoFormat(bool hasAlpha) {
    // WebP handles both alpha and opaque content well at small sizes, so it is
    // the safe modern default for f_auto regardless of transparency.
    (void)hasAlpha;
    return OutputFormat::WebP;
}

int resolveAutoQuality(OutputFormat f) {
    switch (f) {
        case OutputFormat::JPEG: return 82;
        case OutputFormat::WebP: return 80;
        case OutputFormat::AVIF: return 60; // AV1 reaches similar quality lower
        default:                 return 85;
    }
}

void encodeToFile(vips::VImage img, const std::string& path,
                  const EncodeOptions& opts) {
    OutputFormat fmt = opts.format;
    if (fmt == OutputFormat::Auto) {
        fmt = resolveAutoFormat(img.has_alpha());
    }
    int q = opts.quality > 0 ? opts.quality : resolveAutoQuality(fmt);

    switch (fmt) {
        case OutputFormat::WebP:
            img.webpsave(path.c_str(), vips::VImage::option()
                ->set("Q", q)
                ->set("lossless", opts.lossless)
                ->set("effort", opts.webpEffort)
                ->set("strip", opts.stripMetadata));
            return;
        case OutputFormat::JPEG:
            img.jpegsave(path.c_str(), vips::VImage::option()
                ->set("Q", q)
                ->set("interlace", opts.progressive)
                ->set("strip", opts.stripMetadata));
            return;
        case OutputFormat::PNG:
            img.pngsave(path.c_str(), vips::VImage::option()
                ->set("compression", opts.pngCompression)
                ->set("interlace", opts.progressive)
                ->set("strip", opts.stripMetadata));
            return;
        case OutputFormat::AVIF:
            img.heifsave(path.c_str(), vips::VImage::option()
                ->set("Q", q)
                ->set("compression", VIPS_FOREIGN_HEIF_COMPRESSION_AV1)
                ->set("speed", opts.avifSpeed)
                ->set("lossless", opts.lossless)
                ->set("strip", opts.stripMetadata));
            return;
        case OutputFormat::GIF:
            img.gifsave(path.c_str(), vips::VImage::option()
                ->set("strip", opts.stripMetadata));
            return;
        case OutputFormat::TIFF:
            img.tiffsave(path.c_str(), vips::VImage::option()
                ->set("Q", q)
                ->set("strip", opts.stripMetadata));
            return;
        case OutputFormat::Auto:
            break; // resolved above
    }
    throw ProcessingError("unsupported output format in encodeToFile");
}

} // namespace voxire
