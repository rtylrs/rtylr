#pragma once

#include <string>
#include <vector>

namespace voxire {

// ProcResult captures the outcome of a subprocess: its exit code and the
// combined stdout+stderr it produced.
struct ProcResult {
    int exitCode = -1;
    std::string output;

    bool ok() const { return exitCode == 0; }
};

// runProcess executes args[0] with the remaining args as argv (no shell, so
// arguments are passed verbatim with no injection risk), capturing combined
// output. Throws ProcessingError if the process cannot be spawned.
ProcResult runProcess(const std::vector<std::string>& args);

} // namespace voxire
