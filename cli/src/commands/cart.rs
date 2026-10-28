use clap::Subcommand;

use crate::api::{Cart, CartParams, Thing};
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum CartCommand {
    /// Show the things in your cart, grouped by owner
    Show,
    /// Put things into your cart
    Add {
        /// IDs of things you can see
        #[arg(required = true, allow_hyphen_values = true)]
        thing_ids: Vec<String>,
    },
    /// Take things out of your cart
    Remove {
        /// IDs of things in your cart
        #[arg(required = true, allow_hyphen_values = true)]
        thing_ids: Vec<String>,
    },
    /// Empty your cart
    Clear,
}

pub async fn run(command: CartCommand, session: &Session, output: &Output) -> anyhow::Result<()> {
    let cart = match command {
        CartCommand::Show => session.fetch(async |api| api.get_cart().await).await?,
        CartCommand::Add { thing_ids } => {
            // the API replaces the whole cart
            let mut ids = thing_ids_of(&session.fetch(async |api| api.get_cart().await).await?);
            for id in thing_ids {
                if !ids.contains(&id) {
                    ids.push(id);
                }
            }
            update(ids, session).await?
        }
        CartCommand::Remove { thing_ids } => {
            let mut ids = thing_ids_of(&session.fetch(async |api| api.get_cart().await).await?);
            ids.retain(|id| !thing_ids.contains(id));
            update(ids, session).await?
        }
        CartCommand::Clear => update(Vec::new(), session).await?,
    };
    let things = if output.is_text() {
        describe(&cart, session).await
    } else {
        Vec::new()
    };
    output.print(&cart, |cart| show(cart, &things))
}

fn thing_ids_of(cart: &Cart) -> Vec<String> {
    cart.entries
        .iter()
        .flatten()
        .map(|entry| entry.thing_id.clone())
        .collect()
}

async fn update(thing_ids: Vec<String>, session: &Session) -> anyhow::Result<Cart> {
    let params = CartParams {
        thing_ids: Some(thing_ids),
    };
    session
        .fetch(async |api| api.update_cart(params.clone()).await)
        .await
}

/// The things in the cart, as far as they are still visible, since the
/// cart itself only knows their IDs.
async fn describe(cart: &Cart, session: &Session) -> Vec<Thing> {
    let mut things = Vec::new();
    for id in thing_ids_of(cart) {
        if let Ok(thing) = session.fetch(async |api| api.get_thing(&id).await).await {
            things.push(thing);
        }
    }
    things
}

fn show(cart: &Cart, things: &[Thing]) -> String {
    let entries = cart.entries.as_deref().unwrap_or_default();
    let mut owners: Vec<&str> = Vec::new();
    for entry in entries {
        if !owners.contains(&entry.owner_id.as_str()) {
            owners.push(&entry.owner_id);
        }
    }
    let mut out = Vec::new();
    for owner in owners {
        let owned: Vec<_> = entries
            .iter()
            .filter(|entry| entry.owner_id == owner)
            .collect();
        let thing = |id: &str| things.iter().find(|thing| thing.id == id);
        let owner_name = owned
            .iter()
            .find_map(|entry| thing(&entry.thing_id))
            .map(|thing| text::user(&thing.owner))
            .unwrap_or_else(|| owner.to_owned());
        let mut table = Table::new(&["ID", "NAME", "ADDED"]);
        for entry in owned {
            table.row(vec![
                entry.thing_id.clone(),
                thing(&entry.thing_id)
                    .map(|thing| thing.name.clone())
                    .unwrap_or_default(),
                text::date(&entry.created_at),
            ]);
        }
        out.push(format!("Owned by {owner_name}:\n{}", table.render("")));
    }
    if out.is_empty() {
        return "Your cart is empty.".to_owned();
    }
    out.join("\n\n")
}
