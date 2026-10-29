use clap::Subcommand;
use serde_json::Value;

use crate::api::{SchemaCollection, SchemaIcons};
use crate::error::NotFound;
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum PropertiesCommand {
    /// List the property names the instance has schemas for, with their
    /// types and units
    List,
    /// Show the schema of a property: its type, units, allowed values and
    /// aliases
    Show {
        /// Name or alias of the property
        name: String,
    },
}

pub async fn run(
    command: PropertiesCommand,
    session: &Session,
    output: &Output,
) -> anyhow::Result<()> {
    let collection = session
        .fetch(async |api| api.list_property_schemas().await)
        .await?;
    match command {
        PropertiesCommand::List => output.print(&collection, |collection| {
            let mut table = Table::new(&["NAME", "TYPE", "UNITS", "ALIASES", "TITLE"]);
            for (name, entry) in schemas(collection) {
                let schema = Schema::of(entry);
                table.row(vec![
                    name.clone(),
                    schema.kind.to_owned(),
                    schema.units.join(", "),
                    aliases(collection, name).join(", "),
                    schema.title.to_owned(),
                ]);
            }
            table.render("No property schemas.")
        }),
        PropertiesCommand::Show { name } => {
            let name = resolve(&collection, &name)?;
            let entry = &schemas(&collection)[&name];
            output.print(entry, |entry| {
                let schema = Schema::of(entry);
                text::fields(&[
                    ("name", name.clone()),
                    ("title", schema.title.to_owned()),
                    ("description", schema.description.to_owned()),
                    ("type", schema.kind.to_owned()),
                    ("units", schema.units.join(", ")),
                    ("range", schema.range),
                    ("values", schema.values.join(", ")),
                    ("pattern", schema.pattern.to_owned()),
                    ("aliases", aliases(&collection, &name).join(", ")),
                    (
                        "icons",
                        entry.icons.as_deref().unwrap_or_default().join(", "),
                    ),
                ])
            })
        }
    }
}

fn schemas(collection: &SchemaCollection) -> &std::collections::BTreeMap<String, SchemaIcons> {
    static EMPTY: std::collections::BTreeMap<String, SchemaIcons> =
        std::collections::BTreeMap::new();
    collection
        .schemas
        .as_ref()
        .map(|schemas| &schemas.additional_properties)
        .unwrap_or(&EMPTY)
}

/// The aliases that point to `name`, sorted.
fn aliases<'a>(collection: &'a SchemaCollection, name: &str) -> Vec<&'a str> {
    let mut aliases: Vec<&str> = collection
        .aliases
        .iter()
        .flat_map(|aliases| &aliases.additional_properties)
        .filter(|(_, target)| *target == name)
        .map(|(alias, _)| alias.as_str())
        .collect();
    aliases.sort();
    aliases
}

/// The property name for a name or an alias, ignoring case like the web app.
fn resolve(collection: &SchemaCollection, name: &str) -> anyhow::Result<String> {
    let name = name.to_lowercase();
    if schemas(collection).contains_key(&name) {
        return Ok(name);
    }
    collection
        .aliases
        .as_ref()
        .and_then(|aliases| aliases.additional_properties.get(&name))
        .cloned()
        .ok_or_else(|| NotFound(format!("no property schema named {name}")).into())
}

/// The parts of a property schema worth showing. The schemas follow the root
/// schemas of the backend: the `value` is a number, a boolean, a date-time
/// string, or a string with an optional enum, format or pattern.
struct Schema<'a> {
    title: &'a str,
    description: &'a str,
    /// The type as `things create --property NAME:TYPE=VALUE` takes it
    kind: &'static str,
    units: Vec<&'a str>,
    values: Vec<&'a str>,
    range: String,
    pattern: &'a str,
}

impl<'a> Schema<'a> {
    fn of(entry: &'a SchemaIcons) -> Self {
        let null = &Value::Null;
        let schema = entry.schema.as_ref().unwrap_or(null);
        let property = |field: &str| &schema["properties"][field];
        let value = property("value");
        let strings = |value: &'a Value| -> Vec<&'a str> {
            value
                .as_array()
                .into_iter()
                .flatten()
                .filter_map(Value::as_str)
                .collect()
        };
        let kind = match value["type"].as_str() {
            Some("number") => "float",
            Some("boolean") => "boolean",
            Some("string") if value["format"] == "date-time" => "datetime",
            _ => "string",
        };
        let bound = |field: &str| value[field].as_f64().map(|n| n.to_string());
        let range = match (bound("minimum"), bound("maximum")) {
            (None, None) => String::new(),
            (min, max) => format!(
                "{} to {}",
                min.as_deref().unwrap_or("-∞"),
                max.as_deref().unwrap_or("∞")
            ),
        };
        Schema {
            title: schema["title"].as_str().unwrap_or_default(),
            description: schema["description"].as_str().unwrap_or_default(),
            kind,
            units: strings(&property("unit")["enum"]),
            values: strings(&value["enum"]),
            range,
            pattern: value["pattern"].as_str().unwrap_or_default(),
        }
    }
}
