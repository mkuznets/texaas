DROP TABLE IF EXISTS texaas_builds;
CREATE TABLE IF NOT EXISTS texaas_builds
(
    id          bigserial PRIMARY KEY,
    created_at  timestamptz                               NOT NULL DEFAULT CURRENT_TIMESTAMP,
    task_id     bigint REFERENCES ocher_tasks (id) UNIQUE NOT NULL,
    base_path   text                                      NOT NULL,
    main_source text                                      NOT NULL,
    compiler    text                                      NOT NULL,
    latex       text                                      NOT NULL
);


DROP TABLE IF EXISTS texaas_cache;
CREATE TABLE IF NOT EXISTS texaas_cache
(
    id       bigserial PRIMARY KEY,
    hash     text        NOT NULL UNIQUE,
    used_at  timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_ready bool        NOT NULL DEFAULT 'f'
);


DROP TABLE IF EXISTS texaas_inputs;
CREATE TABLE IF NOT EXISTS texaas_inputs
(
    id       bigserial PRIMARY KEY,
    build_id bigint REFERENCES texaas_builds (id) NOT NULL,
    cache_id bigint REFERENCES texaas_cache (id)  NOT NULL,
    path     text                                 NOT NULL
);

CREATE INDEX texaas_inputs__build_id_idx ON texaas_inputs USING HASH (build_id);

---- create above / drop below ----

DROP TABLE IF EXISTS texaas_inputs;
DROP TABLE IF EXISTS texaas_cache;
DROP TABLE IF EXISTS texaas_builds;
