//! The server address and login cookies, kept between invocations.
//!
//! The server is stored in `config.toml` in the config directory and the
//! cookies in `cookies.json` in the data directory. `STASHCTL_HOME`
//! replaces both directories, e.g. for tests or several accounts.

use std::fs::{self, File, OpenOptions};
use std::io::{BufReader, BufWriter};
use std::path::{Path, PathBuf};
use std::sync::Arc;

use anyhow::{Context, anyhow};
use cookie_store::CookieStore;
use directories::ProjectDirs;
use reqwest_cookie_store::CookieStoreMutex;
use serde::{Deserialize, Serialize};
use url::Url;

use crate::api::{ApiOpError, HttpClient};
use crate::error::{NotLoggedIn, api_error};

#[derive(Default, Deserialize, Serialize)]
struct Config {
    server: Option<Url>,
}

struct Paths {
    config: PathBuf,
    cookies: PathBuf,
}

impl Paths {
    fn new() -> anyhow::Result<Self> {
        if let Some(home) = std::env::var_os("STASHCTL_HOME") {
            let home = PathBuf::from(home);
            return Ok(Self {
                config: home.join("config.toml"),
                cookies: home.join("cookies.json"),
            });
        }
        let dirs = ProjectDirs::from("", "", "stashctl")
            .ok_or_else(|| anyhow!("cannot determine the home directory, set STASHCTL_HOME"))?;
        Ok(Self {
            config: dirs.config_dir().join("config.toml"),
            cookies: dirs.data_local_dir().join("cookies.json"),
        })
    }
}

pub struct Session {
    pub server: Url,
    pub api: HttpClient,
    cookies: Arc<CookieStoreMutex>,
    paths: Paths,
}

impl Session {
    /// Opens the session for `server`, or for the server of the last login.
    pub fn open(server: Option<Url>) -> anyhow::Result<Self> {
        let paths = Paths::new()?;
        let server = match server {
            Some(server) => server,
            None => read_config(&paths.config)?.server.ok_or(NotLoggedIn)?,
        };
        let cookies = Arc::new(CookieStoreMutex::new(read_cookies(&paths.cookies)?));
        let http = reqwest::Client::builder()
            .cookie_provider(cookies.clone())
            .build()?;
        let api =
            HttpClient::with_client(http).with_base_url(server.as_str().trim_end_matches('/'));
        Ok(Self {
            server,
            api,
            cookies,
            paths,
        })
    }

    /// Calls the API, refreshing the access token once if it has expired.
    // passes the generated client's error type through unchanged
    #[allow(clippy::result_large_err)]
    pub async fn call<T, E: std::fmt::Debug>(
        &self,
        request: impl AsyncFn(&HttpClient) -> Result<T, ApiOpError<E>>,
    ) -> Result<T, ApiOpError<E>> {
        match request(&self.api).await {
            Err(ApiOpError::Api(err)) if err.status == 401 => {
                if self.api.refresh_token().await.is_err() {
                    return Err(ApiOpError::Api(err));
                }
                request(&self.api).await
            }
            result => result,
        }
    }

    /// Calls the API like [`Session::call`] and keeps refreshed tokens. A
    /// rejected session becomes [`NotLoggedIn`].
    pub async fn fetch<T, E: std::fmt::Debug + Send + Sync + 'static>(
        &self,
        request: impl AsyncFn(&HttpClient) -> Result<T, ApiOpError<E>>,
    ) -> anyhow::Result<T> {
        let result = self.call(request).await;
        self.save_cookies()?;
        result.map_err(|err| match err.api() {
            Some(err) if err.status == 401 => NotLoggedIn.into(),
            _ => api_error(err),
        })
    }

    /// Stores the server and the current cookies for later invocations.
    pub fn save(&self) -> anyhow::Result<()> {
        write_config(
            &self.paths.config,
            &Config {
                server: Some(self.server.clone()),
            },
        )?;
        self.save_cookies()
    }

    fn save_cookies(&self) -> anyhow::Result<()> {
        let cookies = self
            .cookies
            .lock()
            .map_err(|_| anyhow!("cookie store poisoned"))?;
        let file = create_private(&self.paths.cookies)?;
        cookie_store::serde::json::save(&cookies, &mut BufWriter::new(file))
            .map_err(|err| anyhow!("writing {}: {err}", self.paths.cookies.display()))
    }

    /// Removes the stored cookies.
    pub fn clear(&self) -> anyhow::Result<()> {
        self.cookies
            .lock()
            .map_err(|_| anyhow!("cookie store poisoned"))?
            .clear();
        match fs::remove_file(&self.paths.cookies) {
            Err(err) if err.kind() != std::io::ErrorKind::NotFound => {
                Err(err).with_context(|| format!("removing {}", self.paths.cookies.display()))
            }
            _ => Ok(()),
        }
    }
}

fn read_config(path: &Path) -> anyhow::Result<Config> {
    match fs::read_to_string(path) {
        Ok(text) => toml::from_str(&text).with_context(|| format!("reading {}", path.display())),
        Err(err) if err.kind() == std::io::ErrorKind::NotFound => Ok(Config::default()),
        Err(err) => Err(err).with_context(|| format!("reading {}", path.display())),
    }
}

fn write_config(path: &Path, config: &Config) -> anyhow::Result<()> {
    if let Some(dir) = path.parent() {
        fs::create_dir_all(dir).with_context(|| format!("creating {}", dir.display()))?;
    }
    fs::write(path, toml::to_string(config)?).with_context(|| format!("writing {}", path.display()))
}

fn read_cookies(path: &Path) -> anyhow::Result<CookieStore> {
    match File::open(path) {
        Ok(file) => cookie_store::serde::json::load(BufReader::new(file))
            .map_err(|err| anyhow!("reading {}: {err}", path.display())),
        Err(err) if err.kind() == std::io::ErrorKind::NotFound => Ok(CookieStore::default()),
        Err(err) => Err(err).with_context(|| format!("reading {}", path.display())),
    }
}

/// Creates or truncates a file only the current user can read, as it holds
/// the session tokens.
fn create_private(path: &Path) -> anyhow::Result<File> {
    if let Some(dir) = path.parent() {
        fs::create_dir_all(dir).with_context(|| format!("creating {}", dir.display()))?;
    }
    let mut options = OpenOptions::new();
    options.write(true).create(true).truncate(true);
    #[cfg(unix)]
    std::os::unix::fs::OpenOptionsExt::mode(&mut options, 0o600);
    options
        .open(path)
        .with_context(|| format!("writing {}", path.display()))
}
