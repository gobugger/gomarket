-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_set_order_timestamptz ON orders;
DROP TRIGGER IF EXISTS trigger_check_order_vendor ON order_items;

-- Drop functions
DROP FUNCTION IF EXISTS set_order_timestamptz();
DROP FUNCTION IF EXISTS check_order_item();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS bans;
DROP TABLE IF EXISTS ticket_responses;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS vendor_licenses;
DROP TABLE IF EXISTS vendor_applications;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS withdrawals;
DROP TABLE IF EXISTS deposits;
DROP TABLE IF EXISTS dispute_offers;
DROP TABLE IF EXISTS order_chat_messages;
DROP TABLE IF EXISTS product_reviews;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS order_invoices;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS terms_of_services;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS delivery_methods;
DROP TABLE IF EXISTS price_tiers;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS sessions;

-- Drop custom types
DROP TYPE IF EXISTS withdrawal_status;
DROP TYPE IF EXISTS dispute_offer_status;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS invoice_status;

-- Drop extensions
DROP EXTENSION IF EXISTS pgcrypto;
