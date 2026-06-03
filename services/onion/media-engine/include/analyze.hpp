#pragma once

#include <string>
#include <vips/vips8>

namespace voxire {

// ImageInfo is the basic metadata reported by the `info` subcommand.
struct ImageInfo {
    int  width  = 0;
    int  height = 0;
    int  bands  = 0;
    bool hasAlpha = false;
    std::string interpretation;
};

// probeInfo reads dimensions, band count, alpha, and colorspace from img.
ImageInfo probeInfo(vips::VImage img);

// infoToJSON serializes ImageInfo as a compact JSON object.
std::string infoToJSON(const ImageInfo& info);

} // namespace voxire
