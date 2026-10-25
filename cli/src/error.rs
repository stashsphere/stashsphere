use std::fmt::Debug;
use std::process::ExitCode;

use serde::{Deserialize, Serialize};

use crate::api::ApiOpError;

/// A failed API call, reduced to what a user or agent can act on.
#[derive(Debug, Serialize, thiserror::Error)]
#[error("{message} (HTTP {status})")]
pub struct ApiFailure {
    pub status: u16,
    pub message: String,
}

/// The `message` of an error body. Errors raised by echo itself, e.g. for a
/// failed login, lack the `code` that the documented ErrorResponse requires.
#[derive(Deserialize)]
struct ErrorMessage {
    message: String,
}

#[derive(Debug, thiserror::Error)]
#[error("not logged in, run `stashctl login` first")]
pub struct NotLoggedIn;

/// Converts the generated client's error into an [`anyhow::Error`] that
/// carries an [`ApiFailure`] when the server responded.
pub fn api_error<E: Debug + Send + Sync + 'static>(err: ApiOpError<E>) -> anyhow::Error {
    match err {
        ApiOpError::Transport(err) => err.into(),
        ApiOpError::Api(err) => {
            let message = serde_json::from_slice::<ErrorMessage>(&err.raw_body)
                .map(|body| body.message)
                .unwrap_or_else(|_| err.body.trim().to_owned());
            ApiFailure {
                status: err.status,
                message,
            }
            .into()
        }
    }
}

/// Exit codes, so scripts and agents can react without parsing messages.
pub fn exit_code(err: &anyhow::Error) -> ExitCode {
    if err.is::<NotLoggedIn>() {
        return ExitCode::from(3);
    }
    match err.downcast_ref::<ApiFailure>() {
        Some(ApiFailure { status: 401, .. }) => ExitCode::from(3),
        Some(ApiFailure { status: 403, .. }) => ExitCode::from(4),
        Some(ApiFailure { status: 404, .. }) => ExitCode::from(5),
        _ => ExitCode::FAILURE,
    }
}
