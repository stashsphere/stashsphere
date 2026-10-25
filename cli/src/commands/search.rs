use clap::Args;

use crate::commands::{lists, things};
use crate::output::Output;
use crate::session::Session;

#[derive(Args)]
pub struct SearchArgs {
    /// Text to find, case-insensitively, in the name, description and your
    /// private note of things and in the name of lists
    query: String,
}

pub async fn run(args: SearchArgs, session: &Session, output: &Output) -> anyhow::Result<()> {
    let result = session
        .fetch(async |api| api.search(&args.query).await)
        .await?;
    output.print(&result, |result| {
        let things = things::table(result.things.as_deref().unwrap_or_default());
        let lists = lists::table(result.lists.as_deref().unwrap_or_default());
        format!(
            "Things:\n{}\n\nLists:\n{}",
            things.render("No things found."),
            lists.render("No lists found."),
        )
    })
}
