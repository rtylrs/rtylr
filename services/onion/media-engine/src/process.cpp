#include "process.hpp"
#include "common.hpp"

#include <unistd.h>
#include <sys/wait.h>
#include <vector>

namespace voxire {

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

    // Parent: read everything the child writes.
    ::close(pipefd[1]);
    ProcResult result;
    char buf[4096];
    ssize_t n;
    while ((n = read(pipefd[0], buf, sizeof(buf))) > 0) {
        result.output.append(buf, static_cast<size_t>(n));
    }
    ::close(pipefd[0]);

    int status = 0;
    waitpid(pid, &status, 0);
    result.exitCode = WIFEXITED(status) ? WEXITSTATUS(status) : -1;
    return result;
}

} // namespace voxire
