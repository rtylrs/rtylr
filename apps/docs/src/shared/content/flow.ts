export const flow = {
  title: 'Flow (Project Management)',
  body: `
<p>The rtylr Flow app is a project management tool built for teams that want to plan, track, and ship work. It combines the best of issue trackers and project boards in a fast, keyboard-driven interface.</p>

<h2 id="workspaces">Workspaces</h2>
<p>Workspaces are top-level containers that separate different teams or departments. Each workspace has its own projects, members, and settings. An organization can have multiple workspaces (e.g., Engineering, Marketing, Operations).</p>

<h2 id="projects">Projects</h2>
<p>Projects organize related work. Each project has three views:</p>
<ul>
  <li><strong>List</strong> - A flat list of all tasks with sortable columns</li>
  <li><strong>Board</strong> - A kanban board with customizable columns (e.g., Backlog, In Progress, Done)</li>
  <li><strong>Overview</strong> - Project description, progress metrics, and recent activity</li>
</ul>

<h2 id="tasks">Tasks</h2>
<p>Tasks are the core unit of work in Flow. Each task includes:</p>
<ul>
  <li>Title and description (Markdown supported)</li>
  <li>Assignee and reviewer</li>
  <li>Status, priority (Urgent, High, Medium, Low, None), and labels</li>
  <li>Due date and estimated effort</li>
  <li>Sub-tasks for breaking down work</li>
  <li>Comments and activity feed</li>
  <li>Attachments</li>
</ul>

<h2 id="my-issues">My Issues</h2>
<p>A personal view showing all tasks assigned to you across all projects and workspaces. Filter by status, priority, or due date. This is the default landing page in Flow.</p>

<h2 id="cycles">Cycles</h2>
<p>Cycles are time-boxed iterations (similar to sprints). Define a cycle with a start and end date, add tasks, and track progress. At the end of a cycle, incomplete tasks can be moved to the next cycle automatically.</p>

<h2 id="milestones">Milestones</h2>
<p>Milestones mark significant points in a project (e.g., v1.0 launch, beta release). Tasks can be grouped under milestones to track progress toward major goals. Milestones show completion percentage based on associated task status.</p>

<h2 id="roadmap">Roadmap</h2>
<p>The roadmap provides a high-level view of projects and milestones across time. It displays a horizontal timeline with bars representing project phases and milestone markers. Use it for planning and communicating the big picture to stakeholders.</p>

<h2 id="timeline">Timeline</h2>
<p>The timeline (Gantt) view shows tasks on a calendar with start and end dates. Drag to adjust dates, link tasks with dependencies, and visualize the critical path. Timeline is available at the project level.</p>

<h2 id="views">Views</h2>
<p>Create saved views with custom filters, grouping, and sorting. Views can be shared with the team or kept private. Examples: "My overdue tasks," "High priority bugs," "Tasks due this week."</p>

<h2 id="team-workload">Team Workload</h2>
<p>The workload view shows task distribution across team members for a given period. Bars indicate capacity utilization based on estimated effort. Use it to balance work and avoid overloading team members.</p>

<h2 id="triage">Triage</h2>
<p>The triage inbox collects new tasks that have not been assigned to a project or cycle. Team leads review triage items and assign them to the appropriate project, set priority, and assign owners.</p>

<h2 id="keyboard-shortcuts">Keyboard Shortcuts</h2>
<p>Flow is designed for keyboard-first navigation. Key shortcuts include:</p>
<ul>
  <li><code>C</code> - Create new task</li>
  <li><code>J / K</code> - Navigate up/down in lists</li>
  <li><code>Enter</code> - Open selected task</li>
  <li><code>A</code> - Assign task</li>
  <li><code>P</code> - Set priority</li>
  <li><code>L</code> - Add label</li>
  <li><code>D</code> - Set due date</li>
  <li><code>Cmd+K</code> - Command palette</li>
</ul>
  `.trim(),
};
