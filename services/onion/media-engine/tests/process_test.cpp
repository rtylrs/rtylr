#include "process.hpp"
#include "check.hpp"

using namespace voxire;

int main() {
    // Regression: runProcess passed the wait status to reapWithTimeout by
    // value, so waitpid wrote into a copy and every child looked like it had
    // exited 0. Failed ffmpeg/ffprobe jobs were reported as success.
    const ProcResult failed = runProcess({"sh", "-c", "exit 3"});
    CHECK_EQ(failed.exitCode, 3);
    CHECK(!failed.ok());

    const ProcResult ok = runProcess({"sh", "-c", "echo captured; exit 0"});
    CHECK_EQ(ok.exitCode, 0);
    CHECK(ok.ok());
    CHECK(ok.output.find("captured") != std::string::npos);

    // A failed exec is reported through the child's _exit(127).
    const ProcResult missing = runProcess({"voxire-no-such-binary"});
    CHECK_EQ(missing.exitCode, 127);
    CHECK(!missing.ok());

    // A signal-terminated child is not a successful exit.
    const ProcResult killed = runProcess({"sh", "-c", "kill -TERM $$"});
    CHECK_EQ(killed.exitCode, -1);
    CHECK(!killed.ok());

    RUN_TESTS();
}
