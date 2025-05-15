CREATE USER cryptellation_backtests;
ALTER USER cryptellation_backtests PASSWORD 'cryptellation_backtests';
ALTER USER cryptellation_backtests CREATEDB;

CREATE DATABASE cryptellation_backtests;
GRANT ALL PRIVILEGES ON DATABASE cryptellation_backtests TO cryptellation_backtests;
\c cryptellation_backtests postgres
GRANT ALL ON SCHEMA public TO cryptellation_backtests;