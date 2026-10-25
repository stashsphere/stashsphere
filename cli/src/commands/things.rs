use clap::{Args, Subcommand};

use crate::api::{Property, Thing};
use crate::commands::PageArgs;
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum ThingsCommand {
    /// List your things and the things shared with you
    List(ListArgs),
    /// Show a thing with its properties, lists and shares
    Show {
        /// ID of the thing
        id: String,
    },
}

#[derive(Args)]
pub struct ListArgs {
    #[command(flatten)]
    page: PageArgs,

    /// Only things owned by this user ID
    #[arg(long)]
    owner: Option<String>,

    /// Only things whose name starts with this term, ignoring case
    #[arg(long)]
    search: Option<String>,

    /// Sort order as field:direction, e.g. created_at:desc or
    /// access_reason:asc
    #[arg(long)]
    order: Option<String>,
}

pub async fn run(command: ThingsCommand, session: &Session, output: &Output) -> anyhow::Result<()> {
    match command {
        ThingsCommand::List(args) => {
            let things = session
                .fetch(async |api| {
                    api.list_things(
                        args.page.page,
                        args.page.per_page,
                        args.owner.as_deref(),
                        args.search.as_deref(),
                        args.page.paginate(),
                        args.order.as_deref(),
                    )
                    .await
                })
                .await?;
            output.print(&things, |things| {
                let table = table(things.things.as_deref().unwrap_or_default());
                let footer = text::page_footer(
                    things.page,
                    things.total_page_count,
                    things.total_count,
                    ("thing", "things"),
                );
                format!("{}\n{footer}", table.render("No things found."))
            })
        }
        ThingsCommand::Show { id } => {
            let thing = session.fetch(async |api| api.get_thing(&id).await).await?;
            output.print(&thing, show)
        }
    }
}

/// A table with one row per thing.
pub fn table(things: &[Thing]) -> Table {
    let mut table = Table::new(&["ID", "NAME", "QUANTITY", "OWNER"]);
    for thing in things {
        table.row(vec![
            thing.id.clone(),
            thing.name.clone(),
            quantity(thing),
            thing.owner.name.clone(),
        ]);
    }
    table
}

fn show(thing: &Thing) -> String {
    let names = |names: Vec<&str>| names.join(", ");
    let mut out = text::fields(&[
        ("id", thing.id.clone()),
        ("name", thing.name.clone()),
        ("owner", text::user(&thing.owner)),
        ("quantity", quantity(thing)),
        ("created", text::date(&thing.created_at)),
        (
            "sharing",
            thing
                .sharing_state
                .as_ref()
                .map(|state| state.as_str().to_owned())
                .unwrap_or_default(),
        ),
        (
            "lists",
            names(
                thing
                    .lists
                    .iter()
                    .flatten()
                    .map(|list| list.name.as_str())
                    .collect(),
            ),
        ),
        (
            "shared with",
            names(
                thing
                    .shares
                    .iter()
                    .flatten()
                    .map(|share| share.target_user.name.as_str())
                    .collect(),
            ),
        ),
        (
            "images",
            names(
                thing
                    .images
                    .iter()
                    .flatten()
                    .map(|image| image.id.as_str())
                    .collect(),
            ),
        ),
        ("description", thing.description.clone()),
        (
            "private note",
            thing.private_note.clone().unwrap_or_default(),
        ),
    ]);
    let properties: Vec<(&str, String)> = thing
        .properties
        .iter()
        .flatten()
        .map(|property| match property {
            Property::PropertyString(p) => (p.name.as_str(), p.value.clone()),
            Property::PropertyFloat(p) => (
                p.name.as_str(),
                format!("{} {}", p.value, p.unit).trim_end().to_owned(),
            ),
            Property::PropertyDatetime(p) => (p.name.as_str(), text::date(&p.value)),
            Property::PropertyBoolean(p) => (p.name.as_str(), p.value.to_string()),
        })
        .collect();
    if !properties.is_empty() {
        out.push_str("\nproperties:\n");
        for line in text::fields(&properties).lines() {
            out.push_str("  ");
            out.push_str(line);
            out.push('\n');
        }
    }
    out
}

fn quantity(thing: &Thing) -> String {
    format!("{} {}", thing.quantity, thing.quantity_unit)
        .trim_end()
        .to_owned()
}
