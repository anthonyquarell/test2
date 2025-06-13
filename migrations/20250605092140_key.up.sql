CREATE TYPE key_status AS ENUM ('new', 'activated');

CREATE TABLE key (
                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                     created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                     updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                     provider_id TEXT NOT NULL DEFAULT '',
                     provider_product_id TEXT NOT NULL DEFAULT  '',
                     product_id TEXT NOT NULL DEFAULT '',
                     value TEXT NOT NULL DEFAULT'',
                     status key_status NOT NULL DEFAULT 'new',
                     customer_phone TEXT NOT NULL DEFAULT '',
                     order_id TEXT NOT NULL DEFAULT ''
);


ALTER TABLE key ADD CONSTRAINT unique_value UNIQUE (value);
