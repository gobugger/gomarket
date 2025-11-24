-- name: CreateVendorLicense :one
INSERT INTO vendor_licenses (
	price_paid_pico, user_id
) VALUES (
	$1, $2
) RETURNING *;

-- name: GetVendorLicenseForUser :one
SELECT * FROM vendor_licenses
WHERE user_id = $1;

-- name: GetNumberOfVendorLicenses :one
SELECT COUNT(*) FROM vendor_licenses;

-- name: HasVendorLicense :one
SELECT 1
FROM vendor_licenses
WHERE user_id = $1;

-- name: CreateTermsOfService :one
INSERT INTO terms_of_services (
	content, vendor_id
) VALUES (
	$1, $2
) RETURNING *;

-- name: GetTermsOfServiceForVendor :one
SELECT * 
FROM terms_of_services 
WHERE vendor_id = $1
ORDER BY created_at ASC
LIMIT 1;
