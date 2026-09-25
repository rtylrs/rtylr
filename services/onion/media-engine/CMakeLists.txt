cmake_minimum_required(VERSION 3.20)
project(media-engine VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Release build by default
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

# -O3 measurably helps the pure-C++ pixel loops (blurhash/phash DCT, parsing).
set(CMAKE_CXX_FLAGS_RELEASE "-O3 -DNDEBUG")

# Link-time optimisation: inlines the ops/ and transform/ call chains.
# Off by default for reproducible debuggability; enable with -DMEDIA_ENGINE_LTO=ON.
option(MEDIA_ENGINE_LTO "Enable link-time optimisation" OFF)
if(MEDIA_ENGINE_LTO)
    include(CheckIPOSupported)
    check_ipo_supported(RESULT _ipo_ok OUTPUT _ipo_msg)
    if(_ipo_ok)
        set(CMAKE_INTERPROCEDURAL_OPTIMIZATION ON)
    else()
        message(WARNING "LTO requested but not supported: ${_ipo_msg}")
    endif()
endif()

# Native CPU tuning for the deployment host. OFF by default so artifacts stay
# portable; the worker Dockerfile can pass -DMEDIA_ENGINE_ARCH=native.
set(MEDIA_ENGINE_ARCH "generic" CACHE STRING "CPU architecture tuning (generic|native)")
if(MEDIA_ENGINE_ARCH STREQUAL "native")
    add_compile_options(-march=native)
endif()

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
file(GLOB_RECURSE CORE_SOURCES CONFIGURE_DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/src/*.cpp")
list(FILTER CORE_SOURCES EXCLUDE REGEX ".*/main\\.cpp$")

# Core processing library, shared by the CLI binary and the benchmark harness.
add_library(media_engine_core STATIC ${CORE_SOURCES})

target_include_directories(media_engine_core PUBLIC
    include
    ${VIPS_INCLUDE_DIRS}
)

target_link_libraries(media_engine_core PUBLIC
    ${VIPS_LIBRARIES}
)

target_compile_options(media_engine_core PRIVATE
    ${VIPS_CFLAGS_OTHER}
    -Wall
    -Wextra
    -Wpedantic
)

add_executable(media-engine src/main.cpp)
target_link_libraries(media-engine PRIVATE media_engine_core)

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

# ── Benchmarks ───────────────────────────────────────────────────────────────

# Timing harness for the main processing paths. Run before/after a change:
#   ./media-engine-bench --iters 7
option(MEDIA_ENGINE_BENCH "Build the media-engine benchmark harness" ON)
if(MEDIA_ENGINE_BENCH)
    add_executable(media-engine-bench bench/benchmark.cpp)
    target_link_libraries(media-engine-bench PRIVATE media_engine_core)
endif()

# ── Tests ───────────────────────────────────────────────────────────────────

# Unit tests for the pure (vips-free) primitives. Each tests/*_test.cpp builds
# into its own executable and registers with CTest; run with `ctest`.
option(MEDIA_ENGINE_TESTS "Build the media-engine unit tests" ON)
if(MEDIA_ENGINE_TESTS)
    enable_testing()
    file(GLOB TEST_SOURCES CONFIGURE_DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/tests/*_test.cpp")
    # process_test covers the subprocess runner, which lives in a vips-free
    # translation unit; it is built explicitly with that source instead of the
    # header-only glob.
    list(FILTER TEST_SOURCES EXCLUDE REGEX ".*/process_test\\.cpp$")
    add_executable(process_test tests/process_test.cpp src/process.cpp)
    target_include_directories(process_test PRIVATE include tests)
    add_test(NAME process_test COMMAND process_test)
    # image_processor_test exercises the vips-backed resize path end to end
    # (synthetic source -> process -> decoded output), so it links the core
    # library and the vips bindings instead of being header-only.
    list(FILTER TEST_SOURCES EXCLUDE REGEX ".*/image_processor_test\\.cpp$")
    add_executable(image_processor_test tests/image_processor_test.cpp)
    target_link_libraries(image_processor_test PRIVATE media_engine_core)
    target_include_directories(image_processor_test PRIVATE include tests)
    add_test(NAME image_processor_test COMMAND image_processor_test)
    foreach(test_src ${TEST_SOURCES})
        get_filename_component(test_name ${test_src} NAME_WE)
        add_executable(${test_name} ${test_src})
        target_include_directories(${test_name} PRIVATE include tests)
        add_test(NAME ${test_name} COMMAND ${test_name})
    endforeach()
endif()

# ── Install ───────────────────────────────────────────────────────────────────

install(TARGETS media-engine DESTINATION bin)
