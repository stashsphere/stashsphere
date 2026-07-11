CREATE TABLE public_shares (
    id          TEXT       NOT NULL PRIMARY KEY,
    created_at  TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    owner_id    TEXT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    thing_id    TEXT       REFERENCES things(id) ON DELETE CASCADE,
    list_id     TEXT       REFERENCES lists(id) ON DELETE CASCADE,
    CHECK ((thing_id IS NULL) <> (list_id IS NULL))
);

CREATE INDEX public_shares_thing_id_idx ON public_shares(thing_id);
CREATE INDEX public_shares_list_id_idx ON public_shares(list_id);
CREATE INDEX public_shares_owner_id_idx ON public_shares(owner_id);
