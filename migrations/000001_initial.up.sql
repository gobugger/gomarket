CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username TEXT NOT NULL CHECK(length(username) > 3),
	password_hash TEXT NOT NULL,
	pgp_key TEXT NOT NULL DEFAULT '',
	prev_login TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	locale VARCHAR(5) NOT NULL DEFAULT 'en-US',
	currency VARCHAR(3) NOT NULL DEFAULT 'USD',
	twofa_enabled BOOLEAN NOT NULL DEFAULT FALSE CHECK( NOT(pgp_key = '' AND twofa_enabled = TRUE)),
	incognito_enabled BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(username)
);

CREATE TABLE wallets (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	balance_pico NUMERIC NOT NULL DEFAULT 0 CHECK (balance_pico >= 0),
	user_id UUID REFERENCES users(id) NOT NULL,
	UNIQUE(user_id)
);

CREATE TYPE invoice_status AS ENUM ('pending', 'confirmed', 'expired');

CREATE TABLE invoices (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	address VARCHAR(95) NOT NULL DEFAULT '' CHECK((address = '' AND status != 'confirmed'::invoice_status) OR LENGTH(address) = 95),
	amount_pico NUMERIC NOT NULL CHECK (amount_pico > 0),
	status invoice_status NOT NULL DEFAULT 'pending'::invoice_status,
	amount_unlocked_pico NUMERIC NOT NULL DEFAULT 0 CHECK (amount_unlocked_pico >= 0),
	permanent BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE categories (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	parent_id UUID REFERENCES categories(id),
	name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ,
	UNIQUE(name)
);

CREATE TABLE products (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	category_id UUID REFERENCES categories(id) NOT NULL,
	inventory INT NOT NULL CHECK(inventory >= 0),
	ships_from TEXT NOT NULL,
	ships_to TEXT NOT NULL,
	vendor_id UUID REFERENCES users(id) NOT NULL,
	deleted_at TIMESTAMPTZ
);

CREATE TABLE price_tiers (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	product_id UUID REFERENCES products(id) NOT NULL,
	quantity INT NOT NULL CHECK(quantity > 0),
	price_cent BIGINT NOT NULL CHECK(price_cent > 0),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE delivery_methods (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	vendor_id UUID REFERENCES users(id) NOT NULL,
	description TEXT NOT NULL,
	price_cent BIGINT NOT NULL CHECK(price_cent >= 0),
	deleted_at TIMESTAMPTZ,
	UNIQUE(id, vendor_id)
);

CREATE TABLE cart_items (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	customer_id UUID REFERENCES users(id) NOT NULL,
	price_id UUID REFERENCES price_tiers(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE terms_of_services (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	content VARCHAR(4096) DEFAULT '' NOT NULL,
	vendor_id UUID REFERENCES users(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE order_status AS ENUM (
	'pending', 
	'paid',
	'cancelled',
	'accepted',
	'declined',
	'dispatched',
	'finalized',
	'disputed',
	'settled'
);

CREATE TABLE orders (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	status order_status NOT NULL DEFAULT 'pending'::order_status,
	details TEXT NOT NULL,
	total_price_pico NUMERIC NOT NULL CHECK (total_price_pico > 0), -- Total price of order in piconero
	delivery_method_id UUID REFERENCES delivery_methods(id) NOT NULL,
	vendor_id UUID REFERENCES users(id) NOT NULL,
	terms_of_service_id UUID REFERENCES terms_of_services(id) NOT NULL,
	customer_id UUID REFERENCES users(id) NOT NULL,
	num_extends INT NOT NULL DEFAULT 0 CHECK (num_extends >= 0), -- How many times AF timer has been extended
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Timestamps of order events
	paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	dispatched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	finalized_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	disputed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	FOREIGN KEY (delivery_method_id, vendor_id) REFERENCES delivery_methods(id, vendor_id)
);

CREATE TABLE order_items (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	order_id UUID REFERENCES orders(id) NOT NULL,
	price_id UUID REFERENCES price_tiers(id) NOT NULL,
	count INT NOT NULL DEFAULT 1 CHECK(count >= 1),
	UNIQUE(order_id, price_id)
);

CREATE TABLE order_invoices (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	order_id UUID REFERENCES orders(id) NOT NULL,
	invoice_id UUID REFERENCES invoices(id) NOT NULL
);

CREATE TABLE reviews (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	grade INT NOT NULL CHECK(grade >= 0 AND grade <= 5),
	comment TEXT NOT NULL,
	order_id UUID REFERENCES orders(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(order_id)
);

CREATE TABLE product_reviews (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	grade INT NOT NULL CHECK(grade >= 0 AND grade <= 5),
	comment TEXT NOT NULL,
	order_item_id UUID REFERENCES order_items(id) NOT NULL,
	UNIQUE(order_item_id)
);

CREATE TABLE order_chat_messages (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	message TEXT NOT NULL,
	author_id UUID REFERENCES users(id) NOT NULL,
	order_id UUID REFERENCES orders(id) ON DELETE CASCADE NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE dispute_offer_status AS ENUM (
	'pending',
	'declined',
	'accepted',
	'forced'
);

CREATE TABLE dispute_offers (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	refund_factor DOUBLE PRECISION NOT NULL CHECK(refund_factor >= 0 AND refund_factor <= 1),
	status dispute_offer_status NOT NULL DEFAULT 'pending',
	order_id UUID REFERENCES orders(id) ON DELETE CASCADE NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE deposits (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	wallet_id UUID REFERENCES wallets(id) NOT NULL,
	invoice_id UUID REFERENCES invoices(id) ON DELETE CASCADE NOT NULL,
	amount_deposited_pico NUMERIC NOT NULL DEFAULT 0 CHECK(amount_deposited_pico >= 0),
	UNIQUE(wallet_id, invoice_id)
);

CREATE TYPE withdrawal_status AS ENUM ('pending', 'processing');

CREATE TABLE withdrawals (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	amount_pico NUMERIC NOT NULL CHECK (amount_pico > 0),
	destination_address TEXT NOT NULL CHECK (LENGTH(destination_address) = 95),
	status withdrawal_status NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	hash TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vendor_applications (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	existing_vendor BOOLEAN NOT NULL DEFAULT FALSE,
	letter VARCHAR(4096) NOT NULL,
	price_paid_pico NUMERIC NOT NULL CHECK(price_paid_pico >= 0),
	user_id UUID REFERENCES users(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(user_id)
);

CREATE TABLE vendor_licenses (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	price_paid_pico NUMERIC NOT NULL CHECK(price_paid_pico >= 0),
	user_id UUID REFERENCES users(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(user_id)
);

CREATE TABLE tickets (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	subject TEXT NOT NULL,
	message TEXT NOT NULL,
	author_id UUID REFERENCES users(id) NOT NULL,	
	is_open BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ticket_responses (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	message TEXT NOT NULL,
	ticket_id UUID REFERENCES tickets(id) NOT NULL,
	author_name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bans (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID REFERENCES users(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(user_id)
);

CREATE TABLE exchange_rates (
	id SERIAL PRIMARY KEY CHECK(id = 1), -- Enforce one row only
	data JSONB NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	title TEXT NOT NULL,
	content TEXT NOT NULL,
	is_seen BOOLEAN NOT NULL DEFAULT FALSE,
	user_id UUID REFERENCES users(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE settings (
	id SERIAL PRIMARY KEY CHECK(id = 1), -- Enforce one row only
	data JSONB NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default settings if not already present
INSERT INTO settings (id, data)
VALUES (
    1,
    jsonb_build_object(
        'vendor_application_price', 1e12 -- 1XMR
    )
)
ON CONFLICT (id) DO NOTHING;

-- Check vendor for order_items
CREATE OR REPLACE FUNCTION check_order_item()
RETURNS TRIGGER AS $$
DECLARE
	order_vendor_id UUID;
	product_vendor_id UUID;
BEGIN
	SELECT orders.vendor_id INTO order_vendor_id
	FROM orders
	WHERE orders.id = NEW.order_id;

	SELECT products.vendor_id INTO product_vendor_id
	FROM price_tiers
	JOIN products ON products.id = price_tiers.product_id
	WHERE price_tiers.id = NEW.price_id;

	IF order_vendor_id IS NULL OR product_vendor_id IS NULL THEN
		RAISE EXCEPTION 'Order or product vendor not found';
	END IF;

	IF order_vendor_id != product_vendor_id THEN
		RAISE EXCEPTION 'Order vendor % does not match product vendor %',
		order_vendor_id, product_vendor_id;
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_order_vendor
	BEFORE INSERT OR UPDATE ON order_items
	FOR EACH ROW
	EXECUTE FUNCTION check_order_item();

-- Update timestampz for order
CREATE OR REPLACE FUNCTION set_order_timestamptz()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'paid' AND OLD.status IS DISTINCT FROM 'paid' THEN
        NEW.paid_at := NOW();

    ELSIF NEW.status = 'accepted' AND OLD.status IS DISTINCT FROM 'accepted' THEN
        NEW.accepted_at := NOW();

    ELSIF NEW.status = 'dispatched' AND OLD.status IS DISTINCT FROM 'dispatched' THEN
        NEW.dispatched_at := NOW();

    ELSIF NEW.status = 'finalized' AND OLD.status IS DISTINCT FROM 'finalized' THEN
        NEW.finalized_at := NOW();

    ELSIF NEW.status = 'disputed' AND OLD.status IS DISTINCT FROM 'disputed' THEN
        NEW.disputed_at := NOW();

    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_order_timestamptz
	BEFORE UPDATE ON orders
	FOR EACH ROW
	EXECUTE FUNCTION set_order_timestamptz();

