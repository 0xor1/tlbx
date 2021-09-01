{%- func qryGetTokens() -%}
{%- collapsespace -%}
SELECT DISTINCT f.token
FROM fcmTokens f 
JOIN users u ON f.user=u.id 
WHERE topic=? 
AND u.fcmEnabled=1
{%- endcollapsespace -%}
{%- endfunc -%}