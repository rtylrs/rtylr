#include "gravity.hpp"
#include "check.hpp"

using namespace voxire;

int main() {
    CHECK(parseGravity("center") == Gravity::Center);
    CHECK(parseGravity("top") == Gravity::North);
    CHECK(parseGravity("bottom_right") == Gravity::SouthEast);
    CHECK_THROWS(parseGravity("sideways"));

    int x = -1, y = -1;
    // Center a 100x100 box inside 300x300 -> (100,100).
    offsetFor(Gravity::Center, 100, 100, 300, 300, x, y);
    CHECK_EQ(x, 100);
    CHECK_EQ(y, 100);

    offsetFor(Gravity::NorthWest, 100, 100, 300, 300, x, y);
    CHECK_EQ(x, 0);
    CHECK_EQ(y, 0);

    offsetFor(Gravity::SouthEast, 100, 100, 300, 300, x, y);
    CHECK_EQ(x, 200);
    CHECK_EQ(y, 200);

    offsetFor(Gravity::North, 100, 100, 300, 300, x, y);
    CHECK_EQ(x, 100);
    CHECK_EQ(y, 0);

    RUN_TESTS();
}
