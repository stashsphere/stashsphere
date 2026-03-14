This folder contains two kinds of schema that define
StashSphere's property system.

# Root-Schemas

They define how the actual schemas look like.
Basically they define the json schema variants
for their database counterparts, `string`, `float64`,
`boolean` and `datetime`.

These root schemas allow the backend to pre-validate
any schema so that clients get data-schema in a known format
as the data-schemas are not only used to validate data
but also to render forms and icons.

# Data-Schemas

These schemas describe individual `name:value` combinations:

Each schema defines which `type` is used for a `name` together
which values are permitted.
Furthermore,  in the case of `float64` / `number` schemas the units
can be constraint as well.

This makes sure that the entropy of the properties is kept at bay
while making it extendable by adding more schemas upstream
or locally, by supplying them to StashSphere at startup making
it available to clients.

Custom configuration is still TODO.