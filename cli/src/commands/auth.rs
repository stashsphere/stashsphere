use std::io::{IsTerminal, Read};

use anyhow::{Context, bail};
use clap::Args;
use url::Url;

use crate::api::LoginPostParams;
use crate::error::api_error;
use crate::output::Output;
use crate::session::Session;

#[derive(Args)]
pub struct LoginArgs {
    /// Email address of the account
    #[arg(long, env = "STASHCTL_EMAIL")]
    email: String,

    /// Read the password from stdin instead of prompting for it
    #[arg(long)]
    password_stdin: bool,
}

pub async fn login(server: Option<Url>, args: LoginArgs, output: &Output) -> anyhow::Result<()> {
    let session = Session::open(server)?;
    let password = read_password(args.password_stdin)?;
    session
        .api
        .login(LoginPostParams {
            email: args.email,
            password,
        })
        .await
        .map_err(api_error)?;
    session.save()?;
    output.done(&format!("Logged in to {}", session.server))
}

/// The password comes from STASHCTL_PASSWORD, stdin or a prompt, never from
/// a flag, which would leak it into the process list and shell history.
fn read_password(from_stdin: bool) -> anyhow::Result<String> {
    if from_stdin {
        let mut password = String::new();
        std::io::stdin()
            .read_to_string(&mut password)
            .context("reading the password from stdin")?;
        return Ok(password.trim_end_matches(['\r', '\n']).to_owned());
    }
    if let Ok(password) = std::env::var("STASHCTL_PASSWORD") {
        return Ok(password);
    }
    if !std::io::stdin().is_terminal() {
        bail!("no password given, use --password-stdin or STASHCTL_PASSWORD");
    }
    rpassword::prompt_password("Password: ").context("reading the password")
}

pub async fn logout(server: Option<Url>, output: &Output) -> anyhow::Result<()> {
    let session = Session::open(server)?;
    // the stored session is removed even if the server already forgot it
    let result = session.api.logout().await;
    session.clear()?;
    if let Err(err) = result {
        let err = api_error(err);
        if !err.is::<crate::error::ApiFailure>() {
            return Err(err);
        }
    }
    output.done("Logged out")
}

pub async fn whoami(server: Option<Url>, output: &Output) -> anyhow::Result<()> {
    let session = Session::open(server)?;
    let profile = session.fetch(async |api| api.get_profile().await).await?;
    output.print(&profile, |profile| {
        format!("{} <{}> on {}", profile.name, profile.email, session.server)
    })
}
