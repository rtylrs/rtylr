cmake_minimum_required(VERSION 3.20)
project(media-engine VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Release build by default
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

set(CMAKE_CXX_FLAGS_RELEASE "-O2 -DNDEBUG")

# ── Dependencies ─────────────────────────────────────────────────────────────

find_package(PkgConfig REQUIRED)

# libvips C++ bindings
pkg_check_modules(VIPS REQUIRED vips-cpp)

# FFmpeg (optional — used by video_processor when activated)
pkg_check_modules(FFMPEG QUIET
    libavcodec
    libavformat
    libavutil
    libswscale
)

# ── Sources ───────────────────────────────────────────────────────────────────

# Glob every translation unit under src/ (top level plus ops/ and video/) so new
# operations are picked up automatically. CONFIGURE_DEPENDS re-globs on build.
file(GLOB_RECURSE SOURCES CONFIGURE_DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/src/*.cpp")

add_executable(media-engine ${SOURCES})

target_include_directories(media-engine PRIVATE
    include
    ${VIPS_INCLUDE_DIRS}
)

target_link_libraries(media-engine PRIVATE
    ${VIPS_LIBRARIES}
)

target_compile_options(media-engine PRIVATE
    ${VIPS_CFLAGS_OTHER}
    -Wall
    -Wextra
    -Wpedantic
)

# Surface the project version to the binary (used by the `version` subcommand).
target_compile_definitions(media-engine PRIVATE
    MEDIA_ENGINE_VERSION="${PROJECT_VERSION}"
)

if(FFMPEG_FOUND)
    target_compile_definitions(media-engine PRIVATE VOXIRE_FFMPEG_AVAILABLE)
    target_include_directories(media-engine PRIVATE ${FFMPEG_INCLUDE_DIRS})
    target_link_libraries(media-engine PRIVATE ${FFMPEG_LIBRARIES})
    message(STATUS "FFmpeg found — video processing support enabled")
else()
    message(STATUS "FFmpeg not found — video processing will be scaffolded only")
endif()

# ── Tests ───────────────────────────────────────────────────────────────────

# Unit tests for the pure (vips-free) primitives. Each tests/*_test.cpp builds
# into its own executable and registers with CTest; run with `ctest`.
option(MEDIA_ENGINE_TESTS "Build the media-engine unit tests" ON)
if(MEDIA_ENGINE_TESTS)
    enable_testing()
    file(GLOB TEST_SOURCES CONFIGURE_DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/tests/*_test.cpp")
    foreach(test_src ${TEST_SOURCES})
        get_filename_component(test_name ${test_src} NAME_WE)
        add_executable(${test_name} ${test_src})
        target_include_directories(${test_name} PRIVATE include tests)
        add_test(NAME ${test_name} COMMAND ${test_name})
    endforeach()
endif()

# ── Install ───────────────────────────────────────────────────────────────────

install(TARGETS media-engine DESTINATION bin)
