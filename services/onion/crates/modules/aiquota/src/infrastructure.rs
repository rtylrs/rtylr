//! Sqlx-backed repository over the generated queries, mirroring
//! `internal/infrastructure/repository.go`. The pool comes from the
//! `db` crate's shared `Database` pool (the tests point it at a
//! throwaway MySQL via the local-env stack).

pub mod repository;
