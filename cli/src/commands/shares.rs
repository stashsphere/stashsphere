use clap::Subcommand;

use crate::api::{NewShareParams, Share};
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
    /// Share a thing or list you own with another user
    Create {
        /// ID of the thing or list
        object_id: String,

        /// ID of the user to share with
        #[arg(long, value_name = "ID")]
        user: String,
    },
    /// Stop sharing a thing or list with a user
    Delete {
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
        SharesCommand::Create { object_id, user } => {
            let params = NewShareParams {
                object_id,
                target_user_id: user,
            };
            let share = session
                .fetch(async |api| api.create_share(params.clone()).await)
                .await?;
            output.print(&share, show)
        }
        SharesCommand::Delete { id } => {
            session
                .fetch(async |api| api.delete_share(&id).await)
                .await?;
            output.done(&format!("Deleted share {id}"))
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
