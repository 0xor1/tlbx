{% import "github.com/0xor1/tlbx/pkg/sqlh" %}

{%- func qryUserFullGet(byID bool) -%}
{%- collapsespace -%}
SELECT id,
    email,
    handle,
    alias,
    hasAvatar,
    fcmEnabled,
    registeredOn,
    activatedOn,
    newEmail,
    activateCode,
    changeEmailCode,
    lastPwdResetOn,
    loginLinkCodeCreatedOn,
    loginLinkCode
FROM users
WHERE
{%- if byID -%}
    id
{%- else -%}
    email
{%- endif -%}
=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryUserUpdate() -%}
{%- collapsespace -%}
UPDATE users
SET email=?,
    handle=?,
    alias=?,
    hasAvatar=?,
    fcmEnabled=?,
    registeredOn=?,
    activatedOn=?,
    newEmail=?,
    activateCode=?,
    changeEmailCode=?,
    lastPwdResetOn=?,
    loginLinkCodeCreatedOn=?,
    loginLinkCode=?
WHERE id=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryUserInsert() -%}
{%- collapsespace -%}
INSERT INTO users (
    id,
    email,
    handle,
    alias,
    hasAvatar,
    fcmEnabled,
    registeredOn,
    activatedOn,
    activateCode
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryUserDelete() -%}
{%- collapsespace -%}
DELETE FROM users
WHERE id=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryUsersGet(n int) -%}
{%- collapsespace -%}
SELECT id,
    handle,
    alias,
    hasAvatar
FROM users
WHERE id IN ({%s sqlh.PList(n)%})
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryPwdDelete() -%}
{%- collapsespace -%}
DELETE FROM pwds
WHERE id=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryPwdGet() -%}
{%- collapsespace -%}
SELECT id,
    salt,
    pwd,
    n,
    r,
    p
FROM pwds
WHERE id=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryPwdUpdate() -%}
{%- collapsespace -%}
INSERT INTO pwds (
    id,
    salt,
    pwd,
    n,
    r,
    p
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) ON DUPLICATE KEY UPDATE
salt=VALUE(salt),
pwd=VALUE(pwd),
n=VALUE(n),
r=VALUE(r),
p=VALUE(p)
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryFifthOldestTokenCreatedOn() -%}
{%- collapsespace -%}
SELECT createdOn
FROM fcmTokens
WHERE user=?
ORDER BY createdOn DESC
LIMIT 4, 1
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryFCMTokensDeleteOldest() -%}
{%- collapsespace -%}
DELETE FROM fcmTokens
WHERE user=?
AND createdOn<=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryFCMTokenInsert() -%}
{%- collapsespace -%}
INSERT INTO fcmTokens (
    topic,
    token,
    user,
    client,
    createdOn
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
) ON DUPLICATE KEY UPDATE 
topic=VALUES(topic),
token=VALUES(token),
user=VALUES(user),
client=VALUES(client),
createdOn=VALUES(createdOn)
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryDistinctFCMTokens() -%}
{%- collapsespace -%}
SELECT DISTINCT token
FROM fcmTokens
WHERE user=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryFCMTokensDelete() -%}
{%- collapsespace -%}
DELETE FROM fcmTokens
WHERE user=?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryFCMUnregister() -%}
{%- collapsespace -%}
DELETE FROM fcmTokens
WHERE user=?
AND client=?
{%- endcollapsespace -%}
{%- endfunc -%}

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