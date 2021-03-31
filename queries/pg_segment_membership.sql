-- Query over the segments and active users of a given vendor.
SELECT vendor,
    segment_name,
    user_id,
    first::TIMESTAMP,
    last::TIMESTAMP
FROM segment_memberships
WHERE vendor = 6 AND segment_name = 'seg-113'
    AND member
ORDER BY segment_name;