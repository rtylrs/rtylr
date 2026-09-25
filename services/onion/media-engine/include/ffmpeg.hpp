#pragma once

#include <string>

namespace voxire {

// ffmpegBin returns the ffmpeg executable to invoke: the FFMPEG_BIN environment
// variable when set, otherwise "ffmpeg" (resolved via PATH).
std::string ffmpegBin();

// ffprobeBin returns the ffprobe executable: FFPROBE_BIN, else "ffprobe".
std::string ffprobeBin();

} // namespace voxire
