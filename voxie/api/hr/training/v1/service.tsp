import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.training.v1;

@route("/hr/training/v1")
interface TrainingService {
  @get
  @route("/courses")
  listTrainingCourses(
    ...ListTrainingCoursesRequest,
  ): ListTrainingCoursesResponse;

  @post
  @route("/courses")
  createTrainingCourse(
    ...CreateTrainingCourseRequest,
  ): CreateTrainingCourseResponse;
}
