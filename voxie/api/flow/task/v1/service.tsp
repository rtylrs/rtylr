import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.task.v1;

@route("/flow/task/v1")
interface TaskService {
  @get
  @route("/tasks")
  listTasks(...ListTasksRequest): ListTasksResponse;

  @get
  @route("/tasks/my")
  getMyTasks(...GetMyTasksRequest): GetMyTasksResponse;

  @get
  @route("/tasks/my-project")
  getMyTasksProject(...GetMyTasksProjectRequest): GetMyTasksProjectResponse;

  @get
  @route("/tasks/search")
  searchTasks(...SearchTasksRequest): SearchTasksResponse;

  @post
  @route("/tasks")
  createTask(...CreateTaskRequest): CreateTaskResponse;

  @get
  @route("/tasks/{id}")
  getTaskDetail(...GetTaskDetailRequest): GetTaskDetailResponse;

  @put
  @route("/tasks/{id}")
  updateTask(...UpdateTaskRequest): UpdateTaskResponse;

  @patch
  @route("/tasks/{id}/status")
  updateTaskStatus(...UpdateTaskStatusRequest): UpdateTaskStatusResponse;

  @delete
  @route("/tasks/{id}")
  deleteTask(...DeleteTaskRequest): DeleteTaskResponse;

  @post
  @route("/tasks/subtasks")
  createSubtask(...CreateSubtaskRequest): CreateSubtaskResponse;

  @get
  @route("/tasks/{task_id}/subtasks")
  getSubtasks(...GetSubtasksRequest): GetSubtasksResponse;

  @post
  @route("/tasks/{task_id}/comments")
  addComment(...AddCommentRequest): AddCommentResponse;

  @delete
  @route("/tasks/comments/{id}")
  deleteComment(...DeleteCommentRequest): DeleteCommentResponse;

  @get
  @route("/tasks/{task_id}/activity")
  getTaskActivity(...GetTaskActivityRequest): GetTaskActivityResponse;

  @post
  @route("/tasks/attachments")
  createAttachment(...CreateAttachmentRequest): CreateAttachmentResponse;

  @get
  @route("/tasks/{task_id}/attachments")
  getAttachments(...GetAttachmentsRequest): GetAttachmentsResponse;

  @delete
  @route("/tasks/attachments/{id}")
  deleteAttachment(...DeleteAttachmentRequest): DeleteAttachmentResponse;

  @post
  @route("/tasks/{task_id}/time-logs")
  createTimeLog(...CreateTimeLogRequest): CreateTimeLogResponse;

  @get
  @route("/tasks/{task_id}/time-logs")
  getTimeLogs(...GetTimeLogsRequest): GetTimeLogsResponse;

  @delete
  @route("/tasks/time-logs/{id}")
  deleteTimeLog(...DeleteTimeLogRequest): DeleteTimeLogResponse;

  @post
  @route("/tasks/{task_id}/dependencies")
  addTaskDependency(...AddTaskDependencyRequest): AddTaskDependencyResponse;

  @delete
  @route("/tasks/{task_id}/dependencies/{depends_on_task_id}")
  removeTaskDependency(
    ...RemoveTaskDependencyRequest,
  ): RemoveTaskDependencyResponse;
}
