import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.training.v1;

/** TrainingCourse is a training course offered to staff. */
model TrainingCourse {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;
  description: string;
  duration_hours: float64;
  is_required: boolean;
  created_at: string;
}

model ListTrainingCoursesRequest {}

model ListTrainingCoursesResponse {
  courses: TrainingCourse[];
}

model CreateTrainingCourseRequest {
  @Onkt.required
  name: string;

  description: string;
  duration_hours: float64;
  is_required: boolean;
}

model CreateTrainingCourseResponse {
  @encode(string)
  id: int64;

  message: string;
}
