-- name: CreateVendorLicense :one
INSERT INTO vendor_licenses (
	price_paid_pico, user_id
) VALUES (
	$1, $2
) RETURNING *;

-- name: GetVendorLicenseForUser :one
SELECT * FROM vendor_licenses
WHERE user_id = $1;

-- name: UpdateVendorInfo :one
UPDATE vendor_licenses
SET vendor_info = $2
WHERE id = $1
RETURNING *;

-- name: GetNumberOfVendorLicenses :one
SELECT COUNT(*) FROM vendor_licenses;

-- name: HasVendorLicense :one
SELECT 1
FROM vendor_licenses
WHERE user_id = $1;
