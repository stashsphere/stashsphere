use std::process::ExitCode;

use serde::Serialize;
use serde_json::json;

use crate::error::{ApiFailure, exit_code};

/// Prints results to stdout and errors to stderr, as text or JSON.
pub struct Output {
    json: bool,
}

impl Output {
    pub fn new(json: bool) -> Self {
        Self { json }
    }

    /// Prints `value` as JSON, or the text produced by `human`.
    pub fn print<T: Serialize>(
        &self,
        value: &T,
        human: impl FnOnce(&T) -> String,
    ) -> anyhow::Result<()> {
        if self.json {
            println!("{}", serde_json::to_string_pretty(value)?);
        } else {
            println!("{}", human(value));
        }
        Ok(())
    }

    /// Prints a confirmation for commands without a result.
    pub fn done(&self, message: &str) -> anyhow::Result<()> {
        self.print(&json!({ "ok": true }), |_| message.to_owned())
    }

    pub fn error(&self, err: &anyhow::Error) -> ExitCode {
        if self.json {
            let status = err
                .downcast_ref::<ApiFailure>()
                .map(|failure| failure.status);
            let body = json!({ "error": { "message": format!("{err:#}"), "status": status } });
            eprintln!("{body}");
        } else {
            eprintln!("error: {err:#}");
        }
        exit_code(err)
    }
}
