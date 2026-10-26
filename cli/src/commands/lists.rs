use clap::{Args, Subcommand};

use crate::api::{List, NewListParams, UpdateListParams};
use crate::commands::{PageArgs, Sharing, new_sharing_state, things};
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
    /// Create a list
    Create(CreateArgs),
    /// Rename a list you own, change its sharing or its things
    Update(UpdateArgs),
    /// Delete a list you own, keeping its things
    Delete {
        /// ID of the list
        id: String,
    },
}

#[derive(Args)]
pub struct CreateArgs {
    /// Name of the list
    #[arg(long)]
    name: String,

    /// Who can see the list. Defaults to the sharing state of your profile.
    #[arg(long, value_enum)]
    sharing: Option<Sharing>,

    /// ID of a thing you own to put in the list, can be repeated
    #[arg(long = "thing", value_name = "ID")]
    things: Vec<String>,
}

#[derive(Args)]
pub struct UpdateArgs {
    /// ID of the list
    id: String,

    /// New name
    #[arg(long)]
    name: Option<String>,

    /// Who can see the list
    #[arg(long, value_enum)]
    sharing: Option<Sharing>,

    /// ID of a thing you own to add to the list, can be repeated
    #[arg(long, value_name = "ID")]
    add_thing: Vec<String>,

    /// ID of a thing to remove from the list, can be repeated
    #[arg(long, value_name = "ID")]
    remove_thing: Vec<String>,
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
        ListsCommand::Create(args) => {
            let params = NewListParams {
                name: args.name,
                sharing_state: new_sharing_state(args.sharing, session).await?,
                thing_ids: args.things,
            };
            let list = session
                .fetch(async |api| api.create_list(params.clone()).await)
                .await?;
            output.print(&list, |list| {
                format!("Created list {} ({})", list.name, list.id)
            })
        }
        ListsCommand::Update(args) => {
            // the API replaces the whole list, so the unchanged fields are
            // sent as they are
            let list = session
                .fetch(async |api| api.get_list(&args.id).await)
                .await?;
            let mut thing_ids: Vec<String> = list
                .things
                .into_iter()
                .flatten()
                .map(|thing| thing.id)
                .filter(|id| !args.remove_thing.contains(id))
                .collect();
            for id in args.add_thing {
                if !thing_ids.contains(&id) {
                    thing_ids.push(id);
                }
            }
            let params = UpdateListParams {
                name: args.name.unwrap_or(list.name),
                sharing_state: match args.sharing {
                    Some(sharing) => sharing.into(),
                    None => list.sharing_state.unwrap_or_default(),
                },
                thing_ids,
            };
            let list = session
                .fetch(async |api| api.update_list(&args.id, params.clone()).await)
                .await?;
            output.print(&list, show)
        }
        ListsCommand::Delete { id } => {
            session
                .fetch(async |api| api.delete_list(&id).await)
                .await?;
            output.done(&format!("Deleted list {id}"))
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
