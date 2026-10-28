use url::Url;

use crate::error::api_error;
use crate::output::Output;
use crate::session::Session;
use crate::text;

/// Shows how to sign up for and log in to the instance. Needs no login.
pub async fn run(server: Option<Url>, output: &Output) -> anyhow::Result<()> {
    let session = Session::open(server)?;
    // a plain call, so that asking an unknown server keeps the stored session
    let info = session
        .call(async |api| api.get_info().await)
        .await
        .map_err(api_error)?;
    output.print(&info, |info| {
        let providers: Vec<_> = info
            .oidc_providers
            .iter()
            .flatten()
            .map(|provider| format!("{} ({})", provider.display_name, provider.name))
            .collect();
        text::fields(&[
            ("server", session.server.to_string()),
            (
                "sign up",
                if info.invite_required {
                    "with an invite code"
                } else {
                    "open"
                }
                .to_owned(),
            ),
            ("login providers", providers.join("\n")),
        ])
    })
}
