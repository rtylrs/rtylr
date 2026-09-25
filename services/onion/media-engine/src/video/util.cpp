#include "video/util.hpp"
#include "ffmpeg.hpp"
#include "process.hpp"
#include "common.hpp"

namespace voxire {
namespace video {

std::string scaleFilter(int width, int height) {
    const std::string w = width > 0 ? std::to_string(width) : "-1";
    const std::string h = height > 0 ? std::to_string(height) : "-1";
    return "scale=" + w + ":" + h;
}

void runFFmpeg(const std::vector<std::string>& args) {
    std::vector<std::string> cmd;
    cmd.reserve(args.size() + 2);
    cmd.push_back(ffmpegBin());
    cmd.push_back("-y"); // overwrite output without prompting
    for (const auto& a : args) {
        cmd.push_back(a);
    }
    const ProcResult r = runProcess(cmd);
    if (!r.ok()) {
        throw ProcessingError("ffmpeg failed (exit " + std::to_string(r.exitCode) +
                              "): " + r.output);
    }
}

} // namespace video
} // namespace voxire
