-- Helper function for automated partition creation
CREATE OR REPLACE FUNCTION create_check_results_partition(p_year INT, p_month INT)
RETURNS VOID AS $$
DECLARE
    start_date DATE;
    end_date DATE;
    table_name TEXT;
BEGIN
    start_date := MAKE_DATE(p_year, p_month, 1);
    end_date := start_date + INTERVAL '1 month';
    table_name := FORMAT('check_results_%s_%s', p_year, LPAD(p_month::TEXT, 2, '0'));

    EXECUTE FORMAT(
        'CREATE TABLE IF NOT EXISTS %I (CHECK (timestamp >= %L AND timestamp < %L)) INHERITS (check_results);',
        table_name, start_date, end_date
    );
    EXECUTE FORMAT('CREATE INDEX IF NOT EXISTS %I ON %I (target_id, timestamp DESC);', 'idx_' || table_name || '_tt', table_name);
END;
$$ LANGUAGE plpgsql;
