CREATE TABLE element_properties (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    element_id   uuid        NOT NULL REFERENCES elements(id) ON DELETE CASCADE,
    key          text        NOT NULL,
    value        text        NOT NULL,
    source       text        NOT NULL DEFAULT 'model',
    collected_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX element_properties_element_idx ON element_properties(element_id);
CREATE INDEX element_properties_key_idx     ON element_properties(key);
