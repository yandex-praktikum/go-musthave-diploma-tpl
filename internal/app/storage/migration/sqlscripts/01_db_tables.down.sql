BEGIN;
------------
-- TABLES --
------------
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS balances;

DROP TYPE IF EXISTS STATS CASCADE;

---------------
-- FUNCTIONS --
---------------

----------
-- DATA --
----------

COMMIT;