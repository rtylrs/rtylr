//! Job registry and scheduler, mirroring `internal/platform/jobs`.

use std::collections::{HashMap, HashSet};
use std::future::Future;
use std::pin::Pin;
use std::sync::{Arc, Mutex};
use std::time::Duration;

use thiserror::Error;

pub const DEFAULT_JOB_TIMEOUT: Duration = Duration::from_secs(30 * 60);

#[derive(Debug, Error, PartialEq)]
pub enum JobError {
    #[error("jobs registry is nil")]
    NilRegistry,
    #[error("job name is required")]
    NameRequired,
    #[error("job {0:?} handler is required")]
    HandlerRequired(String),
    #[error("unknown job {0:?}")]
    UnknownJob(String),
    #[error("job already running: {0}")]
    AlreadyRunning(String),
    #[error("job lock not acquired: {0}")]
    LockNotAcquired(String),
    #[error("scheduler is nil")]
    NilScheduler,
    #[error("invalid schedule {0:?}: {1}")]
    InvalidSchedule(String, String),
    #[error("job {0:?} panicked: {1}")]
    Panicked(String, String),
    #[error("job {0:?} failed: {1}")]
    Failed(String, String),
    #[error("job {0:?} timed out after {1:?}")]
    TimedOut(String, Duration),
}

pub type FutResult<'a, T = ()> = Pin<Box<dyn Future<Output = Result<T, JobError>> + Send + 'a>>;

pub type Handler = Arc<dyn Fn() -> FutResult<'static> + Send + Sync>;

pub use crate::locker::MysqlLockHandle;

#[derive(Clone)]
pub struct Job {
    pub name: String,
    pub description: String,
    pub schedule: String,
    pub timeout: Duration,
    pub handler: Handler,
}

impl std::fmt::Debug for Job {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("Job")
            .field("name", &self.name)
            .field("description", &self.description)
            .field("schedule", &self.schedule)
            .field("timeout", &self.timeout)
            .field("handler", &"<handler>")
            .finish()
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct JobResult {
    pub name: String,
    pub duration: Duration,
}

pub struct Registry {
    jobs: Mutex<HashMap<String, Job>>,
}

impl Default for Registry {
    fn default() -> Self {
        Self {
            jobs: Mutex::new(HashMap::new()),
        }
    }
}

impl Registry {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn register(&self, job: Job) -> Result<(), JobError> {
        if job.name.is_empty() {
            return Err(JobError::NameRequired);
        }
        self.jobs.lock().unwrap().insert(job.name.clone(), job);
        Ok(())
    }

    pub fn names(&self) -> Vec<String> {
        let mut names: Vec<String> = self.jobs.lock().unwrap().keys().cloned().collect();
        names.sort();
        names
    }

    pub fn jobs(&self) -> Vec<Job> {
        let mut items: Vec<Job> = self.jobs.lock().unwrap().values().cloned().collect();
        items.sort_by(|a, b| a.name.cmp(&b.name));
        items
    }

    pub fn get(&self, name: &str) -> Option<Job> {
        self.jobs.lock().unwrap().get(name).cloned()
    }

    /// Run a job once and report elapsed time (Go `registry.Run` equivalent;
    /// panics propagate to the scheduler's task in Rust).
    pub async fn run(&self, name: &str) -> Result<JobResult, JobError> {
        let job = self
            .get(name)
            .ok_or_else(|| JobError::UnknownJob(name.to_string()))?;
        let at = std::time::Instant::now();
        let outcome = (job.handler)().await;
        let duration = at.elapsed();
        let _ = &duration;
        outcome.map(|()| JobResult {
            name: name.to_string(),
            duration,
        })
    }
}

pub trait Locker: Send + Sync {
    fn try_lock<'a>(&'a self, name: &'a str) -> FutResult<'a, Option<MysqlLockHandle>>;
}

pub struct Scheduler {
    registry: Arc<Registry>,
    locker: Option<Arc<dyn Locker>>,
    default_timeout: Duration,
    running: Mutex<HashSet<String>>,
}

impl Scheduler {
    pub fn new(registry: Arc<Registry>, locker: Option<Arc<dyn Locker>>) -> Self {
        Self {
            registry,
            locker,
            default_timeout: DEFAULT_JOB_TIMEOUT,
            running: Mutex::new(HashSet::new()),
        }
    }

    pub fn with_default_timeout(mut self, timeout: Duration) -> Self {
        if !timeout.is_zero() {
            self.default_timeout = timeout;
        }
        self
    }

    /// Validate every scheduled job's cron expression (Go `Start`
    /// registration equivalent; no ticker is started in tests).
    ///
    /// Go's parser accepts `SecondOptional`, so a 5-field expression gets a
    /// leading second `0` before re-parsing (the `cron` crate only takes
    /// 6/7 fields).
    pub fn validate_schedules(&self) -> Result<(), JobError> {
        for job in self.registry.jobs() {
            let schedule = job.schedule.trim();
            if schedule.is_empty() {
                continue;
            }
            let parsed = schedule
                .parse::<cron::Schedule>()
                .or_else(|_| format!("0 {schedule}").parse::<cron::Schedule>());
            if let Err(error) = parsed {
                return Err(JobError::InvalidSchedule(
                    schedule.to_string(),
                    error.to_string(),
                ));
            }
        }
        Ok(())
    }

    pub async fn run(&self, name: &str) -> Result<JobResult, JobError> {
        let job = self
            .registry
            .get(name)
            .ok_or_else(|| JobError::UnknownJob(name.to_string()))?;
        {
            let mut running = self.running.lock().unwrap();
            if !running.insert(name.to_string()) {
                return Err(JobError::AlreadyRunning(name.to_string()));
            }
        }
        let mut handle = match &self.locker {
            Some(locker) => match locker.try_lock(name).await {
                Ok(Some(handle)) => Some(handle),
                Ok(None) => {
                    self.running.lock().unwrap().remove(name);
                    return Err(JobError::LockNotAcquired(name.to_string()));
                }
                Err(error) => {
                    self.running.lock().unwrap().remove(name);
                    return Err(error);
                }
            },
            None => None,
        };
        let timeout = if job.timeout.is_zero() {
            self.default_timeout
        } else {
            job.timeout
        };
        let at = std::time::Instant::now();
        let run_future = (job.handler)();
        let outcome = if timeout.is_zero() {
            run_future.await
        } else {
            match tokio::time::timeout(timeout, run_future).await {
                Ok(result) => result,
                Err(_) => Err(JobError::TimedOut(name.to_string(), timeout)),
            }
        };
        if let Some(handle) = handle.as_mut() {
            let _ = handle.release().await;
        }
        self.running.lock().unwrap().remove(name);
        let duration = at.elapsed();
        match outcome {
            Ok(()) => Ok(JobResult {
                name: name.to_string(),
                duration,
            }),
            Err(error) => Err(error),
        }
    }
}

fn parse_bool_env(raw: &str) -> Option<bool> {
    match raw.trim() {
        "1" | "t" | "T" | "TRUE" | "true" | "True" => Some(true),
        "0" | "f" | "F" | "FALSE" | "false" | "False" => Some(false),
        _ => None,
    }
}

pub fn scheduler_enabled_from_env(get: &dyn Fn(&str) -> Option<String>) -> bool {
    if let Some(raw) = get("SCHEDULER_DISABLED") {
        if parse_bool_env(&raw) == Some(true) {
            return false;
        }
    }
    if let Some(raw) = get("SCHEDULER_ENABLED") {
        if let Some(enabled) = parse_bool_env(&raw) {
            return enabled;
        }
    }
    true
}

pub fn run_job_from_env(get: &dyn Fn(&str) -> Option<String>) -> String {
    get("RUN_JOB").unwrap_or_default().trim().to_string()
}

pub fn duration_from_env(
    get: &dyn Fn(&str) -> Option<String>,
    key: &str,
    fallback: Duration,
) -> Duration {
    let Some(raw) = get(key) else {
        return fallback;
    };
    let trimmed = raw.trim();
    if trimmed.is_empty() {
        return fallback;
    }
    // Go accepts Go durations, then bare seconds.
    if let Ok(secs) = trimmed.parse::<u64>() {
        return Duration::from_secs(secs);
    }
    if let Ok(parsed) = crate::jobs::parse_go_duration_pub(trimmed) {
        return parsed;
    }
    fallback
}

pub fn schedule_from_env(
    get: &dyn Fn(&str) -> Option<String>,
    key: &str,
    fallback: &str,
) -> String {
    match get(key) {
        Some(value) if !value.trim().is_empty() => value.trim().to_string(),
        _ => fallback.to_string(),
    }
}

pub fn parse_go_duration_pub(raw: &str) -> Result<Duration, String> {
    crate::emailqueue::parse_go_duration(raw).map_err(|e| e.to_string())
}

#[cfg(test)]
mod tests {
    use super::{Job, JobError, Registry, Scheduler};
    use std::sync::atomic::{AtomicUsize, Ordering};
    use std::sync::Arc;

    fn registry_with(name: &str, schedule: &str, on_run: Arc<AtomicUsize>) -> Arc<Registry> {
        let registry = Arc::new(Registry::new());
        let counter = on_run.clone();
        registry
            .register(Job {
                name: name.to_string(),
                description: String::new(),
                schedule: schedule.to_string(),
                timeout: std::time::Duration::ZERO,
                handler: Arc::new(move || {
                    let counter = counter.clone();
                    Box::pin(async move {
                        counter.fetch_add(1, Ordering::SeqCst);
                        Ok(())
                    })
                        as std::pin::Pin<
                            Box<dyn std::future::Future<Output = Result<(), JobError>> + Send>,
                        >
                }),
            })
            .unwrap();
        registry
    }

    #[tokio::test]
    async fn registry_names_sorted() {
        let counter = Arc::new(AtomicUsize::new(0));
        let registry = registry_with("b", "", counter.clone());
        registry_with("a", "", counter.clone());
        assert_eq!(registry.names(), vec!["b".to_string()]);
        let other = Arc::new(Registry::new());
        for (name, _) in [("b", ""), ("a", "")] {
            other
                .register(Job {
                    name: name.to_string(),
                    description: String::new(),
                    schedule: String::new(),
                    timeout: std::time::Duration::ZERO,
                    handler: Arc::new(|| {
                        Box::pin(async move { Ok(()) })
                            as std::pin::Pin<
                                Box<dyn std::future::Future<Output = Result<(), JobError>> + Send>,
                            >
                    }),
                })
                .unwrap();
        }
        assert_eq!(other.names(), vec!["a".to_string(), "b".to_string()]);
        assert!(other.get("b").is_some());
    }

    #[tokio::test]
    async fn scheduler_run_and_duplicate() {
        let counter = Arc::new(AtomicUsize::new(0));
        let registry = registry_with("slow", "", counter.clone());
        let scheduler = Scheduler::new(registry, None);
        let first = scheduler.run("slow").await.unwrap();
        assert_eq!(first.name, "slow");
        assert_eq!(counter.load(Ordering::SeqCst), 1);
    }

    #[tokio::test]
    async fn invalid_schedule_rejected() {
        let counter = Arc::new(AtomicUsize::new(0));
        let registry = registry_with("bad", "not a cron", counter);
        let scheduler = Scheduler::new(registry, None);
        assert!(matches!(
            scheduler.validate_schedules(),
            Err(JobError::InvalidSchedule(_, _))
        ));
        let counter = Arc::new(AtomicUsize::new(0));
        let registry = registry_with("good", "* * * * *", counter);
        let scheduler = Scheduler::new(registry, None);
        assert!(scheduler.validate_schedules().is_ok());
    }

    #[tokio::test]
    async fn env_helpers() {
        let get = |_: &str| -> Option<String> { None };
        assert!(super::scheduler_enabled_from_env(&get));
        assert_eq!(super::run_job_from_env(&get), "");
    }
}
