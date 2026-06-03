#include "video/trim.hpp"
#include "video/util.hpp"
#include "common.hpp"

namespace voxire {
namespace video {

void trim(const std::string& inputPath, const std::string& outputPath,
          double startSecs, double durationSecs) {
    if (startSecs < 0) {
        throw ProcessingError("trim start must not be negative");
    }
    if (durationSecs <= 0) {
        throw ProcessingError("trim duration must be greater than 0");
    }
    // -ss before -i seeks quickly; -t bounds the copied duration. Stream copy
    // avoids a re-encode (cuts land on the nearest keyframe).
    runFFmpeg({
        "-ss", std::to_string(startSecs),
        "-i", inputPath,
        "-t", std::to_string(durationSecs),
        "-c", "copy",
        outputPath,
    });
}

} // namespace video
} // namespace voxire
