#pragma once

#include "common.hpp"
#include <string>
#include <vips/vips8>

namespace voxire {

// OutputFormat enumerates the encoders the engine can write. Auto defers the
// choice to resolveAutoFormat (the f_auto behavior).
enum class OutputFormat {
    Auto,
    WebP,
    JPEG,
    PNG,
    AVIF,
    GIF,
    TIFF,
};

OutputFormat parseOutputFormat(const std::string& s);

// extensionFor returns the canonical file extension (with leading dot) for a
// concrete format. Auto resolves to webp.
std::string extensionFor(OutputFormat f);

// EncodeOptions controls how an image is written to disk.
struct EncodeOptions {
    OutputFormat format       = OutputFormat::WebP;
    int          quality      = 0;     // 0 => q_auto (resolveAutoQuality)
    bool         lossless     = false; // webp/avif
    bool         progressive  = false; // jpeg/png interlacing
    bool         stripMetadata = true; // drop EXIF/ICC unless explicitly kept

    // Encoder effort knobs, tuned for delivery latency by default:
    //   webpEffort     0-6 (libwebp method; default 4). 2 is ~3x faster than
    //                  the default for ~2% larger files.
    //   avifSpeed      0-9 (libaom usage; unset is glacial). 6 is ~3x faster
    //                  than the default at modest size cost.
    //   pngCompression 0-9 zlib level (default 6). Lower trades size for speed.
    int          webpEffort      = 2;
    int          avifSpeed       = 6;
    int          pngCompression  = 6;
};

// resolveAutoFormat picks a concrete format for OutputFormat::Auto given whether
// the image carries transparency. Prefers modern lossy codecs.
OutputFormat resolveAutoFormat(bool hasAlpha);

// resolveAutoQuality returns a sensible default quality (1-100) per format,
// used when EncodeOptions.quality is 0.
int resolveAutoQuality(OutputFormat f);

// encodeToFile writes img to path using opts, dispatching to the right libvips
// saver. Throws ProcessingError on an unsupported format.
void encodeToFile(vips::VImage img, const std::string& path,
                  const EncodeOptions& opts);

} // namespace voxire
