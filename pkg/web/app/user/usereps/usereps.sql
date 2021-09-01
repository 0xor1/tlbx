{%- func qryJinInsert() -%}
{%- collapsespace -%}
INSERT INTO jin(
    user,
    val
)
VALUES (
    ?,
    ?
)
ON DUPLICATE KEY UPDATE 
val=VALUES(val)
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryJinSelect() -%}
{%- collapsespace -%}
SELECT val
FROM jin
WHERE user=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryJinDelete() -%}
{%- collapsespace -%}
Delete FROM jin
WHERE user=?
{%- endcollapsespace -%}
{%- endfunc -%}