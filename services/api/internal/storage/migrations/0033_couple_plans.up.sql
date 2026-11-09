CREATE TABLE couple_plans (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_id bigint NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    partner_id bigint NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    pair_version bigint NOT NULL,
    draft_id text NOT NULL,
    proposed_by bigint NOT NULL,
    revision bigint NOT NULL DEFAULT 1,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'declined', 'cancelled')),
    starts_at timestamptz NOT NULL,
    card jsonb NOT NULL,
    CHECK (owner_id <> partner_id AND proposed_by IN (owner_id, partner_id)),
    UNIQUE (owner_id, draft_id)
);

CREATE INDEX couple_plans_owner_start_idx ON couple_plans (owner_id, starts_at);
CREATE INDEX couple_plans_partner_start_idx ON couple_plans (partner_id, starts_at);

CREATE TABLE couple_plan_notices (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    plan_id bigint NOT NULL REFERENCES couple_plans(id) ON DELETE CASCADE,
    revision bigint NOT NULL,
    recipient_id bigint NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    reminder boolean NOT NULL DEFAULT false,
    due_at timestamptz NOT NULL,
    processed_at timestamptz,
    UNIQUE (plan_id, revision, recipient_id, reminder)
);

CREATE INDEX couple_plan_notices_due_idx ON couple_plan_notices (due_at, id) WHERE processed_at IS NULL;
