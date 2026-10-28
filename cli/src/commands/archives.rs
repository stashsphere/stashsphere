//! Exports of your whole collection as a ZIP archive, and imports of such
//! archives. The server builds and reads the archives in the background,
//! so starting one returns right away; `--wait` polls until it is done.

use std::fs::File;
use std::io::{IsTerminal, Write};
use std::path::PathBuf;
use std::time::Duration;

use anyhow::{Context, bail};
use clap::{Args, Subcommand};
use futures_util::StreamExt;

use crate::api::{ExportStatus, ImportStatus, StartImportRequest};
use crate::commands::images;
use crate::output::Output;
use crate::session::Session;
use crate::text;

/// How often `--wait` asks for the status. The server picks up new jobs
/// about once a minute.
const POLL_INTERVAL: Duration = Duration::from_secs(2);

#[derive(Subcommand)]
pub enum ExportCommand {
    /// Start an export of all your things, lists and images
    Start(WaitArgs),
    /// Show the status of your most recent export
    Status(WaitArgs),
    /// Download your most recent finished export
    Download(DownloadArgs),
}

#[derive(Args)]
pub struct WaitArgs {
    /// Wait until the job is finished, and fail if it failed
    #[arg(long)]
    wait: bool,
}

#[derive(Args)]
pub struct DownloadArgs {
    /// File to write the archive to, "-" for stdout. Defaults to
    /// stashsphere-export-DATE.zip in the current directory.
    #[arg(long, short = 'o', value_name = "PATH")]
    out: Option<PathBuf>,
}

#[derive(Subcommand)]
pub enum ImportCommand {
    /// Import an archive created by an export into your collection
    Start {
        /// ZIP archive created by an export
        file: PathBuf,

        #[command(flatten)]
        wait: WaitArgs,
    },
    /// Show the status of your most recent import
    Status(WaitArgs),
}

pub async fn export(
    command: ExportCommand,
    session: &Session,
    output: &Output,
) -> anyhow::Result<()> {
    let export = match command {
        ExportCommand::Start(args) => {
            let export = session.fetch(async |api| api.start_export().await).await?;
            wait_export(export, args, session).await?
        }
        ExportCommand::Status(args) => {
            let export = session
                .fetch(async |api| api.get_export_status().await)
                .await?;
            wait_export(export, args, session).await?
        }
        ExportCommand::Download(args) => return download(args, session, output).await,
    };
    output.print(&export, show_export)
}

pub async fn import(
    command: ImportCommand,
    session: &Session,
    output: &Output,
) -> anyhow::Result<()> {
    let import = match command {
        ImportCommand::Start { file, wait } => {
            let request = StartImportRequest {
                file: images::read_file(&file)?,
            };
            let import = session
                .fetch(async |api| api.start_import(request.clone()).await)
                .await
                .with_context(|| format!("uploading {}", file.display()))?;
            wait_import(import, wait, session).await?
        }
        ImportCommand::Status(args) => {
            let import = session
                .fetch(async |api| api.get_import_status().await)
                .await?;
            wait_import(import, args, session).await?
        }
    };
    output.print(&import, show_import)
}

async fn wait_export(
    mut export: ExportStatus,
    args: WaitArgs,
    session: &Session,
) -> anyhow::Result<ExportStatus> {
    if !args.wait {
        return Ok(export);
    }
    while !finished(&export.status) {
        tokio::time::sleep(POLL_INTERVAL).await;
        export = session
            .fetch(async |api| api.get_export_status().await)
            .await?;
    }
    if let Some(error) = failed(&export.status, &export.error) {
        bail!("export {} failed: {error}", export.id);
    }
    Ok(export)
}

async fn wait_import(
    mut import: ImportStatus,
    args: WaitArgs,
    session: &Session,
) -> anyhow::Result<ImportStatus> {
    if !args.wait {
        return Ok(import);
    }
    while !finished(&import.status) {
        tokio::time::sleep(POLL_INTERVAL).await;
        import = session
            .fetch(async |api| api.get_import_status().await)
            .await?;
    }
    if let Some(error) = failed(&import.status, &import.error) {
        bail!("import {} failed: {error}", import.id);
    }
    Ok(import)
}

/// Streams the archive to a file or stdout, as it may be larger than the
/// client's limit for buffered responses.
async fn download(args: DownloadArgs, session: &Session, output: &Output) -> anyhow::Result<()> {
    let to_stdout = args.out.as_deref() == Some("-".as_ref());
    if to_stdout && std::io::stdout().is_terminal() {
        bail!("not writing an archive to the terminal, use --out or redirect stdout");
    }
    let path = match args.out {
        Some(path) => path,
        // named like the web app's download, after the export's creation
        None => {
            let export = session
                .fetch(async |api| api.get_export_status().await)
                .await?;
            PathBuf::from(format!(
                "stashsphere-export-{}.zip",
                export.created_at.format("%Y-%m-%d-%H%M")
            ))
        }
    };
    let mut stream = session
        .fetch(async |api| api.download_export_binary_stream().await)
        .await?;
    // opened once the server has the archive, to leave no empty file behind
    let mut sink: Box<dyn Write> = if to_stdout {
        Box::new(std::io::stdout().lock())
    } else {
        Box::new(File::create(&path).with_context(|| format!("writing {}", path.display()))?)
    };
    while let Some(chunk) = stream.next().await {
        sink.write_all(&chunk.context("downloading the export")?)?;
    }
    sink.flush()?;
    if to_stdout {
        return Ok(());
    }
    output.done(&format!("Saved export to {}", path.display()))
}

fn finished(status: &str) -> bool {
    matches!(status, "done" | "error")
}

/// The error of a failed export or import.
fn failed<'a>(status: &str, error: &'a Option<String>) -> Option<&'a str> {
    (status == "error").then(|| error.as_deref().unwrap_or("unknown error"))
}

fn show_export(export: &ExportStatus) -> String {
    text::fields(&[
        ("id", export.id.clone()),
        ("status", export.status.clone()),
        ("created", text::date(&export.created_at)),
        (
            "expires",
            export
                .expires_at
                .as_ref()
                .map(text::date)
                .unwrap_or_default(),
        ),
        ("error", export.error.clone().unwrap_or_default()),
    ])
}

fn show_import(import: &ImportStatus) -> String {
    let count = |count: Option<i64>| count.map(|n| n.to_string()).unwrap_or_default();
    text::fields(&[
        ("id", import.id.clone()),
        ("status", import.status.clone()),
        ("created", text::date(&import.created_at)),
        (
            "completed",
            import
                .completed_at
                .as_ref()
                .map(text::date)
                .unwrap_or_default(),
        ),
        ("things", count(import.things_imported)),
        ("lists", count(import.lists_imported)),
        ("images", count(import.images_imported)),
        ("error", import.error.clone().unwrap_or_default()),
    ])
}
