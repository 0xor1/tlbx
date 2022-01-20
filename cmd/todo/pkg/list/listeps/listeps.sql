{% import . "github.com/0xor1/tlbx/pkg/core" %}
{% import "github.com/0xor1/tlbx/pkg/ptr" %}
{% import "github.com/0xor1/tlbx/pkg/sqlh" %}
{% import "github.com/0xor1/tlbx/cmd/todo/pkg/list" %}

{%- func qryListInsert() -%}
{%- collapsespace -%}
INSERT INTO lists (
    user,
    id,
    createdOn,
    name,
    todoItemCount,
    completedItemCount
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryListUpdate() -%}
{%- collapsespace -%}
UPDATE lists SET 
    name=?
WHERE 
    user=?
AND
    id=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryListsDelete(n int) -%}
{%- collapsespace -%}
DELETE FROM 
    lists 
WHERE
    user=?
AND
    id IN ({%s sqlh.PList(n)%})
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryOnDelete() -%}
{%- collapsespace -%}
DELETE FROM lists
WHERE user=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryListsGet(sqlArgs *sqlh.Args, me ID, args *list.Get) -%}
{%- collapsespace -%}
SELECT id,
    createdOn,
    name,
    todoItemCount,
    completedItemCount
FROM lists
WHERE user=?
{%- code 
    sqlArgs.Append(me)
-%}
{%- if len(args.Base.IDs) > 0 -%}
    AND id IN ({%s sqlh.PList(len(args.Base.IDs)) %})
    ORDER BY FIELD (id,{%s sqlh.PList(len(args.Base.IDs)) %})
    {%- code 
        is := args.Base.IDs.ToIs()
        sqlArgs.Append(is...)
        sqlArgs.Append(is...)
    -%}
{%- else -%}
    {%- if ptr.StringOr(args.NamePrefix, "") != "" -%}
        AND name LIKE ?
        {%- code 
            sqlArgs.Append(Strf("%s%%", *args.NamePrefix))
        -%}
    {%- endif -%}
    {%- if args.CreatedOnMin != nil -%}
        AND createdOn >= ?
        {%- code 
            sqlArgs.Append(*args.CreatedOnMin)
        -%}
    {%- endif -%}
    {%- if args.CreatedOnMax != nil -%}
        AND createdOn <= ?
        {%- code 
            sqlArgs.Append(*args.CreatedOnMax)
        -%}
    {%- endif -%}
    {%- if args.TodoItemCountMin != nil -%}
        AND todoItemCount >= ?
        {%- code 
            sqlArgs.Append(*args.TodoItemCountMin)
        -%}
    {%- endif -%}
    {%- if args.TodoItemCountMax != nil -%}
        AND todoItemCount <= ?
        {%- code 
            sqlArgs.Append(*args.TodoItemCountMax)
        -%}
    {%- endif -%}
    {%- if args.Base.After != nil -%}
        AND {%s string(args.Base.Sort) %} {%s= sqlh.GtLtSymbol(*args.Base.Asc) %}= (SELECT {%s string(args.Base.Sort) %} FROM lists WHERE user=? AND id=?) AND id <> ?
        {%- code 
            sqlArgs.Append(me, *args.Base.After, *args.Base.After)
        -%}
        {%- if args.Base.Sort != list.SortCreatedOn -%}
            AND createdOn {%s= sqlh.GtLtSymbol(*args.Base.Asc)%} (SELECT createdOn FROM lists WHERE user=? AND id=?)
            {%- code 
                sqlArgs.Append(me, *args.Base.After)
            -%}
        {%- endif -%}
    {%- endif -%}
    ORDER BY {%s string(args.Base.Sort) %}
    {%- if args.Base.Sort != list.SortCreatedOn -%}
        , createdOn
    {%- endif -%}
    {%s sqlh.Asc(*args.Base.Asc) %} LIMIT {%d int(sqlh.Limit100(args.Base.Limit)) %}
{%- endif -%}
{%- endcollapsespace -%}
{%- endfunc -%}