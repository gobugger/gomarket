-- name: CreateVendorApplication :one
INSERT INTO vendor_applications (
	existing_vendor, letter, price_paid_pico, user_id
) VALUES (
	$1, $2, $3, $4
) RETURNING *;

-- name: GetVendorApplication :one
SELECT * FROM vendor_applications
WHERE id = $1;

-- name: GetVendorApplications :many
SELECT * FROM vendor_applications;

-- name: DeleteVendorApplication :exec
DELETE FROM vendor_applications WHERE id = $1;
