use std::fs::File;
use std::io::{IsTerminal, Write};
use std::path::{Path, PathBuf};

use anyhow::{Context, bail};
use clap::{Args, Subcommand, ValueEnum};
use futures_util::StreamExt;

use crate::api::{Image, ImageModifyParams, MultipartFile, ReducedImage, UploadImageRequest};
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum ImagesCommand {
    /// List the images you uploaded
    List(ListArgs),
    /// Upload image files, to add them to things with `things create
    /// --image` or `things update --add-image`
    Upload {
        /// Image files, detected as images by their content
        #[arg(required = true)]
        files: Vec<PathBuf>,
    },
    /// Download an image by its hash, shown by `images list` and `things
    /// show`
    Download(DownloadArgs),
    /// Rotate an image you uploaded
    Rotate {
        /// ID of the image
        #[arg(allow_hyphen_values = true)]
        id: String,

        /// left and right turn by 90 degrees, around by 180
        #[arg(value_enum)]
        direction: Direction,
    },
    /// Delete an image you uploaded that no thing or profile uses
    Delete {
        /// ID of the image
        #[arg(allow_hyphen_values = true)]
        id: String,
    },
}

#[derive(Args)]
pub struct ListArgs {
    /// Page to show, starting at 0
    #[arg(long)]
    page: Option<i64>,

    /// Number of images per page
    #[arg(long)]
    per_page: Option<i64>,

    /// Only images that no thing uses
    #[arg(long)]
    unassigned: bool,
}

#[derive(Args)]
pub struct DownloadArgs {
    /// Hash of the image
    hash: String,

    /// File to write the image to. Without it the image goes to stdout,
    /// which must not be a terminal.
    #[arg(long, short = 'o', value_name = "PATH")]
    out: Option<PathBuf>,

    /// Scale JPEG and PNG images down to this width in pixels
    #[arg(long, value_parser = clap::value_parser!(i64).range(20..=8192))]
    width: Option<i64>,
}

#[derive(Clone, Copy, ValueEnum)]
pub enum Direction {
    Left,
    Right,
    Around,
}

impl Direction {
    /// The rotation of the API, counter-clockwise in degrees.
    fn rotation(self) -> i64 {
        match self {
            Direction::Left => 90,
            Direction::Around => 180,
            Direction::Right => 270,
        }
    }
}

pub async fn run(command: ImagesCommand, session: &Session, output: &Output) -> anyhow::Result<()> {
    match command {
        ImagesCommand::List(args) => {
            let images = session
                .fetch(async |api| {
                    api.list_images(args.page, args.per_page, args.unassigned.then_some(true))
                        .await
                })
                .await?;
            output.print(&images, |images| {
                let table = table(images.images.as_deref().unwrap_or_default());
                let footer = text::page_footer(
                    images.page,
                    images.total_page_count,
                    images.total_count,
                    ("image", "images"),
                );
                format!("{}\n{footer}", table.render("No images found."))
            })
        }
        ImagesCommand::Upload { files } => {
            let mut images = Vec::new();
            for path in &files {
                let request = UploadImageRequest {
                    file: read_file(path)?,
                };
                let image = session
                    .fetch(async |api| api.upload_image(request.clone()).await)
                    .await
                    .with_context(|| format!("uploading {}", path.display()))?;
                images.push(image);
            }
            output.print(&images, |images| {
                images
                    .iter()
                    .map(|image| format!("Uploaded image {} ({})", image.name, image.id))
                    .collect::<Vec<_>>()
                    .join("\n")
            })
        }
        ImagesCommand::Download(args) => download(args, session, output).await,
        ImagesCommand::Rotate { id, direction } => {
            let params = ImageModifyParams {
                rotation: direction.rotation(),
            };
            let image = session
                .fetch(async |api| api.modify_image(&id, params.clone()).await)
                .await?;
            output.print(&image, |image| {
                format!("Rotated image {} ({})", image.name, image.id)
            })
        }
        ImagesCommand::Delete { id } => {
            let image = session
                .fetch(async |api| api.delete_image(&id).await)
                .await?;
            output.print(&image, |image| {
                format!("Deleted image {} ({})", image.name, image.id)
            })
        }
    }
}

fn read_file(path: &Path) -> anyhow::Result<MultipartFile> {
    let content = std::fs::read(path).with_context(|| format!("reading {}", path.display()))?;
    let file = MultipartFile::new(content);
    Ok(match path.file_name() {
        Some(name) => file.with_file_name(name.to_string_lossy()),
        None => file,
    })
}

/// Streams the image to a file or stdout, as it may be larger than the
/// client's limit for buffered responses.
async fn download(args: DownloadArgs, session: &Session, output: &Output) -> anyhow::Result<()> {
    if args.out.is_none() && std::io::stdout().is_terminal() {
        bail!("not writing an image to the terminal, use --out or redirect stdout");
    }
    let mut stream = session
        .fetch(async |api| {
            api.get_asset_binary_stream(
                args.hash.clone(),
                args.width,
                None::<String>,
                None::<String>,
            )
            .await
        })
        .await?;
    // opened once the server has the image, to leave no empty file behind
    let mut sink: Box<dyn Write> = match &args.out {
        Some(path) => {
            Box::new(File::create(path).with_context(|| format!("writing {}", path.display()))?)
        }
        None => Box::new(std::io::stdout().lock()),
    };
    while let Some(chunk) = stream.next().await {
        sink.write_all(&chunk.context("downloading the image")?)?;
    }
    sink.flush()?;
    match &args.out {
        Some(path) => output.done(&format!("Saved image to {}", path.display())),
        None => Ok(()),
    }
}

/// A table with one row per image.
fn table(images: &[Image]) -> Table {
    let mut table = Table::new(&["ID", "NAME", "SIZE", "THINGS", "HASH"]);
    for image in images {
        table.row(vec![
            image.id.clone(),
            image.name.clone(),
            size(image.width, image.height),
            image
                .things
                .iter()
                .flatten()
                .map(|thing| thing.name.as_str())
                .collect::<Vec<_>>()
                .join(", "),
            image.hash.clone(),
        ]);
    }
    table
}

/// A line describing an image of a thing.
pub fn describe(image: &ReducedImage) -> String {
    let size = size(image.width, image.height);
    let name = [image.name.as_str(), size.as_str()]
        .into_iter()
        .filter(|part| !part.is_empty())
        .collect::<Vec<_>>()
        .join(" ");
    format!("{} {name} (hash {})", image.id, image.hash)
}

fn size(width: Option<i64>, height: Option<i64>) -> String {
    match (width, height) {
        (Some(width), Some(height)) => format!("{width}x{height}"),
        _ => String::new(),
    }
}
