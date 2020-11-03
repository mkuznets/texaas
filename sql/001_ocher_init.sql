DROP TYPE IF EXISTS ocher_task_status CASCADE;
CREATE TYPE ocher_task_status AS ENUM (
    'NEW',
    'ENQUEUED',
    'EXECUTING',
    'FAILURE',
    'EXPIRED',
    'FINISHED'
    );

DROP TABLE IF EXISTS ocher_tasks CASCADE;
CREATE TABLE IF NOT EXISTS ocher_tasks
(
    id     bigserial PRIMARY KEY,
    status ocher_task_status DEFAULT 'NEW',
    name   text NOT NULL,
    args   bytea             DEFAULT NULL,
    result bytea             DEFAULT NULL
);

CREATE INDEX "ocher_tasks__status_idx" ON ocher_tasks USING HASH (status) WHERE status = 'ENQUEUED';
CREATE INDEX "ocher_tasks__name_idx" ON ocher_tasks USING HASH (name);

-- -----------------------------------------------------------------------------

DROP TYPE IF EXISTS ocher_report_tag CASCADE;
CREATE TYPE ocher_report_tag AS ENUM (
    'REPORT',
    'ERROR'
    );

DROP TABLE IF EXISTS ocher_reports CASCADE;
CREATE TABLE IF NOT EXISTS ocher_reports
(
    id         bigserial PRIMARY KEY,
    task_id    bigint           NOT NULL,
    tag        ocher_report_tag NOT NULL,
    message    text DEFAULT NULL,
    created_at timestamptz      NOT NULL
);

CREATE INDEX "ocher_reports__task_id_idx" ON ocher_reports (task_id);
CREATE INDEX "ocher_reports__task_id_tag_idx" ON ocher_reports (task_id, tag);

DROP TABLE IF EXISTS ocher_statuses CASCADE;
CREATE TABLE IF NOT EXISTS ocher_statuses
(
    id         bigserial PRIMARY KEY,
    task_id    bigint            NOT NULL,
    status     ocher_task_status NOT NULL,
    changed_at timestamptz       NOT NULL
);

CREATE INDEX "ocher_statuses__task_id_idx" ON ocher_statuses (task_id);
CREATE INDEX "ocher_statuses__task_id_status_idx" ON ocher_statuses (task_id, status) WHERE status = 'ENQUEUED';


-- -----------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION ocher_status_change()
    RETURNS trigger AS
$$
BEGIN
    IF NEW.status = 'ENQUEUED' THEN
        PERFORM pg_notify(concat('ocher_', regexp_replace(NEW.name, '[^a-zA-Z0-9_]+', '_', 'g')), NEW.id::text);
    END IF;

    IF (TG_OP = 'INSERT' OR NEW.status <> OLD.status) THEN
        INSERT INTO ocher_statuses (task_id, status, changed_at)
        VALUES (NEW.id, NEW.status, STATEMENT_TIMESTAMP());
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS ocher_status_watcher ON ocher_tasks;
CREATE TRIGGER ocher_status_watcher
    AFTER INSERT OR UPDATE
    ON ocher_tasks
    FOR EACH ROW
    WHEN (NEW.status <> 'EXECUTING')
EXECUTE PROCEDURE ocher_status_change();


---- create above / drop below ----

DROP TRIGGER IF EXISTS ocher_notifier ON ocher_tasks;
DROP FUNCTION IF EXISTS ocher_tasks_notify CASCADE;
DROP TRIGGER IF EXISTS ocher_status_watcher ON ocher_tasks;
DROP FUNCTION IF EXISTS ocher_status_change CASCADE;
DROP TABLE IF EXISTS ocher_tasks CASCADE;
DROP TABLE IF EXISTS ocher_reports CASCADE;
DROP TYPE IF EXISTS ocher_task_status CASCADE;
