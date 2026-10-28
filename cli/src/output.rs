use std::io::IsTerminal;
use std::process::ExitCode;

use clap::ValueEnum;
use serde::Serialize;
use serde_json::json;

use crate::error::{ApiFailure, exit_code};

#[derive(Clone, Copy, Default, ValueEnum)]
pub enum Format {
    /// Tables and key/value blocks for humans
    #[default]
    Text,
    /// The API response as JSON, pretty-printed on a terminal and compact
    /// otherwise
    Json,
}

/// Prints results to stdout and errors to stderr, as text or JSON.
pub struct Output {
    format: Format,
}

impl Output {
    pub fn new(format: Format) -> Self {
        Self { format }
    }

    /// Whether results are printed as text, for commands that fetch more
    /// to describe a result to humans.
    pub fn is_text(&self) -> bool {
        matches!(self.format, Format::Text)
    }

    /// Prints `value` as JSON, or the text produced by `text`.
    pub fn print<T: Serialize>(
        &self,
        value: &T,
        text: impl FnOnce(&T) -> String,
    ) -> anyhow::Result<()> {
        match self.format {
            Format::Json => println!("{}", to_json(value)?),
            Format::Text => println!("{}", text(value).trim_end()),
        }
        Ok(())
    }

    /// Prints a confirmation for commands without a result.
    pub fn done(&self, message: &str) -> anyhow::Result<()> {
        self.print(&json!({ "ok": true }), |_| message.to_owned())
    }

    pub fn error(&self, err: &anyhow::Error) -> ExitCode {
        match self.format {
            Format::Json => {
                let status = err
                    .downcast_ref::<ApiFailure>()
                    .map(|failure| failure.status);
                let body = json!({ "error": { "message": format!("{err:#}"), "status": status } });
                eprintln!("{body}");
            }
            Format::Text => eprintln!("error: {err:#}"),
        }
        exit_code(err)
    }
}

fn to_json<T: Serialize>(value: &T) -> serde_json::Result<String> {
    if std::io::stdout().is_terminal() {
        serde_json::to_string_pretty(value)
    } else {
        serde_json::to_string(value)
    }
}
