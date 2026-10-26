use chrono::{DateTime, Utc};
use clap::{Args, Subcommand};

use crate::api::Notification;
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum NotificationsCommand {
    /// List your notifications
    List(ListArgs),
    /// Mark notifications as acknowledged
    #[command(alias = "ack")]
    Acknowledge {
        /// IDs of the notifications
        #[arg(required = true)]
        ids: Vec<String>,
    },
}

#[derive(Args)]
pub struct ListArgs {
    /// Page to show, starting at 0
    #[arg(long)]
    page: Option<i64>,

    /// Number of notifications per page
    #[arg(long)]
    per_page: Option<i64>,

    /// Only notifications that have not been acknowledged
    #[arg(long)]
    unacknowledged: bool,
}

pub async fn run(
    command: NotificationsCommand,
    session: &Session,
    output: &Output,
) -> anyhow::Result<()> {
    match command {
        NotificationsCommand::List(args) => {
            let notifications = session
                .fetch(async |api| {
                    api.list_notifications(
                        args.page,
                        args.per_page,
                        args.unacknowledged.then_some(true),
                    )
                    .await
                })
                .await?;
            output.print(&notifications, |notifications| {
                let mut table = Table::new(&["ID", "DATE", "ACK", "EVENT"]);
                for notification in notifications.notifications.iter().flatten() {
                    let (id, created_at, acknowledged, event) = describe(notification);
                    table.row(vec![
                        id.to_owned(),
                        text::date(created_at),
                        if acknowledged { "yes" } else { "no" }.to_owned(),
                        event,
                    ]);
                }
                let footer = text::page_footer(
                    notifications.page,
                    notifications.total_page_count,
                    notifications.total_count,
                    ("notification", "notifications"),
                );
                format!("{}\n{footer}", table.render("No notifications."))
            })
        }
        NotificationsCommand::Acknowledge { ids } => {
            for id in &ids {
                session
                    .fetch(async |api| api.acknowledge_notification(id).await)
                    .await?;
            }
            let noun = if ids.len() == 1 {
                "notification"
            } else {
                "notifications"
            };
            output.done(&format!("Acknowledged {} {noun}", ids.len()))
        }
    }
}

fn describe(notification: &Notification) -> (&str, &DateTime<Utc>, bool, String) {
    match notification {
        Notification::FriendRequestNotification(n) => (
            &n.id,
            &n.created_at,
            n.acknowledged,
            format!(
                "friend request from {} ({})",
                n.content.sender.name, n.content.state
            ),
        ),
        Notification::FriendRequestReactionNotification(n) => (
            &n.id,
            &n.created_at,
            n.acknowledged,
            format!(
                "friend request {} {}",
                n.content.request_id,
                if n.content.accepted {
                    "accepted"
                } else {
                    "rejected"
                }
            ),
        ),
        Notification::ThingSharedNotification(n) => (
            &n.id,
            &n.created_at,
            n.acknowledged,
            format!(
                "thing {} shared by user {}",
                n.content.thing_id, n.content.sharer_id
            ),
        ),
        Notification::ListSharedNotification(n) => (
            &n.id,
            &n.created_at,
            n.acknowledged,
            format!(
                "list {} shared by user {}",
                n.content.list_id, n.content.sharer_id
            ),
        ),
        Notification::ThingsAddedToListNotification(n) => (
            &n.id,
            &n.created_at,
            n.acknowledged,
            format!(
                "things added to list {} by user {}",
                n.content.list_id, n.content.added_by_id
            ),
        ),
    }
}
