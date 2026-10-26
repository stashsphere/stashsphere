use clap::Subcommand;

use crate::api::UserProfile;
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum UsersCommand {
    /// List the users of the instance
    List,
    /// Show the profile of a user
    Show {
        /// ID of the user
        #[arg(allow_hyphen_values = true)]
        id: String,
    },
}

pub async fn run(command: UsersCommand, session: &Session, output: &Output) -> anyhow::Result<()> {
    match command {
        UsersCommand::List => {
            let users = session.fetch(async |api| api.list_users().await).await?;
            output.print(&users, |users| {
                let mut table = Table::new(&["ID", "NAME", "FULL NAME"]);
                for user in users {
                    table.row(vec![
                        user.id.clone(),
                        user.name.clone(),
                        user.full_name.clone().unwrap_or_default(),
                    ]);
                }
                table.render("No users found.")
            })
        }
        UsersCommand::Show { id } => {
            let user = session.fetch(async |api| api.get_user(&id).await).await?;
            output.print(&user, show)
        }
    }
}

fn show(user: &UserProfile) -> String {
    text::fields(&[
        ("id", user.id.clone()),
        ("name", user.name.clone()),
        ("full name", user.full_name.clone().unwrap_or_default()),
        ("information", user.information.clone().unwrap_or_default()),
    ])
}
