/**Enable UUID generation**/
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

/**Create credit_tags table**/
DO language 'plpgsql'
$$
DECLARE var_ql text :=  'CREATE TABLE credit_tags('
                        || 'id uuid DEFAULT uuid_generate_v4 (),'
                        || 'first_name varchar(36) NOT NULL, '
                        || 'last_name varchar(36) NOT NULL, '
                        || 'full_name varchar(72) NOT NULL, '
                        || 'social_security_number integer, '
                        || string_agg('X'|| to_char(i, 'fm0000') || ' integer', ',')
                        || ');'
FROM generate_series(1,200) As i;
BEGIN
  raise notice 'Value: %', var_ql;
  EXECUTE var_ql;
END;
$$ ;

/**Get median of column**/
CREATE FUNCTION _final_median(anyarray) RETURNS float8 AS $$
WITH q AS
    (
    SELECT val
    FROM unnest($1) val
    WHERE VAL IS NOT NULL
    ORDER BY 1
    ),
     cnt AS
    (
    SELECT COUNT(*) AS c FROM q
    )
SELECT AVG(val)::float8
FROM
     (
     SELECT val FROM q
     LIMIT  2 - MOD((SELECT c FROM cnt), 2)
     OFFSET GREATEST(CEIL((SELECT c FROM cnt) / 2.0) - 1,0)
     ) q2;
$$ LANGUAGE SQL IMMUTABLE;

CREATE AGGREGATE median(anyelement) (
  SFUNC=array_append,
  STYPE=anyarray,
  FINALFUNC=_final_median,
  INITCOND='{}'
);
