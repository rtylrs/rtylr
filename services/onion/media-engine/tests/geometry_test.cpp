#include "geometry.hpp"
#include "check.hpp"

using namespace voxire;

int main() {
    Size full = parseSize("256x128");
    CHECK_EQ(full.width, 256);
    CHECK_EQ(full.height, 128);

    Size wOnly = parseSize("256x");
    CHECK_EQ(wOnly.width, 256);
    CHECK_EQ(wOnly.height, 0);

    Size hOnly = parseSize("x128");
    CHECK_EQ(hOnly.width, 0);
    CHECK_EQ(hOnly.height, 128);

    CHECK(Size{}.empty());
    CHECK_EQ(Size({4, 5}).area(), 20L);

    CHECK_THROWS(parseSize("nope"));
    CHECK_THROWS(parseSize("-1x10"));

    Box box{2, 3, 10, 20};
    CHECK_EQ(box.right(), 12);
    CHECK_EQ(box.bottom(), 23);
    CHECK(!box.empty());
    CHECK(Box({0, 0, 0, 5}).empty());

    RUN_TESTS();
}
