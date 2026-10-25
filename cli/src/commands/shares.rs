use clap::Subcommand;

use crate::api::Share;
use crate::output::Output;
use crate::session::Session;
use crate::text;

#[derive(Subcommand)]
pub enum SharesCommand {
    /// Show a share of a thing or list with another user
    Show {
        /// ID of the share
        id: String,
    },
}

pub async fn run(command: SharesCommand, session: &Session, output: &Output) -> anyhow::Result<()> {
    match command {
        SharesCommand::Show { id } => {
            let share = session.fetch(async |api| api.get_share(&id).await).await?;
            output.print(&share, show)
        }
    }
}

fn show(share: &Share) -> String {
    let (id, kind, object, owner, target) = match share {
        Share::ThingShare(share) => (
            &share.id,
            "thing",
            format!("{} ({})", share.object.name, share.object.id),
            &share.owner,
            &share.target_user,
        ),
        Share::ListShare(share) => (
            &share.id,
            "list",
            format!("{} ({})", share.object.name, share.object.id),
            &share.owner,
            &share.target_user,
        ),
    };
    text::fields(&[
        ("id", id.clone()),
        ("type", kind.to_owned()),
        (kind, object),
        ("owner", text::user(owner)),
        ("shared with", text::user(target)),
    ])
}
