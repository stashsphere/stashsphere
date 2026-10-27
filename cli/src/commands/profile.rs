use std::io::{IsTerminal, Read};

use anyhow::{Context, bail};
use clap::{Args, Subcommand};

use crate::api::{
    PatchPasswordParams, Profile, ProfileUpdateParams, ScheduleDeletionParams, VerifyEmailParams,
};
use crate::commands::{Sharing, auth, images};
use crate::output::Output;
use crate::session::Session;
use crate::text;

#[derive(Subcommand)]
pub enum ProfileCommand {
    /// Show your profile
    Show,
    /// Change your profile
    Update(UpdateArgs),
    /// Change your password
    Password {
        /// Read the current and the new password from the first two lines
        /// of stdin instead of prompting for them
        #[arg(long)]
        password_stdin: bool,
    },
    /// Email a verification code to your address
    SendVerification,
    /// Verify your email address with the code from the email
    VerifyEmail {
        /// The 8 digit code
        code: String,
    },
    /// Schedule the deletion of your account after a grace period
    DeleteAccount {
        /// Read the password from stdin instead of prompting for it
        #[arg(long)]
        password_stdin: bool,
    },
    /// Cancel the scheduled deletion of your account
    CancelDeletion,
}

#[derive(Args)]
pub struct UpdateArgs {
    /// Name shown to other users
    #[arg(long)]
    name: Option<String>,

    /// Full name
    #[arg(long)]
    full_name: Option<String>,

    /// About you
    #[arg(long)]
    information: Option<String>,

    /// Default sharing of new things and lists
    #[arg(long, value_enum)]
    sharing: Option<Sharing>,

    /// ID of one of your images to show as your picture
    #[arg(
        long,
        value_name = "ID",
        allow_hyphen_values = true,
        conflicts_with = "remove_image"
    )]
    image: Option<String>,

    /// Remove your picture
    #[arg(long)]
    remove_image: bool,
}

pub async fn run(
    command: ProfileCommand,
    session: &Session,
    output: &Output,
) -> anyhow::Result<()> {
    match command {
        ProfileCommand::Show => {
            let profile = session.fetch(async |api| api.get_profile().await).await?;
            output.print(&profile, show)
        }
        ProfileCommand::Update(args) => {
            // the API replaces the whole profile
            let profile = session.fetch(async |api| api.get_profile().await).await?;
            let params = ProfileUpdateParams {
                name: args.name.unwrap_or(profile.name),
                full_name: args.full_name.or(profile.full_name).unwrap_or_default(),
                information: args.information.or(profile.information).unwrap_or_default(),
                default_sharing_state: args
                    .sharing
                    .map(Into::into)
                    .unwrap_or(profile.default_sharing_state),
                image_id: if args.remove_image {
                    None
                } else {
                    args.image.or(profile.image.map(|image| image.id))
                },
            };
            let profile = session
                .fetch(async |api| api.update_profile(params.clone()).await)
                .await?;
            output.print(&profile, show)
        }
        ProfileCommand::Password { password_stdin } => {
            let (old_password, new_password) = read_password_change(password_stdin)?;
            let params = PatchPasswordParams {
                old_password,
                new_password,
            };
            session
                .fetch(async |api| api.update_password(params.clone()).await)
                .await?;
            output.done("Changed your password")
        }
        ProfileCommand::SendVerification => {
            let profile = session.fetch(async |api| api.get_profile().await).await?;
            session
                .fetch(async |api| api.request_email_verification().await)
                .await?;
            output.done(&format!("Sent a verification code to {}", profile.email))
        }
        ProfileCommand::VerifyEmail { code } => {
            let params = VerifyEmailParams { code };
            session
                .fetch(async |api| api.verify_email(params.clone()).await)
                .await?;
            output.done("Verified your email address")
        }
        ProfileCommand::DeleteAccount { password_stdin } => {
            let params = ScheduleDeletionParams {
                password: auth::read_password(password_stdin)?,
            };
            let profile = session
                .fetch(async |api| api.schedule_account_deletion(params.clone()).await)
                .await?;
            output.print(&profile, |profile| match &profile.purge_at {
                Some(purge_at) => format!(
                    "Your account will be deleted at {}, cancel with: stashctl profile cancel-deletion",
                    text::date(purge_at)
                ),
                None => "Your account is not scheduled for deletion".to_owned(),
            })
        }
        ProfileCommand::CancelDeletion => {
            let profile = session
                .fetch(async |api| api.cancel_account_deletion().await)
                .await?;
            output.print(&profile, |_| {
                "Cancelled the deletion of your account".to_owned()
            })
        }
    }
}

/// The current and the new password, from stdin or prompts. Like
/// [`auth::read_password`], passwords never come from flags.
fn read_password_change(from_stdin: bool) -> anyhow::Result<(String, String)> {
    if from_stdin {
        let mut input = String::new();
        std::io::stdin()
            .read_to_string(&mut input)
            .context("reading the passwords from stdin")?;
        let mut lines = input.lines();
        let (Some(old), Some(new)) = (lines.next(), lines.next()) else {
            bail!("expected the current and the new password on two lines of stdin");
        };
        return Ok((old.to_owned(), new.to_owned()));
    }
    if !std::io::stdin().is_terminal() {
        bail!("no passwords given, use --password-stdin");
    }
    let old = rpassword::prompt_password("Current password: ").context("reading the password")?;
    let new = rpassword::prompt_password("New password: ").context("reading the password")?;
    let repeated =
        rpassword::prompt_password("Repeat new password: ").context("reading the password")?;
    if new != repeated {
        bail!("the new passwords do not match");
    }
    Ok((old, new))
}

fn show(profile: &Profile) -> String {
    let verified = match profile.email_verified {
        Some(Some(true)) => " (verified)",
        Some(Some(false)) => " (not verified)",
        _ => "",
    };
    text::fields(&[
        ("id", profile.id.clone()),
        ("name", profile.name.clone()),
        ("full name", profile.full_name.clone().unwrap_or_default()),
        ("email", format!("{}{verified}", profile.email)),
        (
            "default sharing",
            profile.default_sharing_state.as_str().to_owned(),
        ),
        (
            "image",
            profile
                .image
                .as_ref()
                .map(images::describe)
                .unwrap_or_default(),
        ),
        (
            "external login",
            if profile.external_auth { "yes" } else { "" }.to_owned(),
        ),
        (
            "deletion",
            profile
                .purge_at
                .as_ref()
                .map(|purge_at| format!("scheduled for {}", text::date(purge_at)))
                .unwrap_or_default(),
        ),
        (
            "information",
            profile.information.clone().unwrap_or_default(),
        ),
    ])
}
