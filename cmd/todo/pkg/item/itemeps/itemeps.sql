{% import "time" %}
{% import . "github.com/0xor1/tlbx/pkg/core" %}
{% import "github.com/0xor1/tlbx/pkg/ptr" %}
{% import "github.com/0xor1/tlbx/pkg/sqlh" %}
{% import "github.com/0xor1/tlbx/cmd/todo/pkg/item" %}

{%- func qryItemInsert() -%}
{%- collapsespace -%}
INSERT INTO items (
    user,
    list,
    id,
    createdOn,
    name,
    completedOn
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

{%- func qryIncrementListItemCount() -%}
{%- collapsespace -%}
UPDATE lists
SET todoItemCount = todoItemCount + 1
WHERE user=?
AND id=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryItemUpdate() -%}
{%- collapsespace -%}
UPDATE items
SET name=?,
    completedOn=?
WHERE user=?
AND list=?
AND id=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryListCountsToggle(completed bool) -%}
{%- collapsespace -%}
UPDATE lists
SET todoItemCount = todoItemCount{%- if completed -%}-{%- else -%}+{%- endif -%}1,
    completedItemCount = completedItemCount{%- if completed -%}+{%- else -%}-{%- endif -%}1
WHERE user=?
AND id=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryItemsDelete(n int) -%}
{%- collapsespace -%}
DELETE FROM items
WHERE user=?
AND list=?
AND id IN ({%s sqlh.PList(n)%})
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryListRecalculateCounts() -%}
{%- collapsespace -%}
UPDATE lists
SET todoItemCount = (
    SELECT COUNT(id) 
    FROM items 
    WHERE user=? 
    AND list=? 
    AND completedOn=?
), 
completedItemCount = (
    SELECT COUNT(id)
    FROM items
    WHERE user=?
    AND list=?
    AND completedOn<>?
)
WHERE user=?
AND id=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryItemsGet(sqlArgs *sqlh.Args, me ID, args *item.Get) -%}
{%- collapsespace -%}
SELECT id,
    createdOn,
    name,
    completedOn
FROM items
WHERE user=?
AND list=?
{%- code 
    if len(args.Base.IDs) > 0 {
        *sqlArgs = *sqlh.NewArgs((len(args.Base.IDs)*2) + 3)
    } else {
        *sqlArgs = *sqlh.NewArgs(20)
    }
    sqlArgs.Append(me)
    sqlArgs.Append(args.List)
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
    {%- code 
        sqlArgs.Append(time.Time{})
    -%}
    AND completedOn
    {%- if ptr.BoolOr(args.Completed, false) -%}
        <>?
        {%- if args.CompletedOnMin != nil -%}
            AND completedOn >= ?
            {%- code 
                sqlArgs.Append(*args.CompletedOnMin)
            -%}
        {%- endif -%}
        {%- if args.CompletedOnMax != nil -%}
            AND completedOn <= ?
            {%- code 
                sqlArgs.Append(*args.CompletedOnMax)
            -%}
        {%- endif -%}
    {%- else -%}
        =?
    {%- endif -%}
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
    {%- if args.Base.After != nil -%}
        AND {%s string(args.Base.Sort) %} {%s= sqlh.GtLtSymbol(*args.Base.Asc) %}= (SELECT {%s string(args.Base.Sort) %} FROM items WHERE user=? AND list=? AND id=?) AND id <> ?
        {%- code 
            sqlArgs.Append(me, args.List, *args.Base.After, *args.Base.After)
        -%}
        {%- if args.Base.Sort != item.SortCreatedOn -%}
            AND createdOn {%s= sqlh.GtLtSymbol(*args.Base.Asc)%} (SELECT createdOn FROM items WHERE user=? AND list=? AND id=?)
            {%- code 
                sqlArgs.Append(me, args.List, *args.Base.After)
            -%}
        {%- endif -%}
    {%- endif -%}
    ORDER BY {%s string(args.Base.Sort) %}
    {%- if args.Base.Sort != item.SortCreatedOn -%}
        , createdOn
    {%- endif -%}
    {%s sqlh.Asc(*args.Base.Asc) %} LIMIT {%d int(sqlh.Limit100(args.Base.Limit)) %}
{%- endif -%}
{%- endcollapsespace -%}
{%- endfunc -%}