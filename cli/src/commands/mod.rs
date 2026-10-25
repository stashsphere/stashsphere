pub mod auth;
pub mod lists;
pub mod notifications;
pub mod search;
pub mod shares;
pub mod things;
pub mod users;

use clap::Args;

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
