use clap::Subcommand;

use crate::api::{FriendRequest, NewFriendRequestParams, UpdateFriendRequestParams, User};
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum FriendsCommand {
    /// List your friends
    List,
    /// List the friend requests you received and sent
    Requests,
    /// Send a friend request to a user
    Send {
        /// ID of the user
        #[arg(allow_hyphen_values = true)]
        user_id: String,
    },
    /// Withdraw a friend request you sent
    Cancel {
        /// ID of the friend request
        #[arg(allow_hyphen_values = true)]
        id: String,
    },
    /// Accept a pending friend request you received
    Accept {
        /// ID of the friend request
        #[arg(allow_hyphen_values = true)]
        id: String,
    },
    /// Reject a pending friend request you received
    Reject {
        /// ID of the friend request
        #[arg(allow_hyphen_values = true)]
        id: String,
    },
    /// End a friendship
    Remove {
        /// ID of the friend
        #[arg(allow_hyphen_values = true)]
        user_id: String,
    },
}

pub async fn run(
    command: FriendsCommand,
    session: &Session,
    output: &Output,
) -> anyhow::Result<()> {
    match command {
        FriendsCommand::List => {
            let friends = session.fetch(async |api| api.list_friends().await).await?;
            output.print(&friends, |friends| {
                let mut table = Table::new(&["ID", "NAME"]);
                for friendship in friends.friend_ships.iter().flatten() {
                    table.row(vec![
                        friendship.friend.id.clone(),
                        friendship.friend.name.clone(),
                    ]);
                }
                table.render("No friends yet.")
            })
        }
        FriendsCommand::Requests => {
            let requests = session
                .fetch(async |api| api.list_friend_requests().await)
                .await?;
            output.print(&requests, |requests| {
                let received = requests_table(
                    requests.received.as_deref().unwrap_or_default(),
                    "FROM",
                    |request| &request.sender,
                );
                let sent = requests_table(
                    requests.sent.as_deref().unwrap_or_default(),
                    "TO",
                    |request| &request.receiver,
                );
                format!(
                    "Received:\n{}\n\nSent:\n{}",
                    received.render("No friend requests received."),
                    sent.render("No friend requests sent."),
                )
            })
        }
        FriendsCommand::Send { user_id } => {
            let params = NewFriendRequestParams {
                receiver_id: user_id,
            };
            let request = session
                .fetch(async |api| api.create_friend_request(params.clone()).await)
                .await?;
            output.print(&request, |request| {
                format!(
                    "Sent friend request {} to {}",
                    request.id,
                    text::user(&request.receiver)
                )
            })
        }
        FriendsCommand::Cancel { id } => {
            session
                .fetch(async |api| api.cancel_friend_request(&id).await)
                .await?;
            output.done(&format!("Cancelled friend request {id}"))
        }
        FriendsCommand::Accept { id } => respond(&id, true, session, output).await,
        FriendsCommand::Reject { id } => respond(&id, false, session, output).await,
        FriendsCommand::Remove { user_id } => {
            session
                .fetch(async |api| api.remove_friend(&user_id).await)
                .await?;
            output.done(&format!("Removed friend {user_id}"))
        }
    }
}

/// A table of friend requests, showing the user picked by `user`.
fn requests_table(
    requests: &[FriendRequest],
    header: &'static str,
    user: fn(&FriendRequest) -> &User,
) -> Table {
    let mut table = Table::new(&["ID", header, "STATE", "DATE"]);
    for request in requests {
        table.row(vec![
            request.id.clone(),
            text::user(user(request)),
            request.state.clone(),
            text::date(&request.created_at),
        ]);
    }
    table
}

async fn respond(id: &str, accept: bool, session: &Session, output: &Output) -> anyhow::Result<()> {
    let params = UpdateFriendRequestParams { accept };
    let request = session
        .fetch(async |api| api.respond_to_friend_request(id, params.clone()).await)
        .await?;
    output.print(&request, |request| {
        format!(
            "{} friend request from {}",
            if accept { "Accepted" } else { "Rejected" },
            text::user(&request.sender)
        )
    })
}
