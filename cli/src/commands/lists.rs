use clap::{Args, Subcommand};

use crate::api::List;
use crate::commands::{PageArgs, things};
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum ListsCommand {
    /// List your lists and the lists shared with you
    List(ListArgs),
    /// Show a list with its things
    Show {
        /// ID of the list
        id: String,
    },
}

#[derive(Args)]
pub struct ListArgs {
    #[command(flatten)]
    page: PageArgs,

    /// Only lists owned by this user ID
    #[arg(long)]
    owner: Option<String>,

    /// Sort order as field:direction, e.g. created_at:desc or
    /// access_reason:asc
    #[arg(long)]
    order: Option<String>,
}

pub async fn run(command: ListsCommand, session: &Session, output: &Output) -> anyhow::Result<()> {
    match command {
        ListsCommand::List(args) => {
            let lists = session
                .fetch(async |api| {
                    api.list_lists(
                        args.page.page,
                        args.page.per_page,
                        args.owner.as_deref(),
                        args.page.paginate(),
                        args.order.as_deref(),
                    )
                    .await
                })
                .await?;
            output.print(&lists, |lists| {
                let table = table(lists.lists.as_deref().unwrap_or_default());
                let footer = text::page_footer(
                    lists.page,
                    lists.total_page_count,
                    lists.total_count,
                    ("list", "lists"),
                );
                format!("{}\n{footer}", table.render("No lists found."))
            })
        }
        ListsCommand::Show { id } => {
            let list = session.fetch(async |api| api.get_list(&id).await).await?;
            output.print(&list, show)
        }
    }
}

/// A table with one row per list.
pub fn table(lists: &[List]) -> Table {
    let mut table = Table::new(&["ID", "NAME", "THINGS", "OWNER"]);
    for list in lists {
        table.row(vec![
            list.id.clone(),
            list.name.clone(),
            list.things.as_ref().map_or(0, Vec::len).to_string(),
            list.owner.name.clone(),
        ]);
    }
    table
}

fn show(list: &List) -> String {
    let fields = text::fields(&[
        ("id", list.id.clone()),
        ("name", list.name.clone()),
        ("owner", text::user(&list.owner)),
        ("created", text::date(&list.created_at)),
        (
            "sharing",
            list.sharing_state
                .as_ref()
                .map(|state| state.as_str().to_owned())
                .unwrap_or_default(),
        ),
        (
            "shared with",
            list.shares
                .iter()
                .flatten()
                .map(|share| share.target_user.name.as_str())
                .collect::<Vec<_>>()
                .join(", "),
        ),
    ]);
    let things = things::table(list.things.as_deref().unwrap_or_default());
    format!("{fields}\n{}", things.render("No things in this list."))
}
