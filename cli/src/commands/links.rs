use clap::Subcommand;
use serde::Serialize;

use crate::api::{NewPublicShareParams, PublicShare, PublicThing};
use crate::commands::{images, things};
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum LinksCommand {
    /// List your public links
    List,
    /// Create a public link to a thing or list you own, viewable without an account
    Create {
        /// ID of the thing or list
        #[arg(allow_hyphen_values = true)]
        object_id: String,
    },
    /// Show what a public link shows, as someone without an account sees it
    Show {
        /// Token of the public link
        #[arg(allow_hyphen_values = true)]
        token: String,
    },
    /// Revoke a public link
    Delete {
        /// Token of the public link
        #[arg(allow_hyphen_values = true)]
        token: String,
    },
}

/// An API response together with the URL of its public link.
#[derive(Serialize)]
struct WithUrl<T> {
    #[serde(flatten)]
    inner: T,
    url: String,
}

pub async fn run(command: LinksCommand, session: &Session, output: &Output) -> anyhow::Result<()> {
    match command {
        LinksCommand::List => {
            let links = session
                .fetch(async |api| api.list_public_shares().await)
                .await?;
            let links: Vec<_> = links
                .into_iter()
                .map(|link| WithUrl {
                    url: url(session, &link.id),
                    inner: link,
                })
                .collect();
            output.print(&links, |links| {
                let mut table = Table::new(&["TOKEN", "TYPE", "OBJECT", "CREATED", "URL"]);
                for link in links {
                    table.row(vec![
                        link.inner.id.clone(),
                        link.inner.r#type.clone(),
                        format!("{} ({})", link.inner.object_name, link.inner.object_id),
                        text::date(&link.inner.created_at),
                        link.url.clone(),
                    ]);
                }
                table.render("No public links.")
            })
        }
        LinksCommand::Create { object_id } => {
            let params = NewPublicShareParams { object_id };
            let link = session
                .fetch(async |api| api.create_public_share(params.clone()).await)
                .await?;
            let link = WithUrl {
                url: url(session, &link.id),
                inner: link,
            };
            output.print(&link, |link| link.url.clone())
        }
        LinksCommand::Show { token } => {
            let share = session
                .fetch(async |api| api.get_public_share(&token).await)
                .await?;
            output.print(&share, show)
        }
        LinksCommand::Delete { token } => {
            session
                .fetch(async |api| api.delete_public_share(&token).await)
                .await?;
            output.done(&format!("Revoked public link {token}"))
        }
    }
}

/// The frontend page of a public link.
fn url(session: &Session, token: &str) -> String {
    format!(
        "{}/public/{token}",
        session.server.as_str().trim_end_matches('/')
    )
}

fn show(share: &PublicShare) -> String {
    if let Some(list) = &share.list {
        let fields = text::fields(&[
            ("type", "list".to_owned()),
            ("id", list.id.clone()),
            ("name", list.name.clone()),
            ("owner", list.owner_name.clone()),
            ("created", text::date(&list.created_at)),
        ]);
        let mut table = Table::new(&["ID", "NAME", "QUANTITY"]);
        for thing in list.things.iter().flatten() {
            table.row(vec![thing.id.clone(), thing.name.clone(), quantity(thing)]);
        }
        return format!("{fields}\n{}", table.render("No things in this list."));
    }
    let Some(thing) = &share.thing else {
        return format!("public link {} shows nothing", share.token);
    };
    let mut out = text::fields(&[
        ("type", "thing".to_owned()),
        ("id", thing.id.clone()),
        ("name", thing.name.clone()),
        ("owner", thing.owner_name.clone()),
        ("quantity", quantity(thing)),
        ("created", text::date(&thing.created_at)),
        (
            "images",
            thing
                .images
                .iter()
                .flatten()
                .map(images::describe)
                .collect::<Vec<_>>()
                .join("\n"),
        ),
        ("description", thing.description.clone()),
    ]);
    out.push_str(&things::properties(
        thing.properties.as_deref().unwrap_or_default(),
    ));
    out
}

fn quantity(thing: &PublicThing) -> String {
    format!("{} {}", thing.quantity, thing.quantity_unit)
        .trim_end()
        .to_owned()
}
