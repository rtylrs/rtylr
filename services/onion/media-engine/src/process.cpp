#include "process.hpp"
#include "common.hpp"

#include <unistd.h>
#include <sys/wait.h>
#include <cerrno>
#include <csignal>
#include <vector>

namespace voxire {

namespace {

// A wedged subprocess must never stall the resident --serve loop forever:
// kill it after this many seconds.
constexpr int kProcessTimeoutSeconds = 120;

// Cap captured combined stdout+stderr so a chatty child cannot grow the
// resident process's RSS without bound.
constexpr size_t kMaxCapturedOutput = 1 << 20; // 1 MiB

// reapWithTimeout waits for the child, killing it (then SIGKILL) once the
// deadline passes. The wait status is written through the reference so the
// caller can read the child's real exit code. Returns true if the child was
// reaped.
bool reapWithTimeout(pid_t pid, int& status) {
    for (int elapsed = 0; elapsed < kProcessTimeoutSeconds; ++elapsed) {
        const pid_t r = waitpid(pid, &status, WNOHANG);
        if (r == pid) {
            return true;
        }
        if (r < 0 && errno != EINTR) {
            return false;
        }
        sleep(1);
    }
    ::kill(pid, SIGTERM);
    for (int grace = 0; grace < 5; ++grace) {
        const pid_t r = waitpid(pid, &status, WNOHANG);
        if (r == pid) {
            return true;
        }
        sleep(1);
    }
    ::kill(pid, SIGKILL);
    waitpid(pid, &status, 0);
    return true;
}

} // namespace

ProcResult runProcess(const std::vector<std::string>& args) {
    if (args.empty()) {
        throw ProcessingError("runProcess called with no command");
    }

    int pipefd[2];
    if (pipe(pipefd) != 0) {
        throw ProcessingError("failed to create pipe for subprocess");
    }

    const pid_t pid = fork();
    if (pid < 0) {
        ::close(pipefd[0]);
        ::close(pipefd[1]);
        throw ProcessingError("failed to fork subprocess");
    }

    if (pid == 0) {
        // Child: redirect stdout+stderr into the pipe, then exec.
        dup2(pipefd[1], STDOUT_FILENO);
        dup2(pipefd[1], STDERR_FILENO);
        ::close(pipefd[0]);
        ::close(pipefd[1]);

        std::vector<char*> argv;
        argv.reserve(args.size() + 1);
        for (const auto& a : args) {
            argv.push_back(const_cast<char*>(a.c_str()));
        }
        argv.push_back(nullptr);
        execvp(argv[0], argv.data());
        _exit(127); // exec failed
    }

    // Parent: read everything the child writes, bounded by the capture cap.
    ::close(pipefd[1]);
    ProcResult result;
    char buf[4096];
    ssize_t n;
    while ((n = read(pipefd[0], buf, sizeof(buf))) > 0) {
        if (result.output.size() < kMaxCapturedOutput) {
            const size_t room = kMaxCapturedOutput - result.output.size();
            result.output.append(buf, static_cast<size_t>(n) > room ? room : static_cast<size_t>(n));
        }
        // Past the cap we keep draining the pipe (so the child's writes never
        // block on a full pipe and it can still exit cleanly) but discard.
    }
    ::close(pipefd[0]);

    int status = 0;
    if (!reapWithTimeout(pid, status)) {
        // waitpid failed for a reason other than EINTR, so the child's status is
        // unknown: report failure rather than the zero-initialized status.
        result.exitCode = -1;
        return result;
    }
    result.exitCode = WIFEXITED(status) ? WEXITSTATUS(status) : -1;
    return result;
}

} // namespace voxire
