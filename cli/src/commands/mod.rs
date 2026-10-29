pub mod archives;
pub mod auth;
pub mod cart;
pub mod friends;
pub mod images;
pub mod info;
pub mod links;
pub mod lists;
pub mod notifications;
pub mod profile;
pub mod properties;
pub mod search;
pub mod shares;
pub mod things;
pub mod users;

use clap::{Args, ValueEnum};

use crate::api::SharingState;
use crate::session::Session;

/// Who besides the owner can see a thing or list.
#[derive(Clone, Copy, ValueEnum)]
pub enum Sharing {
    /// Only you and users you share it with
    Private,
    /// Your friends
    Friends,
    /// Your friends and their friends
    FriendsOfFriends,
}

impl From<Sharing> for SharingState {
    fn from(sharing: Sharing) -> Self {
        match sharing {
            Sharing::Private => SharingState::Private,
            Sharing::Friends => SharingState::Friends,
            Sharing::FriendsOfFriends => SharingState::FriendsOfFriends,
        }
    }
}

/// The sharing state for a new object: `sharing`, or the default from the
/// profile, as in the web app.
async fn new_sharing_state(
    sharing: Option<Sharing>,
    session: &Session,
) -> anyhow::Result<SharingState> {
    match sharing {
        Some(sharing) => Ok(sharing.into()),
        None => Ok(session
            .fetch(async |api| api.get_profile().await)
            .await?
            .default_sharing_state),
    }
}

/// Selects a page of a paginated listing.
#[derive(Args)]
pub struct PageArgs {
    /// Page to show, starting at 0
    #[arg(long)]
    page: Option<i64>,

    /// Number of results per page
    #[arg(long)]
    per_page: Option<i64>,

    /// Show all results at once instead of a page
    #[arg(long, conflicts_with_all = ["page", "per_page"])]
    all: bool,
}

impl PageArgs {
    /// The `paginate` query parameter: false to request all results
    fn paginate(&self) -> Option<bool> {
        self.all.then_some(false)
    }
}
