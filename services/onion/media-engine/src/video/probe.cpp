#include "video/probe.hpp"
#include "ffmpeg.hpp"
#include "process.hpp"
#include "common.hpp"

namespace voxire {
namespace video {

std::string probe(const std::string& inputPath) {
    const ProcResult r = runProcess({
        ffprobeBin(),
        "-v", "quiet",
        "-print_format", "json",
        "-show_format",
        "-show_streams",
        inputPath,
    });
    if (!r.ok()) {
        throw ProcessingError("ffprobe failed (exit " +
                              std::to_string(r.exitCode) + "): " + r.output);
    }
    return r.output;
}

} // namespace video
} // namespace voxire
