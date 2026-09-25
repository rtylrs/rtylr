#pragma once

// Minimal dependency-free test harness. Each test file defines main(), runs
// CHECK macros, and ends with RUN_TESTS() to return a non-zero exit on failure.
#include <iostream>

namespace voxtest {
inline int failures = 0;
}

#define CHECK(cond)                                                            \
    do {                                                                       \
        if (!(cond)) {                                                         \
            ++voxtest::failures;                                               \
            std::cerr << "CHECK failed: " #cond " at " << __FILE__ << ":"      \
                      << __LINE__ << "\n";                                     \
        }                                                                      \
    } while (0)

#define CHECK_EQ(a, b)                                                         \
    do {                                                                       \
        auto _va = (a);                                                        \
        auto _vb = (b);                                                        \
        if (!(_va == _vb)) {                                                   \
            ++voxtest::failures;                                               \
            std::cerr << "CHECK_EQ failed: " #a " == " #b " (" << _va          \
                      << " vs " << _vb << ") at " << __FILE__ << ":"           \
                      << __LINE__ << "\n";                                     \
        }                                                                      \
    } while (0)

#define CHECK_THROWS(expr)                                                     \
    do {                                                                       \
        bool _threw = false;                                                   \
        try {                                                                  \
            (void)(expr);                                                      \
        } catch (...) {                                                        \
            _threw = true;                                                     \
        }                                                                      \
        if (!_threw) {                                                         \
            ++voxtest::failures;                                               \
            std::cerr << "CHECK_THROWS failed: " #expr " did not throw at "    \
                      << __FILE__ << ":" << __LINE__ << "\n";                  \
        }                                                                      \
    } while (0)

#define RUN_TESTS() return voxtest::failures == 0 ? 0 : 1
