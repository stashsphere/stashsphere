use anyhow::Context;
use chrono::{DateTime, NaiveDate, Utc};
use clap::{Args, Subcommand};

use crate::api::{
    NewThingParams, Property, PropertyBooleanParam, PropertyDatetimeParam, PropertyFloatParam,
    PropertyParam, PropertyStringParam, Thing,
};
use crate::commands::{PageArgs, Sharing, images, new_sharing_state};
use crate::output::Output;
use crate::session::Session;
use crate::text::{self, Table};

#[derive(Subcommand)]
pub enum ThingsCommand {
    /// List your things and the things shared with you
    List(ListArgs),
    /// Show a thing with its properties, lists and shares
    Show {
        /// ID of the thing
        #[arg(allow_hyphen_values = true)]
        id: String,
    },
    /// Create a thing
    Create(CreateArgs),
    /// Change fields of a thing you own, keeping the others
    Update(UpdateArgs),
    /// Delete a thing you own
    Delete {
        /// ID of the thing
        #[arg(allow_hyphen_values = true)]
        id: String,
    },
}

#[derive(Args)]
pub struct ListArgs {
    #[command(flatten)]
    page: PageArgs,

    /// Only things owned by this user ID
    #[arg(long, allow_hyphen_values = true)]
    owner: Option<String>,

    /// Only things whose name starts with this term, ignoring case
    #[arg(long)]
    search: Option<String>,

    /// Sort order as field:direction, e.g. created_at:desc or
    /// access_reason:asc
    #[arg(long)]
    order: Option<String>,
}

#[derive(Args)]
pub struct CreateArgs {
    /// Name of the thing, at least 4 characters
    #[arg(long)]
    name: String,

    #[command(flatten)]
    fields: FieldArgs,

    /// ID of an image you uploaded, can be repeated. The first image is
    /// shown first.
    #[arg(long = "image", value_name = "ID", allow_hyphen_values = true)]
    images: Vec<String>,
}

#[derive(Args)]
pub struct UpdateArgs {
    /// ID of the thing
    #[arg(allow_hyphen_values = true)]
    id: String,

    /// New name, at least 4 characters
    #[arg(long)]
    name: Option<String>,

    #[command(flatten)]
    fields: FieldArgs,

    /// Remove the property with this name, can be repeated
    #[arg(long, value_name = "NAME")]
    remove_property: Vec<String>,

    /// ID of an image you uploaded to add after the others, can be repeated
    #[arg(long, value_name = "ID", allow_hyphen_values = true)]
    add_image: Vec<String>,

    /// ID of an image to remove from the thing, can be repeated. The image
    /// itself is kept.
    #[arg(long, value_name = "ID", allow_hyphen_values = true)]
    remove_image: Vec<String>,
}

/// Fields of a thing besides the name.
#[derive(Args)]
struct FieldArgs {
    /// Description, visible to everyone who can see the thing
    #[arg(long)]
    description: Option<String>,

    /// Note only you can see
    #[arg(long)]
    private_note: Option<String>,

    /// How many of the thing you have
    #[arg(long)]
    quantity: Option<u64>,

    /// Unit of the quantity, e.g. pcs
    #[arg(long)]
    unit: Option<String>,

    /// Who can see the thing. New things default to the sharing state of
    /// your profile.
    #[arg(long, value_enum)]
    sharing: Option<Sharing>,

    /// Set a property, replacing one with the same name. TYPE is string
    /// (the default), float, datetime or boolean. A float may be followed
    /// by a unit, e.g. 'weight:float=1.5 kg'; a datetime is RFC 3339 or a
    /// YYYY-MM-DD date. Can be repeated.
    #[arg(
        long = "property",
        value_name = "NAME[:TYPE]=VALUE",
        value_parser = parse_property
    )]
    properties: Vec<PropertyParam>,
}

pub async fn run(command: ThingsCommand, session: &Session, output: &Output) -> anyhow::Result<()> {
    match command {
        ThingsCommand::List(args) => {
            let things = session
                .fetch(async |api| {
                    api.list_things(
                        args.page.page,
                        args.page.per_page,
                        args.owner.as_deref(),
                        args.search.as_deref(),
                        args.page.paginate(),
                        args.order.as_deref(),
                    )
                    .await
                })
                .await?;
            output.print(&things, |things| {
                let table = table(things.things.as_deref().unwrap_or_default());
                let footer = text::page_footer(
                    things.page,
                    things.total_page_count,
                    things.total_count,
                    ("thing", "things"),
                );
                format!("{}\n{footer}", table.render("No things found."))
            })
        }
        ThingsCommand::Show { id } => {
            let thing = session.fetch(async |api| api.get_thing(&id).await).await?;
            output.print(&thing, show)
        }
        ThingsCommand::Create(args) => {
            let fields = args.fields;
            let mut params = NewThingParams {
                name: args.name,
                description: fields.description.unwrap_or_default(),
                private_note: fields.private_note.unwrap_or_default(),
                quantity: fields.quantity.unwrap_or_default(),
                quantity_unit: fields.unit.unwrap_or_default(),
                sharing_state: new_sharing_state(fields.sharing, session).await?,
                images_ids: Some(args.images),
                properties: Some(Vec::new()),
            };
            set_properties(&mut params, fields.properties);
            let thing = session
                .fetch(async |api| api.create_thing(params.clone()).await)
                .await?;
            output.print(&thing, |thing| {
                format!("Created thing {} ({})", thing.name, thing.id)
            })
        }
        ThingsCommand::Update(args) => {
            // the API replaces the whole thing, so the unchanged fields are
            // sent as they are
            let thing = session
                .fetch(async |api| api.get_thing(&args.id).await)
                .await?;
            let mut params = params_of(thing)?;
            let fields = args.fields;
            if let Some(name) = args.name {
                params.name = name;
            }
            if let Some(description) = fields.description {
                params.description = description;
            }
            if let Some(private_note) = fields.private_note {
                params.private_note = private_note;
            }
            if let Some(quantity) = fields.quantity {
                params.quantity = quantity;
            }
            if let Some(unit) = fields.unit {
                params.quantity_unit = unit;
            }
            if let Some(sharing) = fields.sharing {
                params.sharing_state = sharing.into();
            }
            if let Some(properties) = &mut params.properties {
                properties.retain(|property| {
                    !args
                        .remove_property
                        .iter()
                        .any(|name| name == property_name(property))
                });
            }
            set_properties(&mut params, fields.properties);
            if let Some(images_ids) = &mut params.images_ids {
                images_ids.retain(|id| !args.remove_image.contains(id));
                for id in args.add_image {
                    if !images_ids.contains(&id) {
                        images_ids.push(id);
                    }
                }
            }
            let thing = session
                .fetch(async |api| api.update_thing(&args.id, params.clone()).await)
                .await?;
            output.print(&thing, show)
        }
        ThingsCommand::Delete { id } => {
            session
                .fetch(async |api| api.delete_thing(&id).await)
                .await?;
            output.done(&format!("Deleted thing {id}"))
        }
    }
}

/// The parameters that recreate `thing` as it is.
fn params_of(thing: Thing) -> anyhow::Result<NewThingParams> {
    let properties = thing
        .properties
        .into_iter()
        .flatten()
        .map(|property| match property {
            Property::PropertyString(p) => {
                PropertyParam::PropertyStringParam(PropertyStringParam {
                    name: p.name,
                    r#type: Default::default(),
                    value: p.value,
                })
            }
            Property::PropertyFloat(p) => PropertyParam::PropertyFloatParam(PropertyFloatParam {
                name: p.name,
                r#type: Default::default(),
                unit: (!p.unit.is_empty()).then_some(p.unit),
                value: p.value,
            }),
            Property::PropertyDatetime(p) => {
                PropertyParam::PropertyDatetimeParam(PropertyDatetimeParam {
                    name: p.name,
                    r#type: Default::default(),
                    value: p.value,
                })
            }
            Property::PropertyBoolean(p) => {
                PropertyParam::PropertyBooleanParam(PropertyBooleanParam {
                    name: p.name,
                    r#type: Default::default(),
                    value: p.value,
                })
            }
        })
        .collect();
    Ok(NewThingParams {
        quantity: u64::try_from(thing.quantity)
            .with_context(|| format!("the thing has a negative quantity {}", thing.quantity))?,
        name: thing.name,
        description: thing.description,
        private_note: thing.private_note.unwrap_or_default(),
        quantity_unit: thing.quantity_unit,
        sharing_state: thing.sharing_state.unwrap_or_default(),
        images_ids: Some(
            thing
                .images
                .into_iter()
                .flatten()
                .map(|image| image.id)
                .collect(),
        ),
        properties: Some(properties),
    })
}

/// Adds `properties` to `params`, each replacing one with the same name.
fn set_properties(params: &mut NewThingParams, properties: Vec<PropertyParam>) {
    let existing = params.properties.get_or_insert_default();
    for property in properties {
        match existing
            .iter()
            .position(|other| property_name(other) == property_name(&property))
        {
            Some(i) => existing[i] = property,
            None => existing.push(property),
        }
    }
}

fn property_name(property: &PropertyParam) -> &str {
    match property {
        PropertyParam::PropertyStringParam(p) => &p.name,
        PropertyParam::PropertyFloatParam(p) => &p.name,
        PropertyParam::PropertyDatetimeParam(p) => &p.name,
        PropertyParam::PropertyBooleanParam(p) => &p.name,
    }
}

/// Parses `NAME[:TYPE]=VALUE`, see the help of `--property`.
fn parse_property(arg: &str) -> Result<PropertyParam, String> {
    let (key, value) = arg.split_once('=').ok_or("expected NAME[:TYPE]=VALUE")?;
    let (name, kind) = match key.rsplit_once(':') {
        Some((name, kind @ ("string" | "float" | "datetime" | "boolean"))) => (name, kind),
        _ => (key, "string"),
    };
    if name.is_empty() {
        return Err("the property name is empty".to_owned());
    }
    let name = name.to_owned();
    Ok(match kind {
        "float" => {
            let value = value.trim();
            let (number, unit) = value
                .split_once(char::is_whitespace)
                .map_or((value, ""), |(number, unit)| (number, unit.trim()));
            PropertyParam::PropertyFloatParam(PropertyFloatParam {
                name,
                r#type: Default::default(),
                unit: (!unit.is_empty()).then(|| unit.to_owned()),
                value: number
                    .parse()
                    .map_err(|_| format!("{number:?} is not a number"))?,
            })
        }
        "datetime" => PropertyParam::PropertyDatetimeParam(PropertyDatetimeParam {
            name,
            r#type: Default::default(),
            value: parse_datetime(value.trim())?,
        }),
        "boolean" => PropertyParam::PropertyBooleanParam(PropertyBooleanParam {
            name,
            r#type: Default::default(),
            value: value
                .trim()
                .parse()
                .map_err(|_| format!("{value:?} is neither true nor false"))?,
        }),
        _ => PropertyParam::PropertyStringParam(PropertyStringParam {
            name,
            r#type: Default::default(),
            value: value.to_owned(),
        }),
    })
}

/// Parses an RFC 3339 timestamp, or a date as midnight UTC.
fn parse_datetime(value: &str) -> Result<DateTime<Utc>, String> {
    if let Ok(datetime) = DateTime::parse_from_rfc3339(value) {
        return Ok(datetime.to_utc());
    }
    NaiveDate::parse_from_str(value, "%Y-%m-%d")
        .map(|date| date.and_time(Default::default()).and_utc())
        .map_err(|_| format!("{value:?} is neither an RFC 3339 timestamp nor a YYYY-MM-DD date"))
}

/// A table with one row per thing.
pub fn table(things: &[Thing]) -> Table {
    let mut table = Table::new(&["ID", "NAME", "QUANTITY", "OWNER"]);
    for thing in things {
        table.row(vec![
            thing.id.clone(),
            thing.name.clone(),
            quantity(thing),
            thing.owner.name.clone(),
        ]);
    }
    table
}

fn show(thing: &Thing) -> String {
    let names = |names: Vec<&str>| names.join(", ");
    let mut out = text::fields(&[
        ("id", thing.id.clone()),
        ("name", thing.name.clone()),
        ("owner", text::user(&thing.owner)),
        ("quantity", quantity(thing)),
        ("created", text::date(&thing.created_at)),
        (
            "sharing",
            thing
                .sharing_state
                .as_ref()
                .map(|state| state.as_str().to_owned())
                .unwrap_or_default(),
        ),
        (
            "lists",
            names(
                thing
                    .lists
                    .iter()
                    .flatten()
                    .map(|list| list.name.as_str())
                    .collect(),
            ),
        ),
        (
            "shared with",
            names(
                thing
                    .shares
                    .iter()
                    .flatten()
                    .map(|share| share.target_user.name.as_str())
                    .collect(),
            ),
        ),
        (
            "images",
            thing
                .images
                .iter()
                .flatten()
                .map(images::describe)
                .collect::<Vec<_>>()
                .join("\n"),
        ),
        (
            "public links",
            names(
                thing
                    .public_shares
                    .iter()
                    .flatten()
                    .map(|share| share.id.as_str())
                    .collect(),
            ),
        ),
        ("description", thing.description.clone()),
        (
            "private note",
            thing.private_note.clone().unwrap_or_default(),
        ),
    ]);
    out.push_str(&properties(thing.properties.as_deref().unwrap_or_default()));
    out
}

/// An indented `properties:` block, or nothing without properties.
pub fn properties(properties: &[Property]) -> String {
    let properties: Vec<(&str, String)> = properties
        .iter()
        .map(|property| match property {
            Property::PropertyString(p) => (p.name.as_str(), p.value.clone()),
            Property::PropertyFloat(p) => (
                p.name.as_str(),
                format!("{} {}", p.value, p.unit).trim_end().to_owned(),
            ),
            Property::PropertyDatetime(p) => (p.name.as_str(), text::date(&p.value)),
            Property::PropertyBoolean(p) => (p.name.as_str(), p.value.to_string()),
        })
        .collect();
    let mut out = String::new();
    if !properties.is_empty() {
        out.push_str("\nproperties:\n");
        for line in text::fields(&properties).lines() {
            out.push_str("  ");
            out.push_str(line);
            out.push('\n');
        }
    }
    out
}

fn quantity(thing: &Thing) -> String {
    format!("{} {}", thing.quantity, thing.quantity_unit)
        .trim_end()
        .to_owned()
}
