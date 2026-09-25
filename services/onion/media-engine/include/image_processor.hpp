#pragma once

#include "common.hpp"
#include <string>
#include <vips/vips8>

namespace voxire {

struct ImageOptions {
    std::string inputPath;
    std::string outputPath;
    std::string format;   // webp, jpeg, png, avif
    int         width   = 0;
    int         height  = 0;
    int         quality = 85;
    FitMode     fit     = FitMode::Cover;
};

class ImageProcessor {
public:
    // process performs resize, format conversion, and quality optimization.
    // Throws ProcessingError on failure.
    void process(const ImageOptions& opts);

private:
    vips::VImage resizeImage(vips::VImage img, int targetW, int targetH, FitMode fit);
    void         saveImage(vips::VImage img, const ImageOptions& opts);
};

} // namespace voxire
