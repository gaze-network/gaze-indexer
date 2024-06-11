-- name: ClearEvents :exec
DELETE FROM events
WHERE tx_hash <> '';
;