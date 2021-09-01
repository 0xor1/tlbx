{%- func qryGameInsert() -%}
{%- collapsespace -%}
INSERT INTO games (
    id,
    type,
    updatedOn,
    isActive,
    serialized
) VALUES (
    ?,
    ?,
    ?,
    1,
    ?
)
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryGameUpdate() -%}
{%- collapsespace -%}
UPDATE games
Set updatedOn=?,
    isActive=?,
    serialized=?
WHERE id=?
AND type=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryGameGet(forUpdate bool) -%}
{%- collapsespace -%}
SELECT type,
    serialized
FROM games
WHERE id=?
{%- if forUpdate -%}
FOR UPDATE
{%- endif -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryGameGetActive(forUpdate bool) -%}
{%- collapsespace -%}
SELECT g.type,
    g.serialized
FROM games g
INNER JOIN players p ON p.game=g.id
WHERE p.id=?
AND g.isActive=1
{%- if forUpdate -%}
FOR UPDATE
{%- endif -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryGameDeleteExpired() -%}
{%- collapsespace -%}
DELETE FROM games
WHERE updatedOn<?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryPlayerInsert() -%}
{%- collapsespace -%}
INSERT INTO players (
    id,
    game
) VALUES (
    ?,
    ?
)
{%- endcollapsespace -%}
{%- endfunc -%}