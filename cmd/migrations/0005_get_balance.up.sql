CREATE OR REPLACE FUNCTION t_gophermart.get_user_stats(s_user_login VARCHAR)
RETURNS TABLE(
    current_balance DECIMAL(10, 2),
    total_withdrawn DECIMAL(10, 2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COALESCE(SUM(CASE WHEN s_type = 'plus' THEN n_value ELSE 0 END), 0)
        - COALESCE(SUM(CASE WHEN s_type = 'minus' THEN n_value ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN s_type = 'minus' THEN n_value ELSE 0 END), 0)
    FROM t_gophermart.t_transactions
    WHERE s_user = s_user_login;
END;
$$ LANGUAGE plpgsql;
