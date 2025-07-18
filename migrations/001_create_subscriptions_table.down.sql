DROP INDEX IF EXISTS idx_subscriptions_end_date;
DROP INDEX IF EXISTS idx_subscriptions_start_date;
DROP INDEX IF EXISTS idx_subscriptions_service_name;
DROP INDEX IF EXISTS idx_subscriptions_user_id;

DROP TABLE IF EXISTS subscriptions;

DROP EXTENSION IF EXISTS "uuid-ossp";
