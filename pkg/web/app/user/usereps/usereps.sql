{% import . "github.com/0xor1/tlbx/pkg/core" %}
{% import "github.com/0xor1/tlbx/pkg/json" %}
{% import sqlh "github.com/0xor1/tlbx/pkg/web/app/sql" %}

{%- func qryJinInsert(args *sqlh.Args, me ID, val interface{}) -%}
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
{%- code 
    *args = *sqlh.NewArgs(2) 
    args.Append(
    me,
    json.MustMarshal(val),
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryJinSelect(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
SELECT val
FROM jin
WHERE user=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.AppendOne(
    me,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryJinDelete(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
Delete FROM jin
WHERE user=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.AppendOne(
    me,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}