-- name: ClearDelegate :execrows
UPDATE nodes
SET "delegated_to" = ''
WHERE "delegate_tx_hash" = NULL;

-- name: SetDelegates :execrows
UPDATE nodes
SET delegated_to = @delegatee
WHERE sale_block = $1 AND
    sale_tx_index = $2 AND
    node_id = ANY (@node_ids::int[]);

-- name: GetNodes :many
SELECT *
FROM nodes
WHERE sale_block = $1 AND
    sale_tx_index = $2 AND
    node_id = ANY (@node_ids::int[]);


-- name: GetNodesByOwner :many
SELECT * 
FROM nodes
WHERE sale_block = $1 AND
    sale_tx_index = $2 AND
    owner_public_key = $3
ORDER BY tier_index;

-- name: GetNodesByPubkey :many
SELECT nodes.*
FROM nodes JOIN events ON nodes.purchase_tx_hash = events.tx_hash
WHERE sale_block = $1 AND
    sale_tx_index = $2 AND
    owner_public_key = $3 AND
    delegated_to = $4;

-- name: AddNode :exec
INSERT INTO nodes(sale_block, sale_tx_index, node_id, tier_index, delegated_to, owner_public_key, purchase_tx_hash, delegate_tx_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetNodeCountByTierIndex :many
SELECT (tiers.tier_index)::int as tier_index, count(nodes.tier_index)
FROM generate_series(@from_tier::int,@to_tier::int) as tiers(tier_index)
LEFT JOIN 
	(select * 
	from nodes 
	where sale_block = $1 and 
		sale_tx_index= $2) 
	as nodes on tiers.tier_index = nodes.tier_index 
group by tiers.tier_index
ORDER BY tiers.tier_index;