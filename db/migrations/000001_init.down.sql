ALTER TABLE orders DROP CONSTRAINT orders_user_id_fkey;
ALTER TABLE withdrawls DROP CONSTRAINT withdrawls_user_id_fkey;
ALTER TABLE balance DROP CONSTRAINT balance_user_id_fkey;

DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS withdrawals;
DROP TABLE IF EXISTS balances;
DROP TABLE IF EXISTS withdraw_balances;