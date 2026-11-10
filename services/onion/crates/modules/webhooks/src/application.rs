//! Application layer: ports (traits), registration validation, delivery
//! signing, and the sender/notifier state machines. Mirrors
//! `internal/application`.

pub mod delivery_repository;
pub mod event_payload;
pub mod notifier;
pub mod registration;
pub mod sender;
pub mod transport;

pub use delivery_repository::{
    AppError, DeliveryRepository, EventReader, NotificationEmail, NotificationQueue,
    NotificationRepository,
};
pub use notifier::Notifier;
pub use registration::validate_registration;
pub use sender::Sender;
pub use transport::validate_destination;
