#include "numeric.hpp"
#include "check.hpp"

using namespace voxire;

int main() {
    CHECK_EQ(clampInt(5, 0, 10), 5);
    CHECK_EQ(clampInt(-3, 0, 10), 0);
    CHECK_EQ(clampInt(99, 0, 10), 10);

    CHECK_EQ(clampDouble(0.5, 0.0, 1.0), 0.5);
    CHECK_EQ(clampDouble(2.0, 0.0, 1.0), 1.0);
    CHECK_EQ(clampDouble(-1.0, 0.0, 1.0), 0.0);

    CHECK_EQ(parseIntStrict("42", "x"), 42);
    CHECK_EQ(parseIntStrict("-7", "x"), -7);
    CHECK_THROWS(parseIntStrict("12abc", "x"));
    CHECK_THROWS(parseIntStrict("", "x"));

    CHECK_EQ(parseDoubleStrict("1.5", "x"), 1.5);
    CHECK_THROWS(parseDoubleStrict("1.5x", "x"));

    RUN_TESTS();
}
